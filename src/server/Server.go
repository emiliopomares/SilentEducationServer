package main

import (
	//"encoding/hex"
	"log"
	//"net/http"
	"encoding/json"
	"io/ioutil"
	"strings"
	"io"
	"strconv"
	"fmt"
	"net"
	"net/http"
	//"crypto/sha256"
	//"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

const MulticastUDPPort string = "9191"
const UnicastUDPPort string = "9190"
const RESTPort string = "8000"
var LocalIP string

type DeviceInfo struct {
        Volume          int     `json:"volume"`
        Threshold       int     `json:"threshold"`
        Duration        int     `json:"duration"`
	Id		string	`json:"id"`
	Name            string  `json:"name"`
        Activation      string  `json:"activation"`
}

type RenameInfo struct {
	Id	string `json:"id"`
	To	string `json:"to"`
}

type RenameCommand struct {
	Rename		RenameInfo `json:"rename"`
}

type DeviceTypeDeclr struct {
	DeviceType	string	`json:"devicetype"`
}

const DefaultPassword string = "1234"

// schema for the prototype: simple PSK, better than nothing!
const SilentEducationPSK string = "4baUV/2T=1a4nGrDS43FGnv6100asRNa35+shd/2b42300aNUFHsdn2m3iUJ86B/d2"

const ServerInfoConfigFile = "./data/config.json"

type ServerInfo struct {
	AdminPasswordHash	string	`json:"adminpasswordhash"`
	UserPasswordHash	string	`json:"userpasswordhash"`
}

const (
        MulticastAddr   = "224.0.0.1"
        maxDatagramSize = 8192
	LoginAccessFile	= "./data/LoginAccessTemplate.html"
	ControlPanelFile = "./data/ControlPanelTemplate.html"
)

var upgrader = websocket.Upgrader{
        ReadBufferSize:  1024,
        WriteBufferSize: 1024,
}

var serverInfo *ServerInfo




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
//    Server discovery                             //
/////////////////////////////////////////////////////

func CommunicateServerIPToSupplicant(supplicant string) {
	addr, err := net.ResolveUDPAddr("udp", makeAddressFromIPandStrPort(supplicant, UnicastUDPPort))
	if err != nil {
		log.Fatal(err)
	}
	c, err := net.DialUDP("udp", nil, addr)
	c.Write([]byte("{\"serverip\":\"" + LocalIP + "\"}"))
}

func msgHandler(src *net.UDPAddr, n int, b []byte) {
	supplicantAddr := string(src.IP.String())
        CommunicateServerIPToSupplicant(supplicantAddr)
}

func makeAddressFromIPandStrPort(ip string, port string) string {
        return ip + ":" + port
}

func makeAddressFromIPandIntPort(ip string, port int) string {
        return ip + ":" + strconv.Itoa(port)
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
		h(src, n, b)
	}
}

func startServiceDiscovery() {
	go serveMulticastUDP(makeAddressFromIPandStrPort(MulticastAddr, MulticastUDPPort), msgHandler)
}


/////////////////////////////////////////////////////
//    WebServer                                    //
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

func ReplyWithInstancedFileTemplate(w http.ResponseWriter, filepath string, keys []string, values []string) {
	rawcontents, _ := ioutil.ReadFile(filepath)
	contents := string(rawcontents)
	for i := 0 ; i < len(keys) ; i++ {
		contents = strings.Replace(contents, keys[i], values[i], 5)
	}
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	io.WriteString(w, contents)
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

func CheckPasswd(user string, pass string) bool {
	return pass == "1234"
}

func AdminUserHandler(w http.ResponseWriter, pass string) {
	if CheckPasswd("Admin", pass) {
		ReplyWithInstancedFileTemplate(w, ControlPanelFile, []string{"<usertype>", "<serverip>", "<isadmin>"}, []string{"Administrador", LocalIP, "true"})
	} else {
		ReplyWithInstancedFileTemplate(w, LoginAccessFile, []string{"<message>", "<serverip>"}, []string{"Datos de acceso incorrectos", LocalIP})
	}
}

func UserUserHandler(w http.ResponseWriter, pass string) {
	if CheckPasswd("Admin", pass) {
                ReplyWithInstancedFileTemplate(w, ControlPanelFile, []string{"<usertype>", "<serverip>", "<isadmin>"}, []string{"Profesor/a", LocalIP, "false"})
        } else {
                ReplyWithInstancedFileTemplate(w, LoginAccessFile, []string{"<message>", "<serverip>"}, []string{"Datos de acceso incorrectos", LocalIP})
        }
}

func NonAuthorizedUserHandler(w http.ResponseWriter, pass string) {
	ReplyWithInstancedFileTemplate(w, LoginAccessFile, []string{"<message>", "<serverip>"}, []string{"Datos de acceso incorrectos", LocalIP})
}

func GetUserHandler(user string) func(http.ResponseWriter, string) {
	if user == "Admin" {
		return AdminUserHandler
	} else if user == "User" {
		return UserUserHandler
	}
	return NonAuthorizedUserHandler
}

func AttemptLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
            fmt.Fprintf(w, "ParseForm() err: %v", err)
            return
        }
	user := r.FormValue("user")
	pass := r.FormValue("passwd")
	UserHandler := GetUserHandler(user)
	UserHandler(w, pass)
}

func LoginScreen(w http.ResponseWriter, r *http.Request) {
	ReplyWithInstancedFileTemplate(w, LoginAccessFile, []string{"<message>", "<serverip>"}, []string{"", LocalIP})
}
//ReplyWithInstancedFileTemplate(w, LoginAccessFile, []string{"<message>", "<serverip>"}, []string{"Datos de acceso incorrectos", LocalIP})

func Healthcheck(w http.ResponseWriter, r *http.Request) {
	JSONResponseFromString(w, "{\"alive\":true}")
}

func setupWebServer() {
        r := mux.NewRouter()
        r.HandleFunc("/", LoginScreen).Methods("GET")
        r.HandleFunc("/login", AttemptLogin).Methods("POST")
	r.HandleFunc("/healthcheck", Healthcheck).Methods("GET")
//        r.HandleFunc("/panel", GetPanel).Methods("PUT")
        http.ListenAndServe(":"+RESTPort, r)
}


/////////////////////////////////////////////////////
//    Websockets                                   //
/////////////////////////////////////////////////////

var devices	       []*DeviceInfo
var deviceConnections  []*websocket.Conn
var webConnections     []*websocket.Conn

func wsBroadcastToBrowsers(cmd string) {
	fmt.Println("  >> wsBroadcastToBrowsers called " + cmd + " and webconnections.length = " + strconv.Itoa(len(webConnections)))
	bytes := []byte(cmd)
	for i:=0 ; i < len(webConnections); i++ {
		_ = webConnections[i].WriteMessage(1, bytes)
	}
}

func wsSendDevicesListToBrowser(socket *websocket.Conn) {
	for i:=0 ; i < len(devices); i++ {
		bytes, _ := json.Marshal(devices[i])
		_ = socket.WriteMessage(1, bytes)
	}
}

func wsBroadcastDeviceInfoToBrowsers(info *DeviceInfo) {
	bytes, _ := json.Marshal(info)
	fmt.Println("  >> about to broadcast these bytes to browsers: ", string(bytes))
	for i:=0 ; i < len(webConnections); i++ {
                _ = webConnections[i].WriteMessage(1, bytes)
        }
}

func findInDeviceArray(id string) (int, bool) {
	for i:=0; i < len(devices); i++ {
		if(devices[i].Id == id) {
			return i,true
		}
	}
	return -1,false
}

func addToDevicesIfNew(info *DeviceInfo) {
	_, found := findInDeviceArray(info.Id)
	if !found {
		fmt.Println(" Adding a new device to list: ", info)
		devices = append(devices, info)
		wsBroadcastDeviceInfoToBrowsers(info)	
	}
}

func wsProcessDeviceCommand(socket *websocket.Conn, cmd string) {
	fmt.Println("  >> wsProcessDeviceCommand called " + cmd)
	var info DeviceInfo
	err := json.Unmarshal([]byte(cmd), &info)
	if err != nil {
		fmt.Println("Error receiving command from device: " + err.Error())
	} else {
		fmt.Println(" info unmarshalled as : " , info)
		addToDevicesIfNew(&info)
		//wsBroadcastToBrowsers(cmd)
	}	
}

func wsProcessBrowserCommand(socket *websocket.Conn, cmd string) {
	var command RenameCommand
	err := json.Unmarshal([]byte(cmd), &command)
	if err == nil {
		fmt.Println("  >> rename command")
		index, exists := findInDeviceArray(command.Rename.Id)
		if exists {
			deviceConnections[index].WriteMessage(1, []byte(cmd))			
		}
	} else {
		fmt.Println("  >> was not a rename command: " , cmd)
	}
}

func wsFunction(w http.ResponseWriter, r *http.Request) {
		fmt.Println("wsFunction called")
                upgrader.CheckOrigin = func(r *http.Request) bool { return true }
                conn, err := upgrader.Upgrade(w, r, nil) // error ignored for sake of simplicity
                if err != nil {
                        fmt.Println("This was the error: ", err)
                }

		msgtype, handshakemsg, err := conn.ReadMessage()
		if err != nil {
			return
		}

		if(string(handshakemsg) == "{\"devicetype\":\"device\"}") {
			fmt.Println("A new device registered ", msgtype)
			deviceConnections = append(deviceConnections, conn)
			for {
                        	// Read message from device
                        	_, msg, err := conn.ReadMessage()
                        	if err != nil {
                                	return
                        	}
                        	wsProcessDeviceCommand(conn, string(msg))
                	}
		} 

		if(string(handshakemsg) == "{\"devicetype\":\"web\"}") {
			wsSendDevicesListToBrowser(conn)
			fmt.Println("A new web browser registered ", msgtype)
			webConnections = append(webConnections, conn)
			for {
                        	// Read message from browser
                        	_, msg, err := conn.ReadMessage()
                        	if err != nil {
                                	return
                        	}
                        	wsProcessBrowserCommand(conn, string(msg))
                	}
		}

}

func setupWebsocket() {
	http.HandleFunc("/", wsFunction)
	http.ListenAndServe(":8081", nil)
}


/////////////////////////////////////////////////////
//    Main function                                //
/////////////////////////////////////////////////////

func main() {
        LocalIP = GetLocalIP()
        fmt.Println("Silent Education Server started on IPv4: " + LocalIP)
	fmt.Println("Panel de control ->  http://127.0.0.1:8080/")
	startServiceDiscovery()
	go setupWebServer()
	setupWebsocket()
}
