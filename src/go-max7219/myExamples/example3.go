package main

import (
	"time"
	"fmt"
	"log"
	"../../go-max7219"
)

func main() {

	mtx := max7219.NewMatrix(8)
	err := mtx.Open(0, 0, 3)
	if err != nil {
		log.Fatal(err)
	}
	defer mtx.Close()

	fmt.Printf("Clearing panels... %d", 8)
	mtx.Device.ClearAll(true)

	delay := time.Duration(60)

	//maxColumn := 8 * 4


	var color int = 1
	
	mtx.Device.SetTextBuffer([]byte("      Pee is stored in the balls"))
	
	for {
		for x:=0; x < 280; x++ {
			mtx.Device.RenderTextBuffer(x, max7219.FontCP437.GetLetterPatterns(), color)
			mtx.Device.FlushPixelFrameBuffer()


			time.Sleep(delay * time.Millisecond)
		}	
		color++
		if(color == 4) {
			color = 1
		}	
	}
} 
