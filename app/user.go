package app

var user *User

type User struct {
	Username          string
	Password          string
	ScheduleID        string
	FacilityID        string
	CurrentBookedDate string
	client            *HttpClient
}
