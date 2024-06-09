package app

import (
	"errors"
	"net/http"
	"os"
	"time"
)

func Process() {
	config, err := readConfig("config.yaml")
	if err != nil {
		logger.Fatal(err)
	}
	setupConfig(config)
	logger.Infof("Initializing with current date %s", currentBookedDate)
	client := GetClient()
	user := NewUser(*config)
	user.client = client
	err = login(client)
	if err != nil {
		logger.Fatalf("Login failed: %v", err)
	}

	for {
		func(h *http.Header) {
			date, err := checkAvailableDate(h.Clone())
			if err != nil {
				logger.Error(err)

				// need to relog-in, have everything refreshed to login successfully
				if errors.Is(err, UnauthError{}) {
					client = NewClient()
					err = login(client)
					if err != nil {
						logger.Errorf("Login failed: %v", err)
					}
				}
				return
			}

			if date == "" {
				logger.Infof("No dates available")
			} else if date > currentBookedDate {
				logger.Infof("Nearest date is further than already booked (%s vs %s)", currentBookedDate, date)
			} else {
				logger.Infof("Found data on %s", date)
				currentBookedDate = date
				availableTime, err := checkAvailableTime(h.Clone(), date)
				if err != nil {
					logger.Error(err)
					return
				}
				err = book(h, date, availableTime)
				if err != nil {
					logger.Error(err)
				} else {
					//
					logger.Infof("Booked time at %s %s", date, availableTime)
					os.Exit(0)
				}
			}
		}(client.Header)

		time.Sleep(GetRandSecond())
	}
}

func Con() {
	config, err := readConfig("config.yaml")
	if err != nil {
		logger.Fatal(err)
	}
	setupConfig(config)
	logger.Infof("Initializing with current date %s", currentBookedDate)
	client := GetClient()
	user := NewUser(*config)
	user.client = client

	err = user.login()
	if err != nil {
		logger.Fatalf("Login failed: %v", err)
	}
	user.dateCheckingManager()
}
