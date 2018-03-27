package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// const (
//  maxWidth  = 1920
//  maxHeight = 1080
// )

func main() {
	address := os.Args[1]

	params := []string{
		"Content-Location: http://192.168.178.22:8000/VID_20130604_223454.mp4",
		"Start-Position: 0.001",
	}

	body := strings.NewReader(strings.Join(params, "\r\n"))

	// reader, writer := io.Pipe()
	// go func() {
	//  defer writer.Close()
	//  // err = jpeg.Encode(writer, img, &jpeg.Options{Quality: 50})
	//  err = png.Encode(writer, img)
	//  if err != nil {
	//    fmt.Println(err)
	//    return
	//  }

	//  // resize.Thumbnail(maxWidth, maxHeight uint, img image.Image, interp resize.InterpolationFunction) image.Image

	// }()
	client := &http.Client{}

	req, err := http.NewRequest(http.MethodPost, "http://"+address+"/play", body)
	if err != nil {
		fmt.Println(err)
		time.Sleep(1 * time.Second)
		os.Exit(1)
	}
	req.Header.Add("User-Agent", "iTunes/10.6 (Macintosh; Intel Mac OS X 10.7.3) AppleWebKit/535.18.5")
	req.Header.Add("Content-Type", "text/parameters")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		time.Sleep(1 * time.Second)
		os.Exit(1)
	}
	fmt.Printf("%+v  \n\n", resp)

	time.Sleep(1 * time.Hour)
}
