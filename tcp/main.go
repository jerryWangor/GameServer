package main

import (
	"gameserver/tcp/tcp"
)

func main() {
	//tcp config
	var config tcp.Config
	config.Address = "127.0.0.1:20001"
	config.MaxConnect = 10000
	config.Timeout = 60

	// 创建
	shandler := tcp.ServeHandler{}
	tcp.ListenAndServeWithSignal(&config, &shandler)
}
