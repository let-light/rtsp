package main

import (
	"fmt"
	"os"
	"time"

	"github.com/let-light/rtsp"
	"github.com/sirupsen/logrus"
)

type TestServer struct{}
type TestSession struct {
	*rtsp.ServSession
}

func (ts *TestServer) OnConnect(ss rtsp.IServSession) {
	fmt.Println("connect")
}

func (ts *TestServer) OnDisconnect(ss rtsp.IServSession) {
	fmt.Println("disconnect")
}

func (ts *TestServer) OnShutdown(s *rtsp.Server) {
	fmt.Println("server shutdown")
}

func (ts *TestServer) OnPlay(ss *rtsp.IServSession) error {
	fmt.Println("play")
	return nil
}

func (ts *TestServer) OnPublish(ss *rtsp.IServSession) error {
	fmt.Println("publish")
	return nil
}

func (ts *TestServer) OnPause(ss *rtsp.IServSession) error {
	fmt.Println("pause")
	return nil
}

func (ts *TestServer) OnResume(ss *rtsp.IServSession) error {
	fmt.Println("resume")
	return nil
}

func (ts *TestServer) OnStream(ss *rtsp.IServSession) error {
	fmt.Println("stream")
	return nil
}

func (ts *TestServer) NewOrGet() rtsp.IServSession {
	return &TestSession{ServSession: rtsp.NewServSession(ts)}
}

func main() {
	ts := &TestServer{}
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	log.SetOutput(os.Stdout)
	s, err := rtsp.NewServer(ts, ts, ":8554", rtsp.Options{
		ReusePort:        true,
		ReuseAddr:        true,
		TCPKeepAlive:     10 * time.Second,
		TCPNoDelay:       true,
		LockOSThread:     true,
		SocketRecvBuffer: 1024 * 1024,
		SocketSendBuffer: 1024 * 1024,
		Logger:           log,
		Multicore:        true,
		NumEventLoop:     0,
	})
	if err != nil {
		panic(err)
	}
	s.Run()
}