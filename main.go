package main

import (
	"github.com/apex/log"
	"github.com/apex/log/handlers/text"
	"github.com/urfave/cli"
	"github.com/zhaoyao/serverkit"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"strings"
)

var app *cli.App

func main() {
	app = cli.NewApp()

	app.Flags = []cli.Flag{
		cli.StringSliceFlag{
			Name: "backend",
		},
	}

	app.Action = func(c *cli.Context) error {
		backends := c.StringSlice("backend")
		return run(backends...)
	}

	app.Run(os.Args)
}

func run(backends ...string) error {
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
	for _, b := range backends {
		parts := strings.Split(b, "|")
		mc, _ := strconv.Atoi(parts[2])
		if mc <= 0 {
			mc = 1
		}
		bc, err := newBackend(bcModeTCP, parts[1], mc)
		if err != nil {
			return err
		}
		if err := router.AddBackend(parts[0], bc); err != nil {
			return err
		}
	}

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

	return nil
}

func initLog() {
	log.SetHandler(text.New(os.Stdout))
	log.SetLevel(log.DebugLevel)
}
