package main

import (
	"context"
	"fmt"
	"gameserver/tcp/tcp"
	"net"
)

func main() {
	var err error
	closeChan := make(chan struct{})
	listener, err := net.Listen("tcp", "127.0.0.1:20001")
	if err != nil {
		fmt.Println("TCP服务器监听失败：", err)
		return
	}
	go tcp.ListenAndServe(listener, ServeHandler(), closeChan)
}

// 建立应用层服务器处理函数
