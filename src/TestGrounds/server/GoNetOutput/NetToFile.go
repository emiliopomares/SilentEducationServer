package main

import (
	"fmt"
	"log"
	"os"
	//"strconv"
	"math"
	"net"
)

var readBank = 0
var writeBank = 0

const nBanks = 4

const bufferSize = 48
//96

const sampleRate = 22050
//8000
const bytesPerSample = 2
const numberOfChannels = 1

var availableFrames = 0

var packetsReceived = 0

var buffer = make([]int16, bufferSize * numberOfChannels * nBanks)

var min int16 = 32767
var max int16 = -32768

var topeBuffer = make([]byte, 80000)
var receivedBytes = 0

func finish() {
	fmt.Println("As I thought...")
}

func main() {


	// networking
	
	//Basic variables
	port := ":9190"
	protocol := "udp"

	//Build the address
	udpAddr, err := net.ResolveUDPAddr(protocol, port)
	if err != nil {
		fmt.Println("Wrong Address")
		return
	}

	file, err := os.Create(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	//Output
	fmt.Println("UDP server listening @ " + udpAddr.String())

	//Create the connection
	udpConn, err := net.ListenUDP(protocol, udpAddr)
	if err != nil {
		fmt.Println(err)
	}

        for i:= range buffer {
                buffer[i] = int16(2000.0*math.Sin(float64(i)/3.0))
         }


	fmt.Printf("frames: %d\n", availableFrames)

	//Keep calling this function
	for {

		addData(file, udpConn)
	}

}

func addData(file *os.File, conn *net.UDPConn) {

	var buf [2048]byte
	n, _ := conn.Read(buf[0:])
	file.Write(buf[:n])
}


