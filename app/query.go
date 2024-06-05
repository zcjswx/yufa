package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
)

var baseHeader *http.Header

type AppointmentDay struct {
	Date string `json:"date"`
}

type AppointmentTime struct {
	BusinessTimes  []string `json:"business_times"`
	AvailableTimes []string `json:"available_times"`
}

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func Init() {
	baseHeader = &http.Header{}
	baseHeader.Set("User-Agent", userAgent)
	baseHeader.Set("Accept-Encoding", "gzip, deflate, br")
	baseHeader.Set("Connection", "keep-alive")
	baseHeader.Set("Cache-Control", "no-cache")
	baseHeader.Set("Referer", baseURI)
	baseHeader.Set("Referrer-Policy", "strict-origin-when-cross-origin")
	baseHeader.Set("Accept", "*/*")
}

func login(client HttpClient) error {
	log.Println("Logging in")
	signInURL := fmt.Sprintf("%s/users/sign_in", baseURI)
	loginReq, err := http.NewRequest("GET", signInURL, nil)
	if err != nil {
		return err
	}
	loginReq.Header = baseHeader.Clone()
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

	baseHeader.Set("X-CSRF-Token", csrfToken)

	loginReq.Header = baseHeader.Clone()
	loginReq.Header.Set("Content-Type", contentType)
	loginReq.Header.Set("Cookie", getCookieBody(extractRelevantCookie(initialResp.Header.Get("Set-Cookie"))))

	resp, err := client.Do(loginReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("status code %v", resp.StatusCode)
	}

	cookies := getCookieBody(extractRelevantCookie(resp.Header.Get("Set-Cookie")))
	baseHeader.Set("Cookie", cookies)
	return nil
}

func checkAvailableDate(header http.Header) (string, error) {
	client := &http.Client{}
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
	client := &http.Client{}
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
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
	}
	defer resp.Body.Close()
	header.Set("Cookie", getCookieBody(extractRelevantCookie(resp.Header.Get("Set-Cookie"))))
	return getAuthenticityToken(resp.Body)

}

func book(header *http.Header, date string, time string) error {
	client := &http.Client{}
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
		log.Printf("status code: %v", resp.StatusCode)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("status code: %v", resp.StatusCode)
		return errors.New("failed to book appointment")
	}

	if strings.Contains(body, "Confirmation and Instructions") {
		log.Printf("booked successfully on %s at %s", date, time)
	}

	return nil
}
