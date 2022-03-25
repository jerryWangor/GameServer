package utils

import (
	"crypto/md5"
	"fmt"
	"strings"
)

type JsonResult struct {
	Code int16       `json:"code"`
	Msg  string      `json:"msg"`
	data interface{} `json:"data"`
}

func ReturnJson(code int16, msg string, data interface{}) JsonResult {
	return JsonResult{Code: code, Msg: msg, data: data}
}

// 获取where条件
func GetWheres(where []string) string {
	var wheres = strings.Join(where, " and ")
	return wheres
}

// md5加密
func GetMd5String(b []byte) string {
	return fmt.Sprintf("%x", md5.Sum(b))
}