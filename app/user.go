package app

import (
	"context"
	"errors"
	"time"
)

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
	}

	<-outChan

}

func (u *User) dateCheckingWorker(ctx context.Context, cancel context.CancelFunc, outChan chan *BookParam, facilityID CityID) {

	header := u.client.Header.Clone()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			date, err := checkAvailableDate(header)
			if err != nil {
				logger.Error(err)

				// need to relog-in, have everything refreshed to log in successfully
				if errors.Is(err, UnauthError{}) {
					cancel()
					close(outChan)
					return
				}
			}
			if date == "" {
				logger.Infof("No dates available")
			} else if date > currentBookedDate {
				logger.Infof("Nearest date is further than already booked (%s vs %s)", currentBookedDate, date)
			} else {
				logger.Infof("Found data on %s", date)
				currentBookedDate = date
				availableTime, err := checkAvailableTime(header, date)
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
			}

			time.Sleep(GetRandSecond())
		}
	}
}
