package app

import (
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

var (
	username, password, scheduleID, facilityID, baseURI, userAgent, contentType, currentBookedDate string
)

type Config struct {
	Username          string `yaml:"username"`
	Password          string `yaml:"password"`
	ScheduleID        string `yaml:"schedule_id"`
	FacilityID        string `yaml:"facility_id"`
	BaseURI           string `yaml:"base_uri"`
	UserAgent         string `yaml:"user_agent"`
	ContentType       string `yaml:"content_type"`
	CurrentBookedDate string `yaml:"current_booked_date"`
}

func readConfig(filename string) (*Config, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(buf, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func setupConfig(config *Config) {
	username = config.Username
	password = config.Password
	scheduleID = config.ScheduleID
	facilityID = config.FacilityID
	baseURI = config.BaseURI
	userAgent = config.UserAgent
	contentType = config.ContentType
	currentBookedDate = config.CurrentBookedDate
}
