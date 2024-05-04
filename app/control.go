package app

import (
	"log"
	"net/http"
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
		date, err := checkAvailableDate(baseHeader.Clone())
		if err != nil {
			log.Println(err)
			continue
		}
		if date == "" {
			log.Println("No dates available")
		} else if date > currentBookedDate {
			log.Printf("Nearest date is further than already booked (%s vs %s)", currentBookedDate, date)
		} else {
			currentBookedDate = date
			availableTime, err := checkAvailableTime(baseHeader.Clone(), date)
			if err != nil {
				log.Println(err)
				continue
			}
			err = book(baseHeader.Clone(), date, availableTime)
			if err != nil {
				log.Println(err)
			} else {
				log.Printf("Booked time at %s %s", date, availableTime)
			}
		}
		time.Sleep(3 * time.Second)
	}
}
