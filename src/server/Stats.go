package main

import (
	"io"
	"encoding/json"
	"strconv"
	"io/ioutil"
	"fmt"
	"time"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/360EntSecGroup-Skylar/excelize"
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

type MonthHeader struct {
	Year	int	`json:"year"`
	Number  int	`json:"number"`
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
	return int(GetDateTime().Weekday())
}

func GetTimeOfDay(sched Schedule) TimeOfDay {
	t := GetDateTime()
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
	for i,v:=range(CurrentGlobalLedger.Ledgers) {
		if(deviceId == v.DeviceId) {
			return &CurrentGlobalLedger.Ledgers[i]
		}
	}
	NewLedger := Ledger{deviceId, []Month{}}
	CurrentGlobalLedger.Ledgers = append(CurrentGlobalLedger.Ledgers, NewLedger)
	return &CurrentGlobalLedger.Ledgers[len(CurrentGlobalLedger.Ledgers)-1]
}

func GetLedgerOrNil(deviceId string) *Ledger {
        for i,v:=range(CurrentGlobalLedger.Ledgers) {
                if(deviceId == v.DeviceId) {
                        return &CurrentGlobalLedger.Ledgers[i]
                }
        }
        return nil
}

func (ledger *Ledger)GetMonth(year int, number int) *Month {
	for i,v:=range(ledger.Months) {
		if v.Number == number && v.Year == year {
			return &ledger.Months[i]
		}
	}
	NewMonth := Month{year, number, []Day{}}
	ledger.Months = append(ledger.Months, NewMonth)
	return &ledger.Months[len(ledger.Months)-1]
}

func (ledger *Ledger)GetMonthOrNil(year int, number int) *Month {
        for i,v:=range(ledger.Months) {
                if v.Number == number && v.Year == year {
                        return &ledger.Months[i]
                }
        }
	return nil
}

func (ledger *Ledger)AddEvent(event string, year int, month int, day int, weekday int, hour int, minute int) {
	monthRef := ledger.GetMonth(year, month)
	fmt.Println("Returned to AddEvent. ledger.Months length: ", len(ledger.Months))
	monthRef.AddEvent(event, day, weekday, hour, minute)
}

func (ledger *Ledger)MonthHeaders() []MonthHeader {
	var result []MonthHeader
	for _, v := range(ledger.Months) {
		m := MonthHeader{v.Year, v.Number}
		result = append(result, m)
	}
	return result
}

func (month *Month)AddEvent(event string, day int, weekday int, hour int, minute int) {
	dayRef := month.GetDay(day, weekday)
	dayRef.AddEvent(event, hour, minute)
}

func (month *Month)GetDay(day int, weekday int) *Day {
	for i,v:=range(month.Days) {
                if v.Number == day {
                        return &month.Days[i]
                }
        }
        NewDay := Day{day, weekday, []Hour{}}
	month.Days = append(month.Days, NewDay)
	return &month.Days[len(month.Days)-1]
}

func (month *Month)GetDayOrNil(day int) *Day {
        for i,v:=range(month.Days) {
                if v.Number == day {
                        return &month.Days[i]
                }
        }
	return nil
}

func (day *Day)GetHour(sched Schedule, hour int, minute int) *Hour {
	hourNumber := GetClassHour(sched, TimeOfDay{hour,minute}) 
        for i,v:=range(day.Hours) {
                if v.Number == hourNumber {
                        return &day.Hours[i]
                }
        }
        NewHour := Hour{hourNumber, 0, 0, 0, false, 0, 0}
	day.Hours = append(day.Hours, NewHour)
	return &day.Hours[len(day.Hours)-1]
}

func (day *Day)GetHourOrNil(hourNumber int) *Hour {
        for i,v:=range(day.Hours) {
                if v.Number == hourNumber {
                        return &day.Hours[i]
                }
        }
	return nil
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
	t := GetDateTime()
	year := int(t.Year())
	month := int(t.Month())
	hour := int(t.Hour())
	day := int(t.Day())
	minute := int(t.Minute())
	weekday := int(t.Weekday())
	fmt.Println("Registering event ", event, " for device ", source, " at ", year, " ", month, " ", hour, " ", minute, " weekday: ", weekday)
	ledger := GetLedger(source)
	fmt.Printf("address of ledger: %p\n", ledger)
	fmt.Printf("address of first element: %p\n\n", &CurrentGlobalLedger.Ledgers[0])
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

/////////////////
// Time & date //
/////////////////

var DateTimeSet bool = false
var CurrentYear int
var CurrentMonth int
var CurrentDay int
var CurrentHour int
var CurrentMinute int

func GetDateTime() time.Time {
	if DateTimeSet {
		return time.Date(CurrentYear, time.Month(CurrentMonth), CurrentDay, CurrentHour, CurrentMinute, 0, 0, time.UTC)
	} else {
		return time.Now()
	}
}

///////////////
// Generate  //
///////////////

func ExtractEvents(monthRef *Month, day int, hour int) (int, int, int) {
	if monthRef == nil {
		return -1, -1, -1
	}
	dayRef := monthRef.GetDayOrNil(day)
	if dayRef == nil {
		return -1, -1, -1
	}
	hourRef := dayRef.GetHourOrNil(hour)
	if hourRef == nil {
		return -1, -1, -1
	}
	return hourRef.Warnings, hourRef.Activations, hourRef.ActivationMinutes
}

func DayOfWeekFromInt_es(weekday int) string {
	switch(weekday) {
		case 1: return "Lunes"
		case 2: return "Martes"
		case 3: return "Miércoles"
		case 4: return "Jueves"
		case 5: return "Viernes"	
	}
	return "N/A"
}

func IntToLetter(i int) string {
	return string(64+i)
}

func CreateMonthlyReport(device string, year int, month int) {
	ledger := GetLedgerOrNil(device)
	if ledger == nil {
		return
	}
	monthRef := ledger.GetMonthOrNil(year, month)

	f := excelize.NewFile()

	styleWarn, _ := f.NewStyle(`{"fill":{"type":"pattern","color":["#EEBB00"],"pattern":1}}`)
	styleAct, _ := f.NewStyle(`{"fill":{"type":"pattern","color":["#EE2211"],"pattern":1}}`)
	styleMins, _ := f.NewStyle(`{"fill":{"type":"pattern","color":["#AA22DD"],"pattern":1}}`)

	nDays := time.Date(year, time.Month(month+1), 0, 12, 0, 0, 0, time.UTC).Day()
	nHours := len(CurrentSchedule.ClassHours)
	f.SetCellValue("Sheet1", "A2", "Día")
	for h := 0 ; h < nHours ; h++ {
		f.SetCellValue("Sheet1", IntToLetter(3+h*3) + "1", "Hora " + strconv.Itoa(h+1))
		f.SetCellValue("Sheet1", IntToLetter(3+h*3) + "2", "Adv.")
		f.SetCellValue("Sheet1", IntToLetter(4+h*3) + "2", "Act.")
		f.SetCellValue("Sheet1", IntToLetter(5+h*3) + "2", "M. act.") 
	}
	row := 3
	for d := 1 ; d <= nDays ; d++ {
		then := time.Date(year, time.Month(month), d, 12, 0, 0, 0, time.UTC)
		weekday := int(then.Weekday())
		if weekday < 6 && weekday > 0 {
			f.SetCellValue("Sheet1", "A" + strconv.Itoa(row), DayOfWeekFromInt_es(weekday))
			f.SetCellValue("Sheet1", "B" + strconv.Itoa(row), strconv.Itoa(d)) 
			for h := 0 ; h < nHours ; h++ {
                		warnings, activations, minact := ExtractEvents(monthRef, d, h+1)
				if warnings != -1 {
					f.SetCellValue("Sheet1", IntToLetter(3+h*3) + strconv.Itoa(row), strconv.Itoa(warnings))
                			f.SetCellStyle("Sheet1", IntToLetter(3+h*3) + strconv.Itoa(row), IntToLetter(3+h*3) + strconv.Itoa(row), styleWarn)
					f.SetCellValue("Sheet1", IntToLetter(4+h*3) + strconv.Itoa(row), strconv.Itoa(activations))
                			f.SetCellStyle("Sheet1", IntToLetter(4+h*3) + strconv.Itoa(row), IntToLetter(4+h*3) + strconv.Itoa(row), styleAct)
					f.SetCellValue("Sheet1", IntToLetter(5+h*3) + strconv.Itoa(row), strconv.Itoa(minact))
        				f.SetCellStyle("Sheet1", IntToLetter(5+h*3) + strconv.Itoa(row), IntToLetter(5+h*3) + strconv.Itoa(row), styleMins)
				}
			}
			row++
		} 
		if weekday == 6 {
			row++
		}	
	}
	//f.setSheet(0, "Informe - " + device)

	err := f.SaveAs("./data/" + device + "-" + strconv.Itoa(year) + "-" + strconv.Itoa(month) + ".xls")
	if err != nil {
		fmt.Println(err)
	}
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
		fmt.Println("Length of months : " , len(deviceLedger.Months))
		w.Header().Set("Content-Type", "application/json")
        	json.NewEncoder(w).Encode(deviceLedger)
	} else {
		JSONResponseFromString(w, "{}")
	}
}

func GetReport(w http.ResponseWriter, r* http.Request) {
        params := mux.Vars(r)
	device := params["unit"]
	year, _ := strconv.Atoi(params["year"])
	month, _ := strconv.Atoi(params["month"])
	CreateMonthlyReport(device, year, month)
	JSONResponseFromString(w, "{\"result\":\"to be continued\"}")
}

func GetFullLedger(w http.ResponseWriter, r* http.Request) {
	w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(CurrentGlobalLedger)
}

func SetDate(w http.ResponseWriter, r* http.Request) {
	params := mux.Vars(r)
	year, _ := strconv.Atoi(params["year"])
	month, _ := strconv.Atoi(params["month"])
	day, _ := strconv.Atoi(params["day"])
	hour, _ := strconv.Atoi(params["hour"])
	minute, _ := strconv.Atoi(params["minute"])
	DateTimeSet = true
	CurrentYear = year
	CurrentMonth = month
	CurrentDay = day
	CurrentHour = hour
	CurrentMinute = minute
	fmt.Println("Datetime set to: " , GetDateTime())
	JSONResponseFromString(w, "{\"result\":\"success\"}")
}

func GetMonthsWithData(w http.ResponseWriter, r* http.Request) {
        params := mux.Vars(r)
	ledger := GetLedgerOrNil(params["unit"])
	if ledger != nil {
		months := ledger.MonthHeaders()
		w.Header().Set("Content-Type", "application/json")
        	json.NewEncoder(w).Encode(months)
	} else {
        	JSONResponseFromString(w, "[]")
	}
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

	router.HandleFunc("/date/{year}/{month}/{day}/{hour}/{minute}", SetDate).Methods("GET")
	router.HandleFunc("/months/{unit}", GetMonthsWithData).Methods("GET")

	router.HandleFunc("/report/{unit}/{year}/{month}", GetReport).Methods("GET")

	fmt.Println(DifferenceInMinutes(TimeOfDay{8,35},TimeOfDay{8,37})) // 2
	fmt.Println(DifferenceInMinutes(TimeOfDay{8,51},TimeOfDay{9,12})) // 21
	fmt.Println(DifferenceInMinutes(TimeOfDay{10,10},TimeOfDay{9,50})) // 20
	
	http.ListenAndServe(":9999", router)
}
