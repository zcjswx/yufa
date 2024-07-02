package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

/*
	1. user log in
	2. check available date in multiple cities in a DateManager():
		-> DataWorker(ctx, Ottawa)
		-> DataWorker(ctx, Montreal)
		-> DataWorker(ctx, Toronto)
	3. if no date find run #2, else, stop all workers, and do book()
*/

var user *User

type User struct {
	Username          string
	Password          string
	ScheduleID        string
	FacilityIDList    []CityID
	CurrentBookedDate string
	client            *MyClient
}

func NewUser(config Config) *User {
	u := &User{
		Username:          config.Username,
		Password:          config.Password,
		ScheduleID:        config.ScheduleID,
		FacilityIDList:    config.FacilityIDList,
		CurrentBookedDate: config.CurrentBookedDate,
	}
	user = u
	return user
}

type BookParam struct {
	FacilityID CityID
	Date       string
	Time       string
}

func (u *User) dateCheckingManager() {

	outChan := make(chan *BookParam)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, facilityID := range u.FacilityIDList {
		go u.dateCheckingWorker(ctx, cancel, outChan, facilityID)
		time.Sleep(GetRandSecond())
	}

	param := <-outChan

	err := u.book(*param)
	if err != nil {
		logger.Error(err)
	} else {
		logger.Infof("Booked time at %s on %s at %s", GetCityName(param.FacilityID), param.Date, param.Time)
	}
	os.Exit(0)
}

func (u *User) dateCheckingWorker(ctx context.Context, cancel context.CancelFunc, outChan chan *BookParam, facilityID CityID) {

	for {
		select {
		case <-ctx.Done():
			return
		default:

			checkTimeToSkip := func() bool {
				_, minutes, _ := time.Now().Clock()
				if minutes%5 != 0 {
					return false
				} else {
					return true
				}
			}

			if checkTimeToSkip() {
				time.Sleep(GetRandSecond())
				continue
			}

			date, err := u.getAvailableDate(facilityID)
			if err != nil {
				logger.Error(err)

				// need to relog-in, have everything refreshed to log in successfully
				if errors.Is(err, UnauthError{}) {
					if err2 := u.login(); err2 != nil {
						cancel()
					}
				}
			}
			if date == "" {
				logger.Infof("No date available at %s", GetCityName(facilityID))
			} else if date > u.CurrentBookedDate {
				logger.Infof("Nearest date is further than already booked (%s vs %s) at %s", u.CurrentBookedDate, date, GetCityName(facilityID))
			} else {
				logger.Infof("Found data at %s on %s", GetCityName(facilityID), date)
				u.CurrentBookedDate = date
				availableTime, err := u.getAvailableTime(date, facilityID)
				if err != nil {
					logger.Error(err)
					continue
				}
				bookParam := &BookParam{
					FacilityID: facilityID,
					Date:       date,
					Time:       availableTime,
				}
				outChan <- bookParam
				cancel()
			}

			time.Sleep(GetRandSecond())
		}
	}
}

func (u *User) login() error {
	logger.Info("Log in")
	signInURL := fmt.Sprintf("%s/users/sign_in", GetConfig().BaseURI)
	loginReq, err := http.NewRequest("GET", signInURL, nil)
	if err != nil {
		return err
	}
	loginReq.Header = u.client.Header.Clone()
	initialResp, err := u.client.Do(loginReq)
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
	data.Set("user[email]", u.Username)
	data.Set("user[password]", u.Password)
	data.Set("policy_confirmed", "1")
	data.Set("commit", "Acessar")

	loginReq, err = http.NewRequest("POST", signInURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	u.client.Header.Set("X-CSRF-Token", csrfToken)

	loginReq.Header = u.client.Header.Clone()
	loginReq.Header.Set("Content-Type", GetConfig().ContentType)

	resp, err := u.client.Do(loginReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		logger.Errorf("status code %v", resp.StatusCode)
	}

	return nil
}

func (u *User) getAvailableDate(facilityID CityID) (string, error) {

	url := fmt.Sprintf("%s/schedule/%s/appointment/days/%v.json?appointments[expedite]=false", GetConfig().BaseURI, u.ScheduleID, facilityID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header = u.client.Header.Clone()

	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := u.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized {
			return "", UnauthError{}
		}

		// Request redirected to log in page, and get 404, cookie expires every 4 hours
		// Todo log in before cookie expires
		if resp.StatusCode == http.StatusNotFound {
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

func (u *User) getAvailableTime(date string, facilityID CityID) (string, error) {
	url := fmt.Sprintf("%s/schedule/%s/appointment/times/%v.json?date=%s&appointments[expedite]=false", GetConfig().BaseURI, u.ScheduleID, facilityID, date)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header = u.client.Header.Clone()
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err := u.client.Do(req)
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

func (u *User) book(param BookParam) error {
	apiURL := fmt.Sprintf("%s/schedule/%s/appointment", GetConfig().BaseURI, u.ScheduleID)
	token := findToken(u.client.Header)
	data := url.Values{}
	//data.Set("utf8", "✓")
	data.Set("authenticity_token", token)
	data.Set("confirmed_limit_message", "1")
	data.Set("use_consulate_appointment_capacity", "true")
	data.Set("appointments[consulate_appointment][facility_id]", strconv.Itoa(int(param.FacilityID)))
	data.Set("appointments[consulate_appointment][date]", param.Date)
	data.Set("appointments[consulate_appointment][time]", param.Time)

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.Header = u.client.Header.Clone()
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", apiURL)

	resp, err := u.client.Do(req)
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
		logger.Infof("booked successfully on %s at %s", param.Date, param.Time)
	}

	return nil
}

func (u *User) findToken(header *http.Header) string {
	apiURL := fmt.Sprintf("%s/schedule/%s/appointment", GetConfig().BaseURI, GetConfig().ScheduleID)
	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header = header.Clone()

	resp, err := u.client.Do(req)
	if err != nil {
	}
	defer resp.Body.Close()
	return getAuthenticityToken(resp.Body)

}
