
An automation tool to get us released from the endless waiting queue to enter US.

- Easy running and deployment
- Few dependency
- Pure api calling (No Selenium)


### How to run

- Go environment
- Rename [config.yaml.template](config.yaml.template) to [config.yaml](config.yaml)
  - In [config.yaml](config.yaml), fill in:
    * username
    * password
    * schedule_id: Manual input required
      ```
      1. Login and open 'Reschedule' page
      2. Look at the 'Address Bar'(URL), it should look like 
      'https://ais.usvisa-info.com/en-ca/niv/schedule/xxxxxxxx/appointment',
      fill copy the numbers and paste them here
    * current_booked_date: The date you already booked your appointment, in the format "yyyy-mm-dd"
    * facility_id_list:
      ``` 
      Calgary     = 89
      Halifax     = 90
      Montreal    = 91
      Ottawa      = 92
      QuebecCity  = 93
      Toronto     = 94
      Vancouver   = 95
    * facility_id: deprecated
- `go run ./cmd`