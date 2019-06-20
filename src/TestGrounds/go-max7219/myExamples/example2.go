package main

import (
	//"time"
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

	//delay := time.Duration(8)

	//maxColumn := 8 * 4


	//var color int = 0
	
	mtx.Device.DrawChar(0, 'S', max7219.FontCP437.GetLetterPatterns(), 1)
mtx.Device.DrawChar(8, 'H', max7219.FontCP437.GetLetterPatterns(), 1)
mtx.Device.DrawChar(16, 'I', max7219.FontCP437.GetLetterPatterns(), 1)
mtx.Device.DrawChar(24, 'T', max7219.FontCP437.GetLetterPatterns(), 1)	



	mtx.Device.FlushPixelFrameBuffer()

} 
