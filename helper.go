package main

import (
	"encoding/json"
	"errors"
	"github.com/go-playground/validator/v10"
	"github.com/mholt/archiver/v3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

type Helper struct {
	tmpDir   string
	fileName string
	dst      string
	GgJson   GgJson
}

type GgJson struct {
	Name    string `validate:"required"`
	Version string `validate:"required"`
}

var validate = validator.New()

func NewHelper(tmpDir, fileName string) (Helper, error) {
	h := Helper{tmpDir: tmpDir, fileName: fileName}

	if err := archiver.Unarchive(filepath.Join(tmpDir, fileName), tmpDir); err != nil {
		logrus.Errorf("error to unarchive %v", err)
		return h, err
	}
	return h, nil
}

func (h *Helper) Check() error {
	files, _ := ioutil.ReadDir(h.tmpDir)

	index := 0
	hasGGJson := false
	for ; index < len(files); index++ {
		if files[index].Name() == "gg.json" {
			hasGGJson = true
			break
		}
	}

	if !hasGGJson {
		logrus.Error("没有找到gg.json文件!")
		return errors.New("文件中未包含gg.json,请重新打包上传")
	}

	if err := h.parse(); err != nil {
		logrus.Errorf("解析文件异常 %v", err)
		return errors.New("解析文件异常！")
	}
	return nil
}

func (h *Helper) parse() error {
	gg := path.Join(h.tmpDir, "gg.json")
	file, err := ioutil.ReadFile(gg)
	if err != nil {
		return err
	}

	ggJson := GgJson{}
	if err = json.Unmarshal(file, &ggJson); err != nil {
		return err
	}

	h.GgJson = ggJson

	if err = validate.Struct(ggJson); err != nil {
		logrus.Errorf("validate ggjson error %v", err)
		return err
	}

	return nil
}

func (h Helper) Save() error {
	store := viper.GetString("store")
	target := filepath.Join(store, h.GgJson.Name, h.GgJson.Version)
	if exists := PathExists(target); exists {
		logrus.Infof("%s 已经存在，进行替换操作(删除)!", target)
		_ = os.RemoveAll(target)
	} else {
		logrus.Infof("%s 没有存在，创建相关目录", target)
	}

	if err := os.MkdirAll(target, os.ModePerm); err != nil {
		logrus.Errorf("创建目录异常 %s, %v", target, err)
		return err
	}

	if err := Copy(filepath.Join(h.tmpDir, h.fileName), filepath.Join(target, h.fileName)); err != nil {
		logrus.Errorf("拷贝文件夹异常 %v", err)
		return err
	}
	return CreateSha1File(filepath.Join(target, h.fileName))
}
