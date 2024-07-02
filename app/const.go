package app

type CityID int

const (
	Calgary    CityID = 89
	Halifax    CityID = 90
	Montreal   CityID = 91
	Ottawa     CityID = 92
	QuebecCity CityID = 93
	Toronto    CityID = 94
	Vancouver  CityID = 95

	UrlAppointmentSuffix string = "%s/schedule/%s/appointment"
	UrlSignIn            string = "%s/users/sign_in"
)

var cityNames = map[CityID]string{
	Calgary:    "Calgary",
	Halifax:    "Halifax",
	Montreal:   "Montreal",
	Ottawa:     "Ottawa",
	QuebecCity: "Quebec City",
	Toronto:    "Toronto",
	Vancouver:  "Vancouver",
}

func GetCityName(id CityID) string {
	name, ok := cityNames[id]
	if !ok {
		return "City ID Not Found"
	}
	return name
}
