package main

import (
	"io"
	"encoding/json"
	"io/ioutil"
	"fmt"
	"time"
	"net/http"

	"github.com/gorilla/mux"
)

type Hour struct {
	Number int	`json:"number"`
	Activations	int `json:"activations"`
	Warnings	int `json:"warnings"`
	ActivationMinutes	int `json:"activationminutes"`
	ActivationFlag  bool
	HourOfLastActivation int
	MinuteOfLastActivation int
}

type Day struct {
	Number  int	`json:"number"`
	Weekday int	`json:"weekday"`
	Hours	[]Hour  `json:"hours"`
}

type Month struct {
	Year	int	`json:"year"`
	Number	int	`json:"number"`
	Days	[]Day   `json:"days"`
}

type TimeOfDay struct {
	Hour	int `json:"hour"`
	Minute	int `json:"minute"`
}

type ClassSchedule struct {
	Start	TimeOfDay `json:"start"`
	End	TimeOfDay `json:"end"`
}

type Schedule struct {	
	DaysAWeek	int 		 `json:"daysaweek"`
	ClassHours	[]ClassSchedule  `json:"classhours"`
}

type Ledger struct {
	DeviceId	string 	`json:"deviceid"`
	Months		[]Month	`json:"months"`	
}

type GlobalLedger struct {
	Ledgers  []Ledger	`json:"ledgers"`	
}

///////////////////////
// standard schedule //
///////////////////////

var First ClassSchedule = ClassSchedule{TimeOfDay{8,30},TimeOfDay{9,20}}
var Second ClassSchedule = ClassSchedule{TimeOfDay{9,25},TimeOfDay{10,15}}
var Third ClassSchedule = ClassSchedule{TimeOfDay{10,20},TimeOfDay{11,10}}
var Fourth ClassSchedule = ClassSchedule{TimeOfDay{11,35},TimeOfDay{12,25}}
var Fifth ClassSchedule = ClassSchedule{TimeOfDay{12,25},TimeOfDay{13,15}}
var Sixth ClassSchedule = ClassSchedule{TimeOfDay{13,20},TimeOfDay{14,10}}
var StdSchedule Schedule = Schedule{5, []ClassSchedule{First, Second, Third, Fourth, Fifth, Sixth}}

var CurrentSchedule Schedule = StdSchedule

var GlobalLedgerFile string = "./data/ledger.json"
var CurrentGlobalLedger GlobalLedger = GlobalLedger{}

func LoadGlobalLedgerFromFile() {
	file, err := ioutil.ReadFile(GlobalLedgerFile)
	if err != nil {
		CurrentGlobalLedger = GlobalLedger{}
	} else {
		err2 := json.Unmarshal([]byte(file), &CurrentGlobalLedger)
		if err2 != nil {
			fmt.Println("Error unmarshalling ledger data, reverting to clean ledger")
			CurrentGlobalLedger = GlobalLedger{}
		}
	}
}

func SaveGlobalLedgerToFile() {
	file, err := json.MarshalIndent(CurrentGlobalLedger, "", " ")
	if err != nil {
		fmt.Println("Error marshalling ledger data")
	} else {
		err2 := ioutil.WriteFile(GlobalLedgerFile, file, 0644)
		if err2 != nil {
			fmt.Println("Error saving ledger data")
		}
	}
}

func GetDayNumber(sched Schedule) int {
	return int(time.Now().Weekday())
}

func GetTimeOfDay(sched Schedule) TimeOfDay {
	t := time.Now()
	return TimeOfDay{t.Hour(),t.Minute()}
}

func IsWithin(classTime TimeOfDay, classB ClassSchedule) bool {
	if classTime.Hour < classB.Start.Hour {
		return false
	}
	if classTime.Hour > classB.End.Hour {
		return false
	}
	if (classTime.Minute < classB.Start.Minute) && (classTime.Hour == classB.Start.Hour) {
		return false
	}
	if (classTime.Minute > classB.End.Minute) && (classTime.Hour == classB.End.Hour) {
		return false
	}
	return true
}

func GetClassHour(sched Schedule, timeOfDay TimeOfDay) int {
	for i,v:=range(sched.ClassHours) {
		if IsWithin(timeOfDay, v) {
			return i+1
		}
	}
	return 0
}

func DifferenceInMinutes(timeA TimeOfDay, timeB TimeOfDay) int {
	if(timeA.Hour == timeB.Hour) {
		diff := timeB.Minute - timeA.Minute
		if diff == 0 {
			diff = 1
		}
		return diff
	} else {
		if(timeA.Hour < timeB.Hour) {
			return 60 + DifferenceInMinutes(TimeOfDay{timeA.Hour+1, timeA.Minute}, timeB)
		} else {
			return 60 + DifferenceInMinutes(TimeOfDay{timeB.Hour+1, timeB.Minute}, timeA)
		}
	}
}

func GetLedger(deviceId string) *Ledger {
	for _,v:=range(CurrentGlobalLedger.Ledgers) {
		if(deviceId == v.DeviceId) {
			return &v
		}
	}
	NewLedger := Ledger{deviceId, []Month{}}
	CurrentGlobalLedger.Ledgers = append(CurrentGlobalLedger.Ledgers, NewLedger)
	return &NewLedger
}

func GetLedgerOrNil(deviceId string) *Ledger {
        for _,v:=range(CurrentGlobalLedger.Ledgers) {
                if(deviceId == v.DeviceId) {
                        return &v
                }
        }
        return nil
}

func (ledger *Ledger)GetMonth(year int, number int) *Month {
	for _,v:=range(ledger.Months) {
		if v.Number == number && v.Year == year {
			return &v
		}
	}
	NewMonth := Month{year, number, []Day{}}
	ledger.Months = append(ledger.Months, NewMonth)
	return &NewMonth
}

func (ledger *Ledger)AddEvent(event string, year int, month int, day int, weekday int, hour int, minute int) {
	monthRef := ledger.GetMonth(year, month)
	monthRef.AddEvent(event, day, weekday, hour, minute)
}

func (month *Month)AddEvent(event string, day int, weekday int, hour int, minute int) {
	dayRef := month.GetDay(day, weekday)
	dayRef.AddEvent(event, hour, minute)
}

func (month *Month)GetDay(day int, weekday int) *Day {
	for _,v:=range(month.Days) {
                if v.Number == day {
                        return &v
                }
        }
        NewDay := Day{day, weekday, []Hour{}}
	return &NewDay
}

func (day *Day)GetHour(sched Schedule, hour int, minute int) *Hour {
	hourNumber := GetClassHour(sched, TimeOfDay{hour,minute}) 
        for _,v:=range(day.Hours) {
                if v.Number == hourNumber {
                        return &v
                }
        }
        NewHour := Hour{hourNumber, 0, 0, 0, false, 0, 0}
	day.Hours = append(day.Hours, NewHour)
	return &NewHour
}

func (day *Day)AddEvent(event string, hour int, minute int) {
	hourRef := day.GetHour(CurrentSchedule, hour, minute)
	hourRef.AddEvent(event, hour, minute)
} 

func (hourRef *Hour)AddEvent(event string, hour int, minute int) {
	if(event == "activation") {
		hourRef.Activations++
		if(hourRef.ActivationFlag == true) { // should not happen, but just in case...
			hourRef.ActivationFlag = false
			duration := DifferenceInMinutes(TimeOfDay{hourRef.HourOfLastActivation, hourRef.MinuteOfLastActivation}, TimeOfDay{hour, minute})
			hourRef.ActivationMinutes += duration
		} else {
			hourRef.ActivationFlag = true
		}
		hourRef.HourOfLastActivation = hour
                hourRef.MinuteOfLastActivation = minute
	} else if(event == "warning") {
		hourRef.Warnings++
	} else if(event == "deactivation") {
		hourRef.ActivationFlag = !hourRef.ActivationFlag
		if(hourRef.ActivationFlag == false) {
			duration := DifferenceInMinutes(TimeOfDay{hourRef.HourOfLastActivation, hourRef.MinuteOfLastActivation}, TimeOfDay{hour, minute})
			hourRef.ActivationMinutes += duration
		}
	}
}

func AddEventToCurrentTime(source string, event string) {
	t := time.Now()
	year := int(t.Year())
	month := int(t.Month())
	hour := int(t.Hour())
	day := int(t.Day())
	minute := int(t.Minute())
	weekday := int(t.Weekday())
	fmt.Println("Registering event ", event, " for device ", source, " at ", year, " ", month, " ", hour, " ", minute, " weekday: ", weekday)
	ledger := GetLedger(source)
	ledger.AddEvent(event, year, month, day, weekday, hour, minute)
	SaveGlobalLedgerToFile()
}

func AddActivationToCurrentTime(source string) {
	AddEventToCurrentTime(source, "activation")
}

func AddWarningToCurrentTime(source string) {
	AddEventToCurrentTime(source, "warning")
}

func AddDeactivationToCurrentTime(source string) {
        AddEventToCurrentTime(source, "deactivation")
}

///////////////
// REST API  //
///////////////

func JSONResponseFromString(w http.ResponseWriter, res string) {
        w.Header().Set("Content-Type", "application/json; charset=UTF-8")
        w.WriteHeader(http.StatusOK)
        io.WriteString(w, res)
}

func RegisterEvent(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	source := params["source"]
	if(params["type"] == "activation") {
		AddActivationToCurrentTime(source)
		JSONResponseFromString(w, "{\"result\":\"success\"}")
	} else if (params["type"] == "warning") {
		AddWarningToCurrentTime(source)
		JSONResponseFromString(w, "{\"result\":\"success\"}")
	} else if (params["type"] == "deactivation") {
		AddDeactivationToCurrentTime(source)
		JSONResponseFromString(w, "{\"result\":\"success\"}")
	} else {
		JSONResponseFromString(w, "{\"result\":\"wrong event type\"}")
	}
}

func GetDailySchedule(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
  	json.NewEncoder(w).Encode(CurrentSchedule)
}

func GetMonthSheet(w http.ResponseWriter, r *http.Request) {
	JSONResponseFromString(w, "{\"result\":\"to be continued\"}")
}

func DeleteMonthSheet(w http.ResponseWriter, r *http.Request) {
	JSONResponseFromString(w, "{\"result\":\"to be continued\"}")
}

func GetAvailableMonthSheets(w http.ResponseWriter, r* http.Request) {
	JSONResponseFromString(w, "{\"result\":\"to be continued\"}")
}

func GetDeviceLedger(w http.ResponseWriter, r* http.Request) {
	params := mux.Vars(r)
	deviceLedger := GetLedgerOrNil(params["unit"])
	if deviceLedger != nil {
		w.Header().Set("Content-Type", "application/json")
        	json.NewEncoder(w).Encode(deviceLedger)
	} else {
		JSONResponseFromString(w, "{}")
	}
}

func GetFullLedger(w http.ResponseWriter, r* http.Request) {
	w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(CurrentGlobalLedger)
}

///////////////
//  Main     //
///////////////

func main() {

	LoadGlobalLedgerFromFile()

	router := mux.NewRouter()
	router.HandleFunc("/event/{source}/{type}", RegisterEvent).Methods("GET")
	router.HandleFunc("/event/{source}/{type}", RegisterEvent).Methods("POST")
	router.HandleFunc("/monthsheet/{month}", GetMonthSheet).Methods("GET")
	router.HandleFunc("/schedule", GetDailySchedule).Methods("GET")
	router.HandleFunc("/monthsheets/{unit}/{month}", DeleteMonthSheet).Methods("DELETE")
	router.HandleFunc("/monthsheets/{unit}", GetAvailableMonthSheets).Methods("GET")

	router.HandleFunc("/ledger/{unit}", GetDeviceLedger).Methods("GET")
	router.HandleFunc("/ledger", GetFullLedger).Methods("GET")

	fmt.Println(DifferenceInMinutes(TimeOfDay{8,35},TimeOfDay{8,37})) // 2
	fmt.Println(DifferenceInMinutes(TimeOfDay{8,51},TimeOfDay{9,12})) // 21
	fmt.Println(DifferenceInMinutes(TimeOfDay{10,10},TimeOfDay{9,50})) // 20
	
	http.ListenAndServe(":9999", router)
}
