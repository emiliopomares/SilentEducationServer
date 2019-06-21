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
	"math/rand"
	"sync"
	"github.com/gorilla/mux"
	"github.com/sacOO7/gowebsocket"
	"log"
	"net/http"
	"io"
	"strings"
	"io/ioutil"
	"net"
	"strconv"
	"time"
)

const MulticastUDPPort string = "9191"
const UnicastUDPPort string = "9190"
const AppAudioPort = "9192"
const ServerAudioPort = "9193"
const RESTPort string = "8000"

// schema for the prototype: simple PSK, better than nothing!
const SilentEducationPSK = "4baUV/2T=1a4nGrDS43FGnv6100asRNa35+shd/2b42300aNUFHsdn2m3iUJ86B/d2"

type ServerInfo struct {
	ServerIP	string	`json:"serverip"`
}

type PingInfo struct {
	Ping		int	`json:"ping"`
}

type DeviceInfo struct {
	Volume		int	`json:"volume"`
	Threshold	int	`json:"threshold"`
	Duration	int	`json:"duration"`
	Id		string  `json:"id"`
	Name		string  `json:"name"`
	Activation	string	`json:"activation"`
	PairPIN		string  `json:"pairpin"`
	PairSecret	string	`json:"pairsecret"`
	PairedDevices 	int	`json:"paireddevices"`
}

var activationStatus int

var serverInfo *ServerInfo
var deviceInfo DeviceInfo

const ServerInfoConfigFile = "./serverconfig.json"
const DeviceConfigFile = "./deviceconfig.json"

const (
	MulticastAddr   = "224.0.0.1"
	maxDatagramSize = 8192
)

var serverIPknown bool



/////////////////////////////////////////////////////
//    Commands                                     //
/////////////////////////////////////////////////////

type RenameInfo struct {
        Id      string `json:"id"`
        To      string `json:"to"`
}

type RenameCommand struct {
        Rename          RenameInfo `json:"rename"`
}



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

func SaveDeviceConfigToFile() {
    marshaledData, _ := json.Marshal(deviceInfo)
    _ = ioutil.WriteFile(DeviceConfigFile, marshaledData, 0644)
    fmt.Printf("Device config info file written")
}

func MakeDefaultConfigFile() DeviceInfo {
	var newdevinfo DeviceInfo
	rname := "SE-" + RandStringRunes(6)
	newdevinfo.Volume = 75
	newdevinfo.Threshold = 50
	newdevinfo.Duration = 20
	newdevinfo.Name = rname
	newdevinfo.Activation = "green"
	newdevinfo.Id = rname
	return newdevinfo
}

func LoadDeviceConfigFromFile() {
	file, err := ioutil.ReadFile(DeviceConfigFile)
        var config DeviceInfo
        if err == nil {
                _ = json.Unmarshal(file, &config)
        	deviceInfo = config
	} else {
                fmt.Println("No device config info file found")
		deviceInfo = MakeDefaultConfigFile()
		SaveDeviceConfigToFile()
        }
}

/////////////////////////////////////////////////////
//    Temp REST API                                //
/////////////////////////////////////////////////////

func WithPSKCheck(next http.HandlerFunc) http.HandlerFunc {
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
        bytes, _ := json.Marshal(deviceInfo)
	JSONResponseFromString(w, string(bytes))
}

func UpdateStatus(w http.ResponseWriter, r *http.Request) {
        var newInfo DeviceInfo
	json.NewDecoder(r.Body).Decode(&newInfo)	
	deviceInfo = newInfo
	SaveDeviceConfigToFile()
        JSONResponseFromString(w, "{\"result\":\"success\"}")
}

func Healthcheck(w http.ResponseWriter, r *http.Request) {
	fmt.Println("GET /healthcheck")
	JSONResponseFromString(w, "{\"alive\":true}")
}

func HandlePing(w http.ResponseWriter, r *http.Request) {
        vars := mux.Vars(r)
	if(vars["ip"] != "") {
		fmt.Println("Requesting ping to: " + vars["ip"])
		delay := Ping(vars["ip"])
        	JSONResponseFromString(w, "{\"result\":"+strconv.Itoa(delay)+"}")
	} else {
		JSONResponseFromStringAndCode(w, "{\"error\":\"destination ip not specified\"}", 400)
	}
}

func setupRESTAPI() {
	r := mux.NewRouter()
	r.HandleFunc("/serverip", WithPSKCheck(GetServerIP)).Methods("GET")
	r.HandleFunc("/status", WithPSKCheck(GetStatus)).Methods("GET")
	r.HandleFunc("/status", WithPSKCheck(UpdateStatus)).Methods("PUT")
	r.HandleFunc("/pairing", WithPSKCheck(CreatePairing)).Methods("POST")
	r.HandleFunc("/pairing", WithPSKCheck(DeletePairing)).Methods("DELETE")
	r.HandleFunc("/pairing", CheckPairing).Methods("GET")
	r.HandleFunc("/healthcheck", Healthcheck).Methods("GET")
	r.HandleFunc("/ping/{ip}", WithPSKCheck(HandlePing)).Methods("GET")
	http.ListenAndServe(":"+RESTPort, r)
}



/////////////////////////////////////////////////////
//    REST API                                     //
/////////////////////////////////////////////////////

func CreatePairing(w http.ResponseWriter, r *http.Request) {
	deviceInfo.PairedDevices++
	SaveDeviceConfigToFile()
	fmt.Println("   >> Pairing created. Paired to " + strconv.Itoa(deviceInfo.PairedDevices) + " devices now")
	JSONResponseFromString(w, "{\"secret\":\""+deviceInfo.PairSecret+"\"}")
}

func DeletePairing(w http.ResponseWriter, r *http.Request) {
	if(deviceInfo.PairedDevices > 0) {
		deviceInfo.PairedDevices--
	}
	SaveDeviceConfigToFile()
	JSONResponseFromString(w, "{\"result\":\"sucess\"}")
}

func CheckPairing(w http.ResponseWriter, r *http.Request) {
	fmt.Println("   >> CheckPairing called with psk: " + r.Header.Get("psk"))
	Secret := r.URL.Query()["s"][0]
	if(Secret == deviceInfo.PairSecret) {
		JSONResponseFromString(w, "{\"result\":true}")
	} else {
		JSONResponseFromString(w, "{\"result\":false}")
	}
}


/////////////////////////////////////////////////////
//    Websockets                                   //
/////////////////////////////////////////////////////

var socket gowebsocket.Socket 

func wsProcessMessage(socket gowebsocket.Socket, command string) {
	var rencomm RenameCommand
	err := json.Unmarshal([]byte(command), &rencomm)
	if err == nil {
		deviceInfo.Name = rencomm.Rename.To
		fmt.Println("    >>> Setting device name to " + deviceInfo.Name)
		SaveDeviceConfigToFile()
	}
}

func ConnectWS(ip string, port string) {
	socket = gowebsocket.New("ws://" + ip + ":" + port + "/")
	
	socket.OnConnectError = func(err error, socket gowebsocket.Socket) {
		log.Fatal("Received connect error - ", err)
	}
  
	socket.OnConnected = func(socket gowebsocket.Socket) {
		log.Println("Connected to server");
	}
  
	socket.OnTextMessage = func(message string, socket gowebsocket.Socket) {
		//log.Println("Received message - " + message)
		wsProcessMessage(socket, message)
	}
  
	socket.OnPingReceived = func(data string, socket gowebsocket.Socket) {
		log.Println("Received ping - " + data)
	}
  
    	socket.OnPongReceived = func(data string, socket gowebsocket.Socket) {
		log.Println("Received pong - " + data)
	}

	socket.OnDisconnected = func(err error, socket gowebsocket.Socket) {
		log.Println("Disconnected from server ")
		return
	}
  
	socket.Connect()
}

func SendTextWS(msg string) {
	socket.SendText(msg)
}

/////////////////////////////////////////////////////
//    Status                                       //
/////////////////////////////////////////////////////

func ProcessPairRequest() {
	var PINtoShow string
	if(deviceInfo.PairedDevices == 0) {
		newPIN := GeneratePIN()
		newSecret := GeneratePairSecret()
		deviceInfo.PairPIN = newPIN
		deviceInfo.PairSecret = newSecret
		SaveDeviceConfigToFile()
		PINtoShow = newPIN
	} else {
		PINtoShow = deviceInfo.PairPIN
	} 
	fmt.Println("   >> Pair request received. PIN: " + PINtoShow)
	
}

func ProcessRespondRequest(Secret string, supplicantAddr string) {
	if(Secret != deviceInfo.PairSecret) {
		return
	}
	addr, err := net.ResolveUDPAddr("udp", makeAddressFromIPandStrPort(supplicantAddr, UnicastUDPPort))
        if err != nil {
                log.Fatal(err)
        }
        c, err := net.DialUDP("udp", nil, addr)
	LocalIP := GetLocalIP()
        c.Write([]byte("{\"serverip\":\"" + LocalIP + "\"}"))
}

func ProcessWhoseSecret(Secret string, supplicantAddr string) {
        if(Secret == deviceInfo.PairSecret) {
                addr, err := net.ResolveUDPAddr("udp", makeAddressFromIPandStrPort(supplicantAddr, UnicastUDPPort))
                if err != nil {
                        fmt.Println("Error")
                        log.Fatal(err)
                }
                c, err := net.DialUDP("udp", nil, addr)
                if err != nil {
                        fmt.Println("Error")
                        log.Fatal(err)
                }
                LocalIP := GetLocalIP()
                c.Write([]byte("{\"deviceip\":\"" + LocalIP + "\"}"))
        }
}

func ProcessWhoisRequest(PIN string, supplicantAddr string) {
	if(PIN == deviceInfo.PairPIN) {
		addr, err := net.ResolveUDPAddr("udp", makeAddressFromIPandStrPort(supplicantAddr, UnicastUDPPort))
		if err != nil {
                	fmt.Println("Error")
                	log.Fatal(err)
        	}
        	c, err := net.DialUDP("udp", nil, addr)
		if err != nil {
			fmt.Println("Error")
                        log.Fatal(err)
		}
		LocalIP := GetLocalIP()
        	c.Write([]byte("{\"deviceip\":\"" + LocalIP + "\"}"))
	}
}

/////////////////////////////////////////////////////
//    Network utils                                //
/////////////////////////////////////////////////////

type PingAvailable struct {
        sync.Mutex
        cond *sync.Cond
}

func CreatePingAvailable() *PingAvailable {
        p := PingAvailable{}
        p.cond = sync.NewCond(&p)
        return &p
}

var ping *PingAvailable

func Ping(dest string) int {
	fmt.Println("    >> Ping "+ dest + "called")
	return PingN(1, dest)
}

func ProcessPing(n int, srcAddr string) {
	fmt.Println("    >> ProcessPing " + strconv.Itoa(n) + " " + srcAddr + " called")
        if n == 0 {
                ping.Lock()
                ping.cond.Signal()
                ping.Unlock()
        } else {
                SendPing(n-1, srcAddr)
        }
}

func PingN(n int, addr string) int {
	fmt.Println("    >> PingN " + strconv.Itoa(n) + " " + addr + " called")
        start := time.Now()
        SendPing(1, addr)
        func(p *PingAvailable) {
                p.Lock()
		fmt.Println("     >>> ... starting wait for ping response ")
                p.cond.Wait()
		fmt.Println("     >>> ... ping response received! ")
                p.Unlock()
                return
        }(ping)
        t := time.Now()
        elapsed := t.Sub(start)
        return int(elapsed/1000000)
}

func SendPing(n int, destAddr string) {
	fmt.Println("   >> SendPing " + strconv.Itoa(n) + " " + destAddr + " called")
        addr, err := net.ResolveUDPAddr("udp", makeAddressFromIPandStrPort(destAddr, UnicastUDPPort))
        if err != nil {
                fmt.Println("Error")
                log.Fatal(err)
        }
        c, err := net.DialUDP("udp", nil, addr)
        fmt.Println("   >>>  .... actually Sending ping...")
        c.Write([]byte("{\"ping\":"+strconv.Itoa(n)+"}"))
}

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
		time.Sleep(15 * time.Second)
		if serverIPknown == false {
			fmt.Println("Echo timeout, retrying...")
		}
	}
}

func SendHandshakeToServer() {
	ConnectWS(serverInfo.ServerIP, "8081")
	SendTextWS("{\"devicetype\":\"device\"}")
	bytes, _ := json.Marshal(deviceInfo)
	SendTextWS(string(bytes))
}

func unicastMsgHandler(src *net.UDPAddr, n int, b []byte) {
	fmt.Println("       >>>>>   unicast handler received: " + string(b[:n]))
	var NewServerInfo ServerInfo
	err := json.Unmarshal(b[:n], &NewServerInfo)
	success := false
	if (err == nil) && (NewServerInfo.ServerIP != "") {
		serverIP := NewServerInfo.ServerIP
                serverIPknown = true
                serverInfo.ServerIP = serverIP
                fmt.Println("Server IP is set to " + serverIP)
                SendHandshakeToServer()
		success = true
	} 
	fmt.Println("    >>>> ... was not server info")	
	var NewPingInfo PingInfo
	err = json.Unmarshal(b[:n], &NewPingInfo)
	if err == nil {
		fmt.Println("      >>>>     unicastMsgHandler: received ping!")
		ProcessPing(NewPingInfo.Ping, src.IP.String())
		success = true
	}

	if success == false {
		fmt.Println("   >>> something was received, but no match")
	}

}

func multicastMsgHandler(src *net.UDPAddr, n int, b []byte) {
	msg := string(b)
//	fmt.Println("Some idiot broadcast the message " + msg)
	supplicantAddr := string(src.IP.String())
	if(strings.HasPrefix(msg, "PairRequest")) {
		fmt.Println("  >> Processing PairRequest....")
		ProcessPairRequest()
	}	
	if(strings.HasPrefix(msg, "Respond")) {
		Secret := string(b[7:31])
		ProcessRespondRequest(Secret, supplicantAddr)
	}
	if(strings.HasPrefix(msg, "Whois")) {
		PIN := string(b[5:9])
		fmt.Println("    >>>> received Whois " + PIN)
		ProcessWhoisRequest(PIN, supplicantAddr)
	}
}

func makeAddressFromIPandStrPort(ip string, port string) string {
	return ip + ":" + port
}

func makeAddressFromIPandIntPort(ip string, port int) string {
	return ip + ":" + strconv.Itoa(port)
}

func listenMulticastUDP() {
	serveMulticastUDP(makeAddressFromIPandStrPort(MulticastAddr, MulticastUDPPort), multicastMsgHandler)
}

func listenUnicastUDP() {
	serveUnicastUDP(makeAddressFromIPandStrPort(MulticastAddr, UnicastUDPPort), unicastMsgHandler)
}

func serveMulticastUDP(a string, h func(*net.UDPAddr, int, []byte)) {
        addr, err := net.ResolveUDPAddr("udp", a)
        if err != nil {
                log.Fatal(err)
        }
        l, err := net.ListenMulticastUDP("udp", nil, addr)
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

func GetLocalIP() string {
        var localIP string
        addr, err := net.InterfaceAddrs()
        if err != nil {
                fmt.Printf("GetLocalIP in communication failed")
                return "localhost"
        }
        for _, val := range addr {
                if ip, ok := val.(*net.IPNet); ok && !ip.IP.IsLoopback() {
                        if ip.IP.To4() != nil {
                                localIP = ip.IP.String()
                        }
                }
        }
        return localIP
}


/////////////////////////////////////////////////////
//    Initializations                              //
/////////////////////////////////////////////////////

func GeneratePIN() string {
	return RandNumericString(4)
}

func GeneratePairSecret() string {
	return RandStringRunes(24)
}

/////////////////////////////////////////////////////
//    Initializations                              //
/////////////////////////////////////////////////////

func initializeDevice() {
	initializeRandom()
	serverInfo = &ServerInfo{}
	serverIPknown = false
	LoadDeviceConfigFromFile()
	ping = CreatePingAvailable()
	StartAudio()
	go listenUnicastUDP()
	go listenMulticastUDP()
	go discoverServerIP()
}

func initializeRandom() {
    rand.Seed(time.Now().UnixNano())
}


/////////////////////////////////////////////////////
//    Audio                                        //
/////////////////////////////////////////////////////

func StartAudio() {

}






/////////////////////////////////////////////////////
//    Randomization                                //
/////////////////////////////////////////////////////

var digits = []byte("0123456789")

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
    b := make([]rune, n)
    for i := range b {
        b[i] = letterRunes[rand.Intn(len(letterRunes))]
    }
    return strings.ToUpper(string(b))
}

func RandNumericString(n int) string {
    b := make([]byte, n)
    for i := range b {
        b[i] = digits[rand.Intn(len(digits))]
    }
    return strings.ToUpper(string(b))
} 


/////////////////////////////////////////////////////
//    Main function				   //
/////////////////////////////////////////////////////

func main() {
	initializeDevice()
	fmt.Println("Device initialized @ " + GetLocalIP())
	if(deviceInfo.PairPIN != "") {
		fmt.Println("Pair PIN for this device: " + deviceInfo.PairPIN)
	}
	setupRESTAPI()
}
