package main

import (
	"encoding/hex"
	"log"
	//"net/http"
	"encoding/json"
	"io/ioutil"
	"fmt"
	"net"
	//"crypto/sha256"
	//"time"
)

const MulticastUDPPort string = "9191"
const UnicastUDPPort string = "9190"
const RESTPort string = "9192"
var LocalIP string

const DefaultPassword string = "1234"

// schema for the prototype: simple PSK, better than nothing!
const SilentEducationPSK string = "4baUV/2T=1a4nGrDS43FGnv6100asRNa35+shd/2b42300aNUFHsdn2m3iUJ86B/d2"

const ServerInfoConfigFile = "./config.json"

type ServerInfo struct {
	AdminPasswordHash	string	`json:"adminpasswordhash"`
	UserPasswordHash	string	`json:"userpasswordhash"`
}

var serverInfo *ServerInfo

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

const (
	srvAddr         = "224.0.0.1"
	maxDatagramSize = 8192
)

func main() {
	LocalIP = GetLocalIP()
	fmt.Println("Silent Education Server started on IPv4: " + LocalIP)
	serveMulticastUDP(srvAddr+":"+MulticastUDPPort, msgHandler)
}

//func CommunicateServerIPToSupplicant(supplicant string) {
//        fmt.Println(" >> CommunicateServerIPToSupplicant called:  " + supplicant)
//	client := &http.Client{}
//	req, err := http.NewRequest("GET", "http://"+supplicant+":"+RESTPort+"/serverip/"+LocalIP, nil)
//	if err != nil {
//		fmt.Println("Error: " + err.Error())
//	} else {
//		req.Header.Set("psk", SilentEducationPSK)	
//		_, _ = client.Do(req)
//	}
//}

func CommunicateServerIPToSupplicant(supplicant string) {
	addr, err := net.ResolveUDPAddr("udp", supplicant+":"+UnicastUDPPort)
	if err != nil {
		log.Fatal(err)
	}
	c, err := net.DialUDP("udp", nil, addr)
	c.Write([]byte("{\"serverip\":\"" + LocalIP + "\"}"))
}


//func ping(a string) {
//	addr, err := net.ResolveUDPAddr("udp", a)
//	if err != nil {
//		log.Fatal(err)
//	}
//	c, err := net.DialUDP("udp", nil, addr)
//	for {
//		c.Write([]byte("hello, world\n"))
//		time.Sleep(1 * time.Second)
//	}
//}



func msgHandler(src *net.UDPAddr, n int, b []byte) {
	log.Println(n, "bytes read from", src)
	log.Println(hex.Dump(b[:n]))
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
		supplicantAddr := string(src.IP.String())
		h(src, n, b)
		fmt.Println("Writing back to UDP socket...")
		CommunicateServerIPToSupplicant(supplicantAddr)
	}
}

