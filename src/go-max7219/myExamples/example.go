package main

import (
	"time"
	"fmt"
	"log"
	"../../go-max7219"
)

func main() {

	mtx := max7219.NewMatrix(8)
	err := mtx.Open(0, 0, 2)
	if err != nil {
		log.Fatal(err)
	}
	defer mtx.Close()

	fmt.Printf("Clearing panels... %d", 8)
	mtx.Device.ClearAll(true)

	//mtx.Device.PutPixel(0, 2, 1, false)
	//mtx.Device.PutPixel(2, 2, 2, false)
	//mtx.Device.PutPixel(4, 2, 3, true)	

	delay := time.Duration(8)

	maxColumn := 8 * 4


	var color int = 0
	
	for {
	
			for x := 0 ; x < maxColumn ; x++ {
	
				if x < (maxColumn) / 3 {
					color = 1
				} else if x < (maxColumn * 2) / 3 {
					color = 3
				} else {
					color = 2
				}	
				mtx.Device.VLine(x, color) 
			        mtx.Device.FlushPixelFrameBuffer()
				time.Sleep(delay * time.Millisecond)
			}


			for x:=maxColumn-1; x >= 0; x-- {
				mtx.Device.VLine(x, 0)	
				mtx.Device.FlushPixelFrameBuffer()
				time.Sleep(delay * time.Millisecond)	
			}
		

	}	
} 
