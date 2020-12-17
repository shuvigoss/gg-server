package main

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var exts = map[string]bool{".zip": true, ".tar.gz": true, ".tar": true}

type QueryResult struct {
	Name  string   `json:"name"`
	Child []string `json:"versions"`
}

func InitRouter(f Flag) {
	router := gin.Default()
	router.POST("/upload", upload)
	router.GET("/query", query)
	router.GET("/download", download)
	router.GET("/check", check)

	port := fmt.Sprintf(":%d", f.Port)
	logrus.Infof("start server with port %s", port)
	if err := router.Run(port); err != nil {
		logrus.Fatal("start server error :" + err.Error())
	}
}

func upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		logrus.Errorf("文件上传异常 %v", err)
		c.JSON(http.StatusOK, FailWithMsg(BizError, "文件上传异常："+err.Error()))
		return
	}

	tmpDir, err := checkFile(file, c)
	if err != nil {
		c.JSON(http.StatusOK, FailWithMsg(BizError, "文件检查异常："+err.Error()))
		return
	}
	defer os.RemoveAll(tmpDir)

	helper, err := NewHelper(tmpDir, filepath.Base(file.Filename))
	if err != nil {
		c.JSON(http.StatusOK, FailWithMsg(BizError, "解压文件异常："+err.Error()))
		return
	}

	if err = helper.Check(); err != nil {
		c.JSON(http.StatusOK, FailWithMsg(BizError, "文件校验异常："+err.Error()))
		return
	}

	if err = helper.Save(); err != nil {
		c.JSON(http.StatusOK, FailWithMsg(BizError, "文件存储异常"+err.Error()))
		return
	}

	c.JSON(http.StatusOK, SuccessMsg(nil, "成功"))
}

func query(c *gin.Context) {
	name := c.Query("name")
	dir := viper.GetString("store")
	res := []QueryResult{}
	dirs := ListDir(dir, name)
	for _, d := range dirs {
		children := ListDir(path.Join(dir, d), "")
		tmp := QueryResult{Name: d, Child: children}
		res = append(res, tmp)
	}

	c.JSON(http.StatusOK, SuccessMsg(res, "OK"))
}

func download(c *gin.Context) {
	name := c.Query("name")
	version := c.Query("version")
	target := getTarget(name, version)
	children := ListDir(target, "")
	file := getZipFile(children, name)
	if file != "" {
		c.Writer.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=%s", file))
		c.Writer.Header().Add("Content-Type", "application/octet-stream")
		c.File(filepath.Join(target, file))
		return
	}
	logrus.Errorf("没有找到相应文件 name :%s, version :%s", name, version)
	c.JSON(http.StatusOK, FailWithMsg(BizError, "文件下载异常,未找到相应内容"))
}

func check(c *gin.Context) {
	name := c.Query("name")
	version := c.Query("version")
	target := getTarget(name, version)
	children := ListDir(target, "")
	file := getZipFile(children, name)
	sha256File, err := ioutil.ReadFile(filepath.Join(target, file+".sha256"))
	if err != nil {
		logrus.Errorf("没有找到sha256文件 %s, %s, %s", name, version, file)
		c.JSON(http.StatusOK, FailWithMsg(BizError, "未找到sha256文件"))
		return
	}

	c.JSON(http.StatusOK, SuccessMsg(string(sha256File), "OK"))
}

func getZipFile(children []string, name string) string {
	for _, f := range children {
		match, _ := regexp.Match("^"+name+".*\\.(zip|gz|tar)$", []byte(f))
		if match {
			return f
		}
	}
	return ""
}

func getTarget(name string, version string) string {
	dir := viper.GetString("store")
	dirs := ListDirAccurate(dir, name)

	target := ""
	for _, d := range dirs {
		if d == name {
			children := ListDir(path.Join(dir, d), "")
			if len(children) == 0 {
				break
			}
			if version == "" {
				sort.Strings(children)
				target = path.Join(dir, d, children[len(children)-1])
				break
			}

			for _, v := range children {
				if v == version {
					target = path.Join(dir, d, v)
					break
				}
			}
			break
		}
	}
	return target
}

func checkFile(file *multipart.FileHeader, c *gin.Context) (string, error) {
	//step1 查看文件后缀是否正确
	if err := checkFileExt(file); err != nil {
		return "", err
	}
	//step2 存储到临时目录
	tempDir, err := saveFileTmp(file, c)
	if err != nil {
		return "", err
	}
	return tempDir, nil

}

func saveFileTmp(file *multipart.FileHeader, c *gin.Context) (string, error) {
	tempDir := path.Join(os.TempDir(), uuid.NewV4().String())
	_ = os.MkdirAll(tempDir, os.ModePerm)
	fileName := filepath.Base(file.Filename)
	logrus.Infof("save upload file %s to %s", fileName, tempDir)
	if err := c.SaveUploadedFile(file, path.Join(tempDir, fileName)); err != nil {
		logrus.Error("save fail")
		return "", err
	}

	return tempDir, nil
}

func checkFileExt(file *multipart.FileHeader) error {
	ext := filepath.Ext(file.Filename)
	_, ok := exts[strings.ToLower(ext)]
	if !ok {
		return errors.New("错误的文件类型，上传文件仅支持zip,tar.gz,tar")
	}
	return nil
}
