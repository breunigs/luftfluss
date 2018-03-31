package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"howett.net/plist"
)

// const (
//  maxWidth  = 1920
//  maxHeight = 1080
// )

func main() {
	address := os.Args[1]

	url := Serve()
	time.Sleep(10 * time.Hour)
	// url := "http://192.168.178.22:8000/output.mp4"

	params := map[string]string{
		"Content-Location": url,
		"Start-Position":   "0.0",
	}

	reader, writer := io.Pipe()
	go func() {
		defer writer.Close()
		encoder := plist.NewEncoder(writer)
		encoder.Encode(params)
	}()

	client := &http.Client{}

	req, err := http.NewRequest(http.MethodPost, "http://"+address+"/play", reader)
	if err != nil {
		fmt.Println(err)
		time.Sleep(1 * time.Second)
		os.Exit(1)
	}
	req.Header.Add("X-Apple-Device-ID", "0x0C4DE9B8169C")
	req.Header.Add("User-Agent", "iTunes/12.5.4 (Macintosh; OS X 10.12.2)")
	req.Header.Add("Content-Type", "application/x-apple-binary-plist")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		time.Sleep(1 * time.Second)
		os.Exit(1)
	}
	fmt.Printf("%+v  \n\n", resp)

	time.Sleep(1 * time.Hour)
}
