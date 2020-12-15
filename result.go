package main

import (
	"net/http"
)

const (
	ServerError   = 500 //服务端异常
	BizError      = 400 //业务异常
	ServerSuccess = 200 //服务正常
)

type Result struct {
	Status  int         `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func SuccessWithCode(code int, data interface{}) Result {
	return Result{
		Status:  code,
		Message: "OK",
		Data:    data,
	}
}

func SuccessMsg(data interface{}, msg string) Result {
	return Result{
		Status:  http.StatusOK,
		Message: msg,
		Data:    data,
	}
}

func FailWithData(code int, data interface{}) Result {
	return Result{
		Status:  code,
		Message: "FAIL",
		Data:    data,
	}
}

func FailWithMsg(code int, message string) Result {
	return Result{
		Status:  code,
		Message: message,
		Data:    nil,
	}
}

func IsSuccess(result *Result) bool {
	return result.Status == http.StatusOK
}
