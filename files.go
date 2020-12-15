package main

import (
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"regexp"
)

func PathExists(p string) bool {
	_, err := os.Stat(p)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func ListDir(dir, keywords string) []string {
	dirs, _ := ioutil.ReadDir(dir)
	res := make([]string, 0)
	if keywords == "" {
		for _, d := range dirs {
			res = append(res, d.Name())
		}
		return res
	}

	compile, err := regexp.Compile("(?i).*" + keywords + ".*")
	if err != nil {
		logrus.Errorf("正则匹配异常 %s", keywords)
		return res
	}

	for _, d := range dirs {
		if compile.Match([]byte(d.Name())) {
			res = append(res, d.Name())
		}
	}
	return res
}

func ListDirAccurate(dir, name string) []string {
	dirs, _ := ioutil.ReadDir(dir)
	res := make([]string, 0)

	for _, d := range dirs {
		if name == d.Name() {
			res = append(res, d.Name())
		}
	}
	return res
}

func Copy(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}
