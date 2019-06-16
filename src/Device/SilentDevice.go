//////////////////////////////////////////////
// SilentEducation, S.L.
//
// Copyright 2019 by Emilio Pomares
//////////////////////////////////////////////

package main

import (
	//"encoding/hex"
	"fmt"
	"encoding/json"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"io"
	"io/ioutil"
	"net"
	"strconv"
	"time"
)

const MulticastUDPPort string = "9191"
const UnicastUDPPort string = "9190"
const RESTPort string = "9192"

// schema for the prototype: simple PSK, better than nothing!
const SilentEducationPSK = "4baUV/2T=1a4nGrDS43FGnv6100asRNa35+shd/2b42300aNUFHsdn2m3iUJ86B/d2"

type ServerInfo struct {
	ServerIP	string	`json:"serverip"`
}

type DeviceStatus struct {
	Volume		int	`json:"volume"`
	Threshold	int	`json:"threshold"`
	Duration	int	`json:"duration"`
	Status		int	`json:"status"`
}

var serverInfo *ServerInfo
var deviceStatus DeviceStatus

const ServerInfoConfigFile = "./config.json"

const (
	MulticastAddr   = "224.0.0.1"
	maxDatagramSize = 8192
)

var serverIPknown bool



/////////////////////////////////////////////////////
//    Configuration                                //
/////////////////////////////////////////////////////

func SaveServerConfigToFile() {
    marshaledData, _ := json.Marshal(serverInfo)
    _ = ioutil.WriteFile(ServerInfoConfigFile, marshaledData, 0644)
    fmt.Printf("Server config info file written")
}

func LoadServerConfigFromFile() {
	file, err := ioutil.ReadFile(ServerInfoConfigFile)
	config := &ServerInfo{}
	if err == nil {
		_ = json.Unmarshal(file, config)
	} else {
		fmt.Println("No server config info file found")
	}
	serverInfo = config
}



/////////////////////////////////////////////////////
//    Temp REST API                                //
/////////////////////////////////////////////////////

func withPSKCheck(next http.HandlerFunc) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
                psk := r.Header.Get("psk")
                if psk == SilentEducationPSK {
                        next.ServeHTTP(w, r)
                } else {
                        fmt.Println("forbidden")
                        JSONResponseFromStringAndCode(w, "{\"result\":\"forbidden\"}", 403)
                }
        }
}

func JSONResponseFromString(w http.ResponseWriter, res string) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
        w.WriteHeader(http.StatusOK)
        io.WriteString(w, res)
}

func JSONResponseFromStringAndCode(w http.ResponseWriter, res string, status int) {
        w.Header().Set("Content-Type", "application/json; charset=UTF-8")
        w.WriteHeader(status)
        io.WriteString(w, res)
}

func GetServerIP(w http.ResponseWriter, r *http.Request) {
        if(serverIPknown) {
		JSONResponseFromString(w, "{\"result\":\""+serverInfo.ServerIP+"\"}")
	} else {
		JSONResponseFromString(w, "{\"result\":\"pending\"}")
	}
}

func GetStatus(w http.ResponseWriter, r *http.Request) {
        bytes, _ := json.Marshal(deviceStatus)
	JSONResponseFromString(w, string(bytes))
}

func UpdateStatus(w http.ResponseWriter, r *http.Request) {
        var newStatus DeviceStatus
	json.NewDecoder(r.Body).Decode(&newStatus)	
	deviceStatus = newStatus
        JSONResponseFromString(w, "{\"result\":\"success\"}")
}

func setupRESTAPI() {
	r := mux.NewRouter()
	r.HandleFunc("/serverip", GetServerIP).Methods("GET")
	r.HandleFunc("/status", GetStatus).Methods("GET")
	r.HandleFunc("/status", UpdateStatus).Methods("PUT")
	http.ListenAndServe(":8080", r)
}



/////////////////////////////////////////////////////
//    Websockets                                   //
/////////////////////////////////////////////////////



/////////////////////////////////////////////////////
//    Server discovery                             //
/////////////////////////////////////////////////////

func discoverServerIP() {
	addr, err := net.ResolveUDPAddr("udp", makeAddressFromIPandStrPort(MulticastAddr, MulticastUDPPort))
	if err != nil {
		fmt.Println("Error")
		log.Fatal(err)
	}
	c, err := net.DialUDP("udp", nil, addr)
	for serverIPknown == false {
		fmt.Println("Sending service discovery packet...")
		c.Write([]byte("SERequestServerIP"))
		fmt.Println("Waiting for echo...")
		time.Sleep(5 * time.Second)
		if serverIPknown == false {
			fmt.Println("Echo timeout, retrying...")
		}
	}
}

func msgHandler(src *net.UDPAddr, n int, b []byte) {
	var NewServerInfo ServerInfo
	err := json.Unmarshal(b[:n], &NewServerInfo)
	if err != nil {
		fmt.Println("Error parsing response from server: " + err.Error())
	} else {
		serverIP := NewServerInfo.ServerIP
		serverIPknown = true
		serverInfo.ServerIP = serverIP
		fmt.Println("Server IP is set to " + serverIP)
	}
}

func makeAddressFromIPandStrPort(ip string, port string) string {
	return ip + ":" + port
}

func makeAddressFromIPandIntPort(ip string, port int) string {
	return ip + ":" + strconv.Itoa(port)
}

func listenUDP() {
	serveUnicastUDP(makeAddressFromIPandStrPort(MulticastAddr, UnicastUDPPort), msgHandler)
}

func serveUnicastUDP(a string, h func(*net.UDPAddr, int, []byte)) {
	addr, err := net.ResolveUDPAddr("udp", a)
	if err != nil {
		log.Fatal(err)
	}
	l, err := net.ListenUDP("udp", addr)
	l.SetReadBuffer(maxDatagramSize)
	for {
		b := make([]byte, maxDatagramSize)
		n, src, err := l.ReadFromUDP(b)
		if err != nil {
			log.Fatal("ReadFromUDP failed:", err)
		}
		h(src, n, b)
	}
}

func initializeDevice() {
	serverInfo = &ServerInfo{}
	serverIPknown = false
	go listenUDP()
	go discoverServerIP()
}




/////////////////////////////////////////////////////
//    Main function				   //
/////////////////////////////////////////////////////

func main() {
	initializeDevice()
	setupRESTAPI()
}
