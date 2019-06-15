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

var serverInfo *ServerInfo

const ServerInfoConfigFile = "./config.json"

const (
	srvAddr         = "224.0.0.1"
	maxDatagramSize = 8192
)

var serverIPknown bool

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

func ReceiveServerIP(w http.ResponseWriter, r *http.Request) {
	fmt.Println(" >> ReceiveServerIP called")
	vars := mux.Vars(r)
	serverInfo.ServerIP = vars["ip"]
	fmt.Println("Server IP set to " + serverInfo.ServerIP)
	SaveServerConfigToFile()
	JSONResponseFromString(w, "{\"result\":\"OK\"}")	 
}

func main() {

	serverIPknown = false

	LoadServerConfigFromFile()

	go serveUnicastUDP(srvAddr+":"+UnicastUDPPort, msgHandler)

	go ping(srvAddr+":"+MulticastUDPPort)

	for {
		time.Sleep(10 * time.Second)
	}
}

func ping(a string) {
	addr, err := net.ResolveUDPAddr("udp", a)
	rcvBuffer := make([]byte, 1024)
	if err != nil {
		fmt.Println("Error")
		log.Fatal(err)
	}
	c, err := net.DialUDP("udp", nil, addr)
	for {
		fmt.Println("writing to udp socket...")
		c.Write([]byte("hello, world\n"))
		fmt.Println("Waiting for data from socket...")
		bytes, err := c.Read(rcvBuffer)
		if err == nil {
			fmt.Println(strconv.Itoa(bytes) + " read: " + string(rcvBuffer))
		} else {
			fmt.Println("There was this err: " + err.Error())
		}
		time.Sleep(5 * time.Second)
	}
}

func msgHandler(src *net.UDPAddr, n int, b []byte) {
	//log.Println(n, "bytes read from", src)
	//log.Println(hex.Dump(b[:n]))
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
		serverIPknown = true
		fmt.Println("server ip set to " , src)
		h(src, n, b)
	}
}

