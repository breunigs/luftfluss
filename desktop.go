package main

import (
	"fmt"
	"image"
	"image/png"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/kbinani/screenshot"
)

func main() {
	address := os.Args[1]

	bounds := screenshot.GetDisplayBounds(0)

	client := &http.Client{}

	var prevImg *image.RGBA
	for {
		img, err := screenshot.CaptureRect(bounds)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if sameImage(img, prevImg) {
			time.Sleep(150 * time.Millisecond)
			continue
		}
		prevImg = img

		reader, writer := io.Pipe()
		go func() {
			defer writer.Close()
			// err = jpeg.Encode(writer, img, &jpeg.Options{Quality: 50})
			err = png.Encode(writer, img)
			if err != nil {
				fmt.Println(err)
				return
			}
		}()

		req, err := http.NewRequest(http.MethodPut, "http://"+address+"/photo", reader)
		if err != nil {
			fmt.Println(err)
			time.Sleep(1 * time.Second)
			continue
		}
		req.Header.Add("X-Apple-Transition", "None")

		_, err = client.Do(req)
		if err != nil {
			fmt.Println(err)
			time.Sleep(1 * time.Second)
			continue
		}
	}
}

func sameImage(img1, img2 *image.RGBA) bool {
	if img1 == nil || img2 == nil {
		return false
	}
	if img1.Bounds() != img2.Bounds() {
		return false
	}
	for i := 0; i < len(img1.Pix); i++ {
		if img1.Pix[i] != img2.Pix[i] {
			return false
		}
	}
	return true
}
