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
	"math"
	"net"
//	"github.com/gordonklaus/portaudio"
)

var readBank = 0
var writeBank = 0

const nBanks = 4

const bufferSize = 96

const sampleRate = 8000
const bytesPerSample = 2

var availableFrames = 0

var packetsReceived = 0

func main() {

	// networking

	//Basic variables
	port := ":8080"
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

	buffer := make([]int16, bufferSize * nBanks)
        for i:= range buffer {
                buffer[i] = int16(2000.0*math.Sin(float64(i)/3.0))
         }
/*
	// audio
	portaudio.Initialize()
	defer portaudio.Terminate()


	stream, err := portaudio.OpenDefaultStream(0, 1, sampleRate, bufferSize,
     		func(out []int16) {
			if availableFrames > 0 {
				for i:=range out {
					out[i] = buffer[i+bufferSize*(availableFrames-1)]
				}
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
*/
	//Keep calling this function
	for {
		fmt.Println("listening for frame...")
		addData(udpConn, buffer)
	}

}

func addData(conn *net.UDPConn, sharedBuffer []int16) {

	var buf [2048]byte
	n, err := conn.Read(buf[0:])
	if n != bufferSize * bytesPerSample {
		fmt.Println("Packet dropped")
		return
	}
	if err != nil {
		fmt.Println("Error Reading")
		return
	} else {
		//fmt.Println(hex.EncodeToString(buf[0:n]))
		//fmt.Printf("Package Done, size: %d  \n", n)
		for i := 0 ; i < bufferSize; i++ {
			sharedBuffer[i+bufferSize*(availableFrames-1)] = int16(buf[i*2]) + int16(buf[i*2+1]) << 8
		}
		availableFrames++
	}
	packetsReceived++
	fmt.Printf("frames: %d, packets received: %d\n", availableFrames, packetsReceived)

}
