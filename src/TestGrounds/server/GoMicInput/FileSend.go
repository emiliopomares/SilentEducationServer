package main

import (
	"github.com/gordonklaus/portaudio"
	"fmt"
	"log"
	"os"
	"os/signal"
	"unsafe"
)

const UnicastUDPPort string = "9190"

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

func main() {
	if len(os.Args) < 2 {
		fmt.Println("missing required argument:  ip address to stream to")
		return
	}
	fmt.Println("Recording.  Press Ctrl-C to stop.")

	file, err := os.Create(os.Args[1])
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	nSamples := 0

	portaudio.Initialize()
	defer portaudio.Terminate()
	in := make([]int16, 48)
	inBytes := int16SliceAsByteSlice(in)
	stream, err := portaudio.OpenDefaultStream(1, 0, 22050, len(in), in)
	chk(err)
	defer stream.Close()

	chk(stream.Start())
	for {
		chk(stream.Read())
		//chk(binary.Write(f, binary.BigEndian, in))
		file.Write(inBytes)
		nSamples += len(in)
		select {
		case <-sig:
			return
		default:
		}
	}
	chk(stream.Stop())

}

func chk(err error) {
	if err != nil {
		panic(err)
	}
}
