package app

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

var (
	// deprecated
	username,
	password,
	scheduleID,
	facilityID,
	baseURI,
	contentType,
	currentBookedDate string
	config *Config
)

type Config struct {
	Username          string   `yaml:"username"`
	Password          string   `yaml:"password"`
	ScheduleID        string   `yaml:"schedule_id"`
	FacilityID        string   `yaml:"facility_id"`
	FacilityIDList    []CityID `yaml:"facility_id_list"`
	BaseURI           string   `yaml:"base_uri"`
	UserAgent         string   `yaml:"user_agent"`
	ContentType       string   `yaml:"content_type"`
	CurrentBookedDate string   `yaml:"current_booked_date"`
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
	config = &cfg

	return &cfg, nil
}

// deprecated
func setupConfig(config *Config) {
	username = config.Username
	password = config.Password
	scheduleID = config.ScheduleID
	facilityID = config.FacilityID
	baseURI = config.BaseURI
	contentType = config.ContentType
	currentBookedDate = config.CurrentBookedDate
}

func GetConfig() Config {
	if config == nil {
		logger.Fatalf("config is empty")
	}
	return *config
}
