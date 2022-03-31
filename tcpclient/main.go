package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"strconv"
)

func main() {
	var err error
	//closeChan := make(chan struct{})
	addr := "127.0.0.1:20001"

	//conn, err := net.Dial("tcp", addr)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		fmt.Println("TCP服务器连接失败：", err)
		return
	}

	for i := 0; i < 10; i++ {
		val := strconv.Itoa(rand.Int())
		_, err = conn.Write([]byte(val + "\n"))
		if err != nil {
			fmt.Println(err)
			return
		}
		bufReader := bufio.NewReader(conn)
		line, _, err := bufReader.ReadLine()
		if err != nil {
			fmt.Println(err)
			return
		}
		if string(line) != val {
			fmt.Println("get wrong response")
			return
		}
	}
	//_ = conn.Close()
	//for i := 0; i < 5; i++ {
	//	// create idle connection
	//	_, _ = net.Dial("tcp", addr)
	//}
	//closeChan <- struct{}{}
	//time.Sleep(time.Second)
}
