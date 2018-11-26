/*
This script creates a simple UDP Server that exports all data received 
through the 8080 socket into the console.
Originally Made By: Roberto E. Zubieta
Salvaged by: Emilio Pomares
G+: https://plus.google.com/u/0/105524772414753584405/
*/

package main

import (
	"encoding/hex"
	"fmt"
	"net"
)

func main() {

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

	//Keep calling this function
	for {
		display(udpConn)
	}

}

func display(conn *net.UDPConn) {

	var buf [2048]byte
	n, err := conn.Read(buf[0:])
	if err != nil {
		fmt.Println("Error Reading")
		return
	} else {
		fmt.Println(hex.EncodeToString(buf[0:n]))
		fmt.Printf("Package Done, size: %d  \n", n)
	}

}
