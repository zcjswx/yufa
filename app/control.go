package app

import (
	"errors"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

func Process() {
	config, err := readConfig("config.yaml")
	if err != nil {
		log.Fatal(err)
	}
	setupConfig(config)
	Init()
	log.Printf("Initializing with current date %s", currentBookedDate)
	client := &http.Client{}
	err = login(client)
	if err != nil {
		log.Fatalf("Login failed: %v", err)
	}

	for {
		func(h *http.Header) {
			date, err := checkAvailableDate(h.Clone())
			if err != nil {
				log.Println(err)

				// need to relog-in, have everything refreshed to login successfully
				if errors.Is(err, UnauthError{}) {
					Init()
					client = &http.Client{}
					err = login(client)
					if err != nil {
						log.Printf("Login failed: %v", err)
					}
				}
				return
			}

			if date == "" {
				log.Println("No dates available")
			} else if date > currentBookedDate {
				log.Printf("Nearest date is further than already booked (%s vs %s)", currentBookedDate, date)
			} else {
				log.Printf("Found data on %s", date)
				currentBookedDate = date
				availableTime, err := checkAvailableTime(h.Clone(), date)
				if err != nil {
					log.Println(err)
					return
				}
				err = book(h, date, availableTime)
				if err != nil {
					log.Println(err)
				} else {
					//
					log.Printf("Booked time at %s %s", date, availableTime)
					os.Exit(0)
				}
			}
		}(baseHeader)

		time.Sleep(
			func() time.Duration {
				numbers := []int{11, 13, 17, 19, 23, 29, 31, 37, 41, 43, 47}
				rand.Seed(time.Now().UnixNano())
				randomIndex := rand.Intn(len(numbers))
				return time.Duration(numbers[randomIndex]) * time.Second
			}())
	}
}
