package main

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"path"
)

func main() {
	flag := InitFlag()

	dir := makeStore(flag)

	viper.Set("flag", flag)
	viper.Set("store", dir)

	InitRouter(flag)
}

func makeStore(flag Flag) string {
	dir := path.Join(flag.Dir, "gg")
	if !PathExists(dir) {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			logrus.Fatalf("mkdir %s error %v ", dir, err)
		}
	}
	return dir
}
