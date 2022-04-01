package utils

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
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

func IntToBytes(n int) []byte {
	data := int64(n)
	bytebuf := bytes.NewBuffer([]byte{})
	binary.Write(bytebuf, binary.BigEndian, data)
	return bytebuf.Bytes()
}

func BytesToInt(bys []byte) int {
	bytebuff := bytes.NewBuffer(bys)
	var data int64
	binary.Read(bytebuff, binary.BigEndian, &data)
	return int(data)
}
