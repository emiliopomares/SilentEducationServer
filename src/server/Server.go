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
	"unsafe"
	//"os"
	"net/http"

	//"crypto/sha256"
	//"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

const MulticastUDPPort string = "9191"
const UnicastUDPPort string = "9190"
const RESTPort string = "8080"
const AudioWSPort = "9196"
const TargetSampleRate = 44100
var LocalIP string

type DeviceInfo struct {
        Volume          int     `json:"volume"`
        Threshold       int     `json:"threshold"`
        Duration        int     `json:"duration"`
		Id				string	`json:"id"`
		Name            string  `json:"name"`
        Activation      string  `json:"activation"`
}

/////////////////////////////////////////////////////
//    Commands                                     //
/////////////////////////////////////////////////////

type RenameInfo struct {
	Id	string `json:"id"`
	To	string `json:"to"`
}

type RenameCommand struct {
	Rename		RenameInfo `json:"rename"`
}

type MessageInfo struct {
        To      int    `json:"to"`
        Msg     string `json:"msg"`
	Color	string `json:"color"`
}

type MessageCommand struct {
        Message          MessageInfo `json:"message"`
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
	fmt.Println("  >> started service discovery")
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
		ReplyWithInstancedFileTemplate(w, ControlPanelFile, []string{"<usertype>", "<serverip>", "<port>", "<isadmin>"}, []string{"Administrador", "localhost", RESTPort, "true"})
	} else {
		ReplyWithInstancedFileTemplate(w, LoginAccessFile, []string{"<message>", "<serverip>", "<port>"}, []string{"Datos de acceso incorrectos", "localhost", RESTPort})
	}
}

func UserUserHandler(w http.ResponseWriter, pass string) {
	if CheckPasswd("Admin", pass) {
                ReplyWithInstancedFileTemplate(w, ControlPanelFile, []string{"<usertype>", "<serverip>", "<port>", "<isadmin>"}, []string{"Profesor/a", "localhost", RESTPort, "false"})
        } else {
                ReplyWithInstancedFileTemplate(w, LoginAccessFile, []string{"<message>", "<serverip>", "<port>"}, []string{"Datos de acceso incorrectos", "localhost", RESTPort})
        }
}

func NonAuthorizedUserHandler(w http.ResponseWriter, pass string) {
	ReplyWithInstancedFileTemplate(w, LoginAccessFile, []string{"<message>", "<serverip>", "<port>"}, []string{"Datos de acceso incorrectos", "localhost", RESTPort})
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
	ReplyWithInstancedFileTemplate(w, LoginAccessFile, []string{"<message>", "<serverip>", "<port>"}, []string{"", "localhost", RESTPort})
}
//ReplyWithInstancedFileTemplate(w, LoginAccessFile, []string{"<message>", "<serverip>", "<port>"}, []string{"Datos de acceso incorrectos", LocalIP, RESTPort})

func Healthcheck(w http.ResponseWriter, r *http.Request) {
	JSONResponseFromString(w, "{\"alive\":true}")
}

func setupWebServer() {
	fmt.Println("  >> web server setup")
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

var devices	       	   []*DeviceInfo
var deviceConnections  []*websocket.Conn
var deviceAudioConnections  []*websocket.Conn
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
	var message MessageCommand
	fmt.Println("    >>> received command from browser: " + cmd)
	err := json.Unmarshal([]byte(cmd), &command)
	if (err == nil) && (command.Rename.Id != "") {
		fmt.Println("         >> rename command")
		index, exists := findInDeviceArray(command.Rename.Id)
		if exists {
			deviceConnections[index].WriteMessage(1, []byte(cmd))			
		}
	} 

	err = json.Unmarshal([]byte(cmd), &message)
	if (err == nil) && (message.Message.Msg != "") {
		fmt.Println("          >> Message command")
		deviceConnections[message.Message.To].WriteMessage(1, []byte(cmd))
	}
}

// this handler sends audio to the device websocket
func wsAudioToDevice(w http.ResponseWriter, r *http.Request) {
	fmt.Println("                             >>>>>>>>>>>>>>>>>>>>>>>>>>>>>> wsAudioToDevice called <<<<<<<<<<<<<<<<<<<<<")
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
    conn, err := upgrader.Upgrade(w, r, nil) // error ignored for sake of simplicity
        if err != nil {
        fmt.Println("This was the error: ", err)
    }

	deviceAudioConnections = append(deviceAudioConnections, conn)

	fmt.Println("   >> device connected its audio websocket")

	_, _, err = conn.ReadMessage()
		if err != nil {
			return
		}
    
}

var DevicesToSendAudioTo  []int

// this handler takes audio coming from the web interface
func wsAudioFromWeb(w http.ResponseWriter, r *http.Request) {
	fmt.Println("  >> wsAudioFromWeb called")
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
    conn, err := upgrader.Upgrade(w, r, nil) // error ignored for sake of simplicity
        if err != nil {
        fmt.Println("This was the error: ", err)
    }

    for {
		_, handshakemsg, err := conn.ReadMessage()
		if err != nil {
			return
		}

		// check if it is a command
		if(len(handshakemsg) < 128) {
			msg := string(handshakemsg)
			if(strings.HasPrefix(msg, "start")) {
				fmt.Println("   >> audio recording start")
				idStrs := strings.Split(msg[6:],":")
				DevicesToSendAudioTo = []int{}
				Devices := 0
				for i:=0 ; i<len(idStrs) ; i++ {
					fmt.Println("    ### trying to parse " + idStrs[i])
					val, err := strconv.Atoi(idStrs[i])
					if err == nil {
						DevicesToSendAudioTo = append(DevicesToSendAudioTo, val)
						fmt.Println("    >>>>> sending audio to device id: ", val)
						Devices++
					}
				}
				fmt.Println("    >>>>> sending audio to " + strconv.Itoa(Devices) + " devices")
				audioStartRecording()
			} else if (msg == "end") {
				fmt.Println("   >> audio recording end")
				audioEndRecording()
			}

			} else {
				// Assume it's data
				audioStreamData(handshakemsg)
			}
	
		//fmt.Println("Something this long received on the audio endpoint: " + strconv.Itoa(len(handshakemsg)))
	}
}

// maybe split this in two: device and web version, just to keep things more tidy
func wsFunction(w http.ResponseWriter, r *http.Request) {
		
        upgrader.CheckOrigin = func(r *http.Request) bool { return true }
        conn, err := upgrader.Upgrade(w, r, nil) // error ignored for sake of simplicity
        if err != nil {
            fmt.Println("This was the error: ", err)
        }

		msgtype, handshakemsg, err := conn.ReadMessage()
		if err != nil {
			return
		}

		data := string(handshakemsg)
		fmt.Println("wsFunction called: " + data)
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
	fmt.Println("  >> setupWebsocket")
	http.HandleFunc("/", wsFunction) // maybe split this in two: device and web version, just to keep things more tidy
	http.HandleFunc("/audioFromWeb/", wsAudioFromWeb)
	http.HandleFunc("/audioToDevice/", wsAudioToDevice)

	http.ListenAndServe(":8081", nil)
}


/////////////////////////////////////////////////////
//    Audio                                        //
/////////////////////////////////////////////////////

var IncomingSampleRate = 44100

func byteSliceAsFloat32Slice(arr []byte) []float32 {
        lf := len(arr) / 4

        // step by step
        pf := &(arr[0])                        // To pointer to the first byte of b
        up := unsafe.Pointer(pf)                  // To *special* unsafe.Pointer, it can be converted to any pointer
        pi := (*[1]float32)(up)                      // To pointer as byte array
        buf := (*pi)[:]                           // Creates slice to our array of 1 byte
        address := unsafe.Pointer(&buf)           // Capture the address to the slice structure
        lenAddr := uintptr(address) + uintptr(8)  // Capture the address where the length and cap size is stored
        capAddr := uintptr(address) + uintptr(16) // WARNING: This is fragile, depending on a go-internal structure.
        lenPtr := (*int)(unsafe.Pointer(lenAddr)) // Create pointers to the length and cap size
        capPtr := (*int)(unsafe.Pointer(capAddr)) //
        *lenPtr = lf                              // Assign the actual slice size and cap
        *capPtr = lf                              //

        return buf
}

func int16SliceAsByteSlice(arr []int16) []byte {
        lf := 2 * len(arr)

        // step by step
        pf := &(arr[0])                        // To pointer to the first byte of b
        up := unsafe.Pointer(pf)                  // To *special* unsafe.Pointer, it can be converted to any pointer
        pi := (*[1]byte)(up)                      // To pointer as byte array
        buf := (*pi)[:]                           // Creates slice to our array of 1 byte
        address := unsafe.Pointer(&buf)           // Capture the address to the slice structure
        lenAddr := uintptr(address) + uintptr(8)  // Capture the address where the length and cap size is stored
        capAddr := uintptr(address) + uintptr(16) // WARNING: This is fragile, depending on a go-internal structure.
        lenPtr := (*int)(unsafe.Pointer(lenAddr)) // Create pointers to the length and cap size
        capPtr := (*int)(unsafe.Pointer(capAddr)) //
        *lenPtr = lf                              // Assign the actual slice size and cap
        *capPtr = lf                              //

        return buf
}

func float32SliceAsByteSlice(arr []float32) []byte {
        lf := 4 * len(arr)

        // step by step
        pf := &(arr[0])                        // To pointer to the first byte of b
        up := unsafe.Pointer(pf)                  // To *special* unsafe.Pointer, it can be converted to any pointer
        pi := (*[1]byte)(up)                      // To pointer as byte array
        buf := (*pi)[:]                           // Creates slice to our array of 1 byte
        address := unsafe.Pointer(&buf)           // Capture the address to the slice structure
        lenAddr := uintptr(address) + uintptr(8)  // Capture the address where the length and cap size is stored
        capAddr := uintptr(address) + uintptr(16) // WARNING: This is fragile, depending on a go-internal structure.
        lenPtr := (*int)(unsafe.Pointer(lenAddr)) // Create pointers to the length and cap size
        capPtr := (*int)(unsafe.Pointer(capAddr)) //
        *lenPtr = lf                              // Assign the actual slice size and cap
        *capPtr = lf                              //

        return buf
}

var Float32AudioBuffer []float32
var Float32AudioBufferOffset int
var Int16AudioBufferOffset int
var Int16AudioBuffer []int16

const MaxSamplesInBuffer = 2000000

func initAudio() {
	Float32AudioBuffer = make([]float32, MaxSamplesInBuffer)
	Int16AudioBuffer = make([]int16, MaxSamplesInBuffer)
	Float32AudioBufferOffset = 0
	Int16AudioBufferOffset = 0
}

func audioStartRecording() {
	Float32AudioBufferOffset = 0
	Int16AudioBufferOffset = 0
}

func audioEndRecording() {
	//////resampledBuffer := resampleFloat32Stream(Float32AudioBuffer[:Float32AudioBufferOffset], 44100, 8000)
	Float32toInt16(Float32AudioBuffer[:Float32AudioBufferOffset], Int16AudioBuffer, Float32AudioBufferOffset)
	bytes := int16SliceAsByteSlice(Int16AudioBuffer[:Float32AudioBufferOffset])
	//bytes := float32SliceAsByteSlice(Float32AudioBuffer[:Float32AudioBufferOffset])


/*
	// write bytes to file!!
	file, err := os.Create("audio.raw")
    if err != nil {
        log.Fatal(err)
    }
    file.Write(bytes)
    file.Close()
    fmt.Println("    >> audio.raw written (in theory)")

*/



	// send to recipients
	
   // for i := 0 ; i < len(deviceAudioConnections) ; i++ {
   	for i := 0 ; i < len(DevicesToSendAudioTo) ; i++ {
    	blockSize := 1024 // @TODO must make sure to add padding if needed
    	sendDeviceId := DevicesToSendAudioTo[i]
    	_ = deviceAudioConnections[sendDeviceId].WriteMessage(1, []byte("start"))
    	for j := 0 ; j < Float32AudioBufferOffset/blockSize; j++ {
    		_ = deviceAudioConnections[sendDeviceId].WriteMessage(1, bytes[j*2*blockSize:(j+1)*2*blockSize])
    	}
    	_ = deviceAudioConnections[sendDeviceId].WriteMessage(1, []byte("end"))
	}	

}

func audioStreamData(frame []byte) {
	floats := byteSliceAsFloat32Slice(frame)
	_ = copy(Float32AudioBuffer[Float32AudioBufferOffset:], floats)
	Float32AudioBufferOffset = Float32AudioBufferOffset + len(floats)
}

func sampleAverage(inData []float32) float32 {
	result := float32(0.0)
	for i:=0 ; i < len(inData) ; i++ {
		result += inData[i]
	}
	return result/float32(len(inData))
}

func resampleFloat32Stream(inData []float32, inSampleRate int, outSampleRate int) []float32 {
	var result []float32
	skipValues := float32(inSampleRate)/float32(outSampleRate)
	sampleWindowStart := float32(0)
	sampleWindowEnd := float32(0)
	if skipValues < float32(1.0) {
		return nil
	}
	
	outSamples := int(float32(len(inData))/skipValues)
	result = make([]float32, outSamples)
	
	for i:=0; i < outSamples; i++ {
		sampleWindowEnd += skipValues
		start := int(sampleWindowStart)
		end := int(sampleWindowEnd)
		result[i] = sampleAverage(inData[start:end])
		sampleWindowStart += skipValues
	}

	return result
}

func Float32toInt16(inData []float32, outData []int16, length int) {
	MaxValue := float32(32767.0)
	for i := 0 ; i < length; i++ {

		sample := int16(inData[i] * MaxValue)
		outData[i] = sample

	}
}


/////////////////////////////////////////////////////
//    Main function                                //
/////////////////////////////////////////////////////

func main() {
 
    LocalIP = GetLocalIP()
    fmt.Println("Silent Education Server started on IPv4: " + LocalIP)
	fmt.Println("Panel de control ->  http://127.0.0.1:" + RESTPort)
	initAudio()
	startServiceDiscovery()
	go setupWebServer()
	setupWebsocket()
}
