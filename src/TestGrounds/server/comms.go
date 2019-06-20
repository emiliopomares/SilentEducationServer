package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"time"
	"net"
)

type Packet struct {
	SourceIP, DestinationIP, ID, Response string
	Content                               []byte
}

const broadcast_addr string = "192.168.1.255"

const RPORT string = ":9222"
const WPORT string = ":9223"

var LocalIP string

// DataValue should ONLY be int og string
type CommData struct {
	Identifier string
	SenderIP	string
	ReceiverIP	string
	MsgID string
	DataType string
	DataValue interface{}
}

type ConnData struct {
	SenderIP string
	MsgID string
	SendTime time.Time
	Status string
}

func Init(readPort string, writePort string) (<-chan Packet, chan<- Packet) {
	receive := make(chan Packet, 10)
	send := make(chan Packet, 10)
	go listen(receive, readPort)
	go broadcast(send, LocalIP, writePort)
	return receive, send
}

func broadcast(send chan Packet, localIP string, port string) {
	fmt.Printf("COMM: Broadcasting message to: %s%s\n", broadcast_addr, port)
	broadcastAddress, err := net.ResolveUDPAddr("udp", broadcast_addr+port)
	printError("ResolvingUDPAddr in Broadcast failed.", err)
	localAddress, err := net.ResolveUDPAddr("udp", GetLocalIP())
	connection, err := net.DialUDP("udp", localAddress, broadcastAddress)
	printError("DialUDP in Broadcast failed.", err)

	localhostAddress, err := net.ResolveUDPAddr("udp", "localhost"+port)
	printError("ResolvingUDPAddr in Broadcast localhost failed.", err)
	lConnection, err := net.DialUDP("udp", localAddress, localhostAddress)
	printError("DialUDP in Broadcast localhost failed.", err)
	defer connection.Close()

	var buffer bytes.Buffer
	encoder := gob.NewEncoder(&buffer)
	for {
		message := <-send
		err := encoder.Encode(message)
		printError("Encode error in broadcast: ", err)
		_, err = connection.Write(buffer.Bytes())
		if err != nil {
			_, err = lConnection.Write(buffer.Bytes())
			printError("Write in broadcast localhost failed", err)
		}
		buffer.Reset()
	}
}

func listen(receive chan Packet, port string) {
	localAddress, _ := net.ResolveUDPAddr("udp", port)
	connection, _ := net.ListenUDP("udp", localAddress)
	defer connection.Close()
	var message Packet

	for {
		inputBytes := make([]byte, 4096)
		length, _, _ := connection.ReadFromUDP(inputBytes)
		buffer := bytes.NewBuffer(inputBytes[:length])
		decoder := gob.NewDecoder(buffer)
		_ = decoder.Decode(&message)
		//if message.Key == com_id {
			receive <- message
		//}
	}
}

func PrintMessage(data Packet) {
	fmt.Printf("=== Data received ===\n")
	fmt.Printf("SenderIP: %s\n", data.SourceIP)
	fmt.Printf("ReceiverIP: %s\n", data.DestinationIP)
	fmt.Printf("Message ID: %s\n", data.ID)
	fmt.Printf("= Data = \n")
	fmt.Printf("Data type: %s\n", data.Response)
	fmt.Printf("Content: %v\n", data.Content)
}

func printError(errMsg string, err error) {
	if err != nil {
		fmt.Println(errMsg)
		fmt.Println(err.Error())
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
/*
func ResolveMsg(senderIP string, receiverIP string, msgID string, response string, content map[string]interface{}) (commData *CommData) {
	message := CommData{
		Key:        com_id,
		SenderIP:   senderIP,
		ReceiverIP: receiverIP,
		MsgID:      msgID,
		Response:   response,
		Content:    content,
	}
	return &message
}
*/
func main() {
	
	LocalIP = GetLocalIP()
	fmt.Println("This.IP = " + LocalIP)
	rcv, _ := Init(RPORT, WPORT)
	p := <-rcv
	fmt.Println(p)
}
