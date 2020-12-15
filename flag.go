package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"
)

type Flag struct {
	Port int
	Dir  string
}

func InitFlag() Flag {
	log.Info("start to init flag!")
	var serverPort = flag.IntP("port", "p", 8080, "start port")
	dir, _ := os.UserHomeDir()

	var dataDir = flag.StringP("dir", "d", dir, "")

	f := Flag{
		Port: *serverPort,
		Dir:  *dataDir,
	}

	log.Infof("flag is %v", f)
	return f
}
