package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

/*		POST	https://ais.usvisa-info.com/en-ca/niv/users/sign_in		->	200
==>		GET		https://ais.usvisa-info.com/en-ca/niv/account			->	302
==>		GET		https://ais.usvisa-info.com/en-ca/niv/groups/[8nums]	->	200
		--> BODY part `to extract
				1. Schedule ID
				2. Appointment date
			************************************************
                                    <div class='card'>
                                        <p class='consular-appt'>
                                            <strong>
                                                Consular Appointment<span>&#58;</span>
                                            </strong>
                                            1 October, 2026, 10:45 Toronto local time at Toronto
 &mdash;
                                            <a href="/en-ca/niv/schedule/[scheduleID]/addresses/consulate">
                                                <span class='fas fa-map-marker-alt'></span>
                                                get directions

                                            </a>
                                        </p>`
									</div>
		************************************************

*/

type AppointmentDay struct {
	Date string `json:"date"`
}

type AppointmentTime struct {
	BusinessTimes  []string `json:"business_times"`
	AvailableTimes []string `json:"available_times"`
}

func login(client *MyClient) error {
	logger.Info("Log in")
	signInURL := fmt.Sprintf("%s/users/sign_in", baseURI)
	loginReq, err := http.NewRequest("GET", signInURL, nil)
	if err != nil {
		return err
	}
	loginReq.Header = client.Header.Clone()
	initialResp, err := client.Do(loginReq)
	if err != nil {
		return err
	}
	defer initialResp.Body.Close()

	body, err := ioutil.ReadAll(initialResp.Body)
	if err != nil {
		return err
	}

	csrfToken := extractCSRFToken(string(body))
	if csrfToken == "" {
		return errors.New("failed to extract CSRF token")
	}

	data := url.Values{}
	data.Set("utf8", "✓")
	data.Set("user[email]", username)
	data.Set("user[password]", password)
	data.Set("policy_confirmed", "1")
	data.Set("commit", "Acessar")

	loginReq, err = http.NewRequest("POST", signInURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	client.Header.Set("X-CSRF-Token", csrfToken)

	loginReq.Header = client.Header.Clone()
	loginReq.Header.Set("Content-Type", contentType)

	resp, err := client.Do(loginReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		logger.Errorf("status code %v", resp.StatusCode)
	}

	return nil
}

func checkAvailableDate(header http.Header) (string, error) {
	client := GetClient()
	url := fmt.Sprintf("%s/schedule/%s/appointment/days/%s.json?appointments[expedite]=false", baseURI, scheduleID, facilityID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header = header

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			return "", UnauthError{}
		}
		return "", errors.New(fmt.Sprintf("Error status code %v", resp.StatusCode))
	}

	var days []AppointmentDay
	err = json.NewDecoder(resp.Body).Decode(&days)
	if err != nil {
		return "", err
	}

	if len(days) > 0 {
		return days[0].Date, nil
	}

	return "", nil
}

func checkAvailableTime(header http.Header, date string) (string, error) {
	client := GetClient()
	url := fmt.Sprintf("%s/schedule/%s/appointment/times/%s.json?date=%s&appointments[expedite]=false", baseURI, scheduleID, facilityID, date)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header = header
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var times AppointmentTime
	err = json.NewDecoder(resp.Body).Decode(&times)
	if err != nil {
		return "", err
	}

	if len(times.BusinessTimes) > 0 {
		return times.BusinessTimes[len(times.BusinessTimes)-1], nil
	}
	if len(times.AvailableTimes) > 0 {
		return times.AvailableTimes[len(times.AvailableTimes)-1], nil
	}

	return "", nil
}

func findToken(header *http.Header) string {
	apiURL := fmt.Sprintf("%s/schedule/%s/appointment", baseURI, scheduleID)
	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header = header.Clone()
	client := GetClient()
	resp, err := client.Do(req)
	if err != nil {
	}
	defer resp.Body.Close()
	return getAuthenticityToken(resp.Body)

}

func book(header *http.Header, date string, time string) error {
	client := GetClient()
	apiURL := fmt.Sprintf("%s/schedule/%s/appointment", baseURI, scheduleID)
	token := findToken(header)
	data := url.Values{}
	//data.Set("utf8", "✓")
	data.Set("authenticity_token", token)
	data.Set("confirmed_limit_message", "1")
	data.Set("use_consulate_appointment_capacity", "true")
	data.Set("appointments[consulate_appointment][facility_id]", facilityID)
	data.Set("appointments[consulate_appointment][date]", date)
	data.Set("appointments[consulate_appointment][time]", time)

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.Header = header.Clone()
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", apiURL)

	resp, err := client.Do(req)
	bodyBytes, _ := io.ReadAll(resp.Body)
	body := string(bodyBytes)

	if err != nil {
		logger.Errorf("status code: %v", resp.StatusCode)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Errorf("status code: %v", resp.StatusCode)
		return errors.New("failed to book appointment")
	}

	if strings.Contains(body, "Confirmation and Instructions") {
		logger.Infof("booked successfully on %s at %s", date, time)
	}

	return nil
}
