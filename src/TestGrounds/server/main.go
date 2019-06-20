/*
This script creates a simple UDP Server that exports all data received 
through the 8080 socket into the console.
Originally Made By: Roberto E. Zubieta
Salvaged by: Emilio Pomares
G+: https://plus.google.com/u/0/105524772414753584405/
*/

package main

import (
	"fmt"
//	"os"
	"strconv"
	"math"
	"net"
	"github.com/gordonklaus/portaudio"
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

	// audio
	portaudio.Initialize()
	defer portaudio.Terminate()


	stream, err := portaudio.OpenDefaultStream(0, numberOfChannels, sampleRate, bufferSize,
     		func(out []int16) {
			if availableFrames > 0 {
				
				for i:=range out {
					out[i] = buffer[i+bufferSize*numberOfChannels*(readBank)]
				}
				readBank = (readBank + 1) % nBanks
				availableFrames--
				return
			}
			for i:=range out {
				out[i] = 0
			}
		})
	stream.Start()
	defer stream.Close()

	fmt.Printf("frames: %d\n", availableFrames)

	//Keep calling this function
	for {

		addData(udpConn)
	}

}

func addData(conn *net.UDPConn) {

	var maxShortVal int16 = 0
	var buf [2048]byte
	n, err := conn.Read(buf[0:])
	fmt.Printf("%d bytes received\n", n)
	if n != bufferSize * bytesPerSample * numberOfChannels {
		fmt.Println("Packet dropped, should be length: " + strconv.Itoa(bufferSize * bytesPerSample * numberOfChannels))
		return
	}
	if err != nil {
		fmt.Println("Error Reading")
		return
	} else {
		//fmt.Println(hex.EncodeToString(buf[0:n]))
		//fmt.Printf("Package Done, size: %d  \n", n)
		if availableFrames < nBanks {
			for i := 0 ; i < bufferSize * numberOfChannels; i++ {
				shortval := int16(buf[i*2]) + int16(buf[i*2+1]) << 8
				if(shortval > maxShortVal) { maxShortVal = shortval }
				if(shortval > max) { max = shortval }
				if(shortval < min) { min = shortval }
				buffer[i+bufferSize*numberOfChannels*(writeBank)] = shortval
			}
			availableFrames++
			writeBank = (writeBank + 1) % nBanks
		} else {
			//fmt.Println("Warning: buffer full")
		}
	}
	//for i := 0 ; i < n; i++ {
	//	topeBuffer[i+receivedBytes] = buf[i]
	//}
	receivedBytes = receivedBytes + n
	packetsReceived++
	//fmt.Printf("min: %d, max %d, this frame: %d, total bytes: %d\n", min, max, maxShortVal, receivedBytes)

}


