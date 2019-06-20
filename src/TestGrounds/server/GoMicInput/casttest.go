package main

import (
	"fmt"
	"unsafe"
)


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
	myArray := []int16{5, 1000, 30000}
	myByteArr := int16SliceAsByteSlice(myArray)
	fmt.Println(len(myByteArr))
	fmt.Println(myByteArr[0], myByteArr[1], myByteArr[2], myByteArr[3])
	myArray[0] = 256
	fmt.Println(myByteArr[0], myByteArr[1], myByteArr[2], myByteArr[3])
}
