package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"time"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xinerama"
)

// TODO: don't hardcode resolution stuffs
const (
	terabyte           = 1024 * 1024 //* 1024 * 1024
	default_screen_res = "1920x1080"
	default_screen_pos = "0,0"
)

func displayNum() string {
	d := os.Getenv("DISPLAY")
	if d == "" {
		return ":0"
	}
	return d
}

func displayBounds() (res, pos string) {
	res = default_screen_res
	pos = default_screen_pos

	c, err := xgb.NewConn()
	if err != nil {
		return
	}
	defer c.Close()

	err = xinerama.Init(c)

	reply, err := xinerama.QueryScreens(c).Reply()
	if err != nil || int(reply.Number) <= 0 {
		return
	}

	screen := reply.ScreenInfo[0]
	x := int(screen.XOrg)
	y := int(screen.YOrg)
	w := int(screen.Width)
	h := int(screen.Height)

	res = fmt.Sprintf("%dx%d", w, h)
	pos = fmt.Sprintf("%d,%d", x, y)
	return
}

func runFfmpeg(w http.ResponseWriter, req *http.Request) {
	log.Printf("new request: %+v", req)
	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Connection", "keep-alive")

	rng := req.Header.Get("Range")

	// fake we support partial responses
	if rng != "" {
		w.Header().Set("Accept-Ranges", "bytes")
	}

	res, pos := displayBounds()
	cmd := exec.Command("ffmpeg",
		"-loglevel", "error",
		"-video_size", res,
		"-framerate", "10",
		"-f", "x11grab",
		"-i", displayNum()+".0+"+pos,
		"-vf", "scale=1920:1080,format=yuv420p",
		"-g", "10",
		"-vcodec", "libx264",
		"-tune", "zerolatency",
		"-preset", "fast",
		"-f", "mp4",
		"-movflags", "frag_keyframe+empty_moov+isml",
		"-")

	if rng == "bytes=0-1" {
		log.Printf("serve: found probing request, serving partial content")
		w.Header().Set("Content-Range", fmt.Sprintf("bytes 0-1/%d", terabyte))
		w.Header().Set("Content-Length", "2")
		w.WriteHeader(http.StatusPartialContent)
		if _, err := w.Write([]byte{0, 0}); err != nil {
			log.Printf("serve: failed to write fake response (%v)", err)
		}
		return
	}

	// cmd := exec.Command("cat", "/home/stefan/code/gocode/src/github.com/breunigs/luftfluss/output.mp4")

	cmd.Stderr = os.Stderr

	if rng != "" {
		w.Header().Set("Content-Range", fmt.Sprintf("bytes 0-%d/%d", terabyte-1, terabyte))
		w.Header().Set("Content-Length", fmt.Sprintf("%d", terabyte))
		w.WriteHeader(http.StatusPartialContent)
	}

	log.Printf("serve: connecting le stdout to writer")
	cmd.Stdout = w

	log.Printf("serve: starting le command")
	if err := cmd.Start(); err != nil {
		log.Fatalf("ffmpeg: failed to start (%v)", err)
	}

	log.Printf("ffmpeg: started desktop stream")
	err := cmd.Wait()
	if strings.Contains(fmt.Sprintf("%s", err), "broken pipe") {
		log.Printf("ffmpeg: http connection was closed (%v)", err)
		time.Sleep(10 * time.Second)
	} else if err != nil {
		log.Fatalf("ffmpeg: exited unexpectedly (%v)", err)
	}
}

func getOutboundIP() net.IP {
	// since we send no data, this doesn't actually connect
	conn, err := net.Dial("udp", "129.206.27.40:80")
	if err != nil {
		log.Fatalf("serve: fail to obtain outbound IP (%v)", err)
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP
}

func Serve() string {
	// setup listener
	// listener, err := net.Listen("tcp", ":0")
	listener, err := net.Listen("tcp", ":8000")
	if err != nil {
		log.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	ip := getOutboundIP()
	url := fmt.Sprintf("http://%s:%d/totally_legit.mp4", ip, port)

	srv := &http.Server{}

	// handle shutdowns
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		// We received an interrupt signal, shut down.
		if err := srv.Shutdown(context.Background()); err != nil {
			// Error from closing listeners, or context timeout:
			log.Printf("serve: shutting down (%v)", err)
		}
	}()

	// serve content
	http.HandleFunc("/", runFfmpeg)
	go func() {
		if err = srv.Serve(listener); err != nil {
			log.Fatal(err)
		}
	}()

	log.Printf("serve: desktop stream at %s", url)
	return url
}
