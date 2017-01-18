package main

import (
	"github.com/apex/log"
	"github.com/apex/log/handlers/text"
	"github.com/zhaoyao/serverkit"
	"net/http"
	_ "net/http/pprof"
	"os"
)

func main() {
	printBanner()
	initLog()
	go func() {
		http.ListenAndServe("localhost:6060", nil)
	}()

	srv, err := serverkit.NewServer(nil)
	if err != nil {
		log.WithError(err).Fatal("init failed")
	}

	router := newRouter()

	srv.Register(&serverkit.TCPService{
		Bind:     "0.0.0.0",
		Port:     4320,
		InitConn: newTCPConn(router),
	})

	srv.Register(router)

	if err := srv.Start(); err != nil {
		log.WithError(err).Fatal("start failed")
	}

	log.Info("avalon started")
	srv.StopOnSignal()
}

func initLog() {
	log.SetHandler(text.New(os.Stdout))
	log.SetLevel(log.DebugLevel)
}
