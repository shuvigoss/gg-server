package main

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/thoas/go-funk"
	"io/ioutil"
	"os"
	"path"
	"regexp"
)

type ScaffoldInfo struct {
	BasePath     string
	Name         string
	VersionInfos []VersionInfo
}

type VersionInfo struct {
	Version    string
	ZipFile    string
	sha256File string
	Sha256     string
}

func NewScaffoldInfo(basePath, name, version string) (ScaffoldInfo, error) {
	info := ScaffoldInfo{BasePath: basePath, Name: name}
	if !PathExists(path.Join(basePath, name)) {
		return info, errors.New(fmt.Sprintf("在%s下没有名为%s的脚手架", basePath, name))
	}
	if err := info.parseAll(); err != nil {
		return info, errors.New(fmt.Sprintf("ScaffoldInfo初始化异常 %v", err))
	}
	return info, nil
}

func (info ScaffoldInfo) GetTheLastVersionInfo() VersionInfo {
	return info.VersionInfos[0]
}

func (info ScaffoldInfo) GetByVersion(version string) (VersionInfo, bool) {
	if version == "" {
		return info.GetTheLastVersionInfo(), true
	}
	for _, vi := range info.VersionInfos {
		if vi.Version == version {
			return vi, true
		}
	}
	return VersionInfo{}, false
}

func (info *ScaffoldInfo) parseAll() error {
	readDir, err := ioutil.ReadDir(path.Join(info.BasePath, info.Name))
	if err != nil {
		return err
	}
	vis := make([]VersionInfo, 0)
	funk.ForEach(readDir, func(fileInfo os.FileInfo) {
		versionInfo, success := detailInfo(info.BasePath, info.Name, fileInfo.Name())
		if success {
			vis = append(vis, versionInfo)
		}
	})
	info.VersionInfos = vis
	return nil
}

func detailInfo(dir, name, version string) (VersionInfo, bool) {
	vi := VersionInfo{Version: version}

	files, _ := ioutil.ReadDir(path.Join(dir, name, version))
	funk.ForEach(files, func(f os.FileInfo) {
		match, _ := regexp.Match("^"+name+".*\\.(zip|gz|tar)$", []byte(f.Name()))
		if match {
			vi.ZipFile = f.Name()
		}
	})
	if vi.ZipFile == "" {
		return vi, false
	}
	vi.sha256File = vi.ZipFile + ".sha256"
	sha256Path := path.Join(dir, name, version, vi.sha256File)
	sha256, err := ioutil.ReadFile(sha256Path)
	if err != nil {
		logrus.Errorf("读取 %s 文件异常 %v", sha256Path, err)
		return vi, false
	}
	vi.Sha256 = string(sha256)
	return vi, true
}
