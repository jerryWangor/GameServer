package tcp

/**
 * A tcp server
 */

import (
	"bufio"
	"context"
	"fmt"
	"gameserver/tcp/sync/atomic"
	"gameserver/tcp/sync/wait"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Config stores tcp server properties
type Config struct {
	Address    string        `yaml:"address"`
	MaxConnect uint32        `yaml:"max-connect"`
	Timeout    time.Duration `yaml:"timeout"`
}

// handler 是应用层服务器的抽象
type Handler interface {
	Handle(ctx context.Context, conn net.Conn)
	Close() error
}

// ListenAndServeWithSignal 监听中断信号并通过 closeChan 通知服务器关闭
func ListenAndServeWithSignal(cfg *Config, handler Handler) error {
	closeChan := make(chan struct{})
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		sig := <-sigCh
		switch sig {
		case syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			closeChan <- struct{}{}
		}
	}()
	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return err
	}
	log.Println(fmt.Sprintf("bind: %s, start listening...", cfg.Address))
	ListenAndServe(listener, handler, closeChan)
	return nil
}

// 监听并提供服务，并在收到 closeChan 发来的关闭通知后关闭
func ListenAndServe(listener net.Listener, handler Handler, closeChan <-chan struct{}) {
	// 监听关闭通知
	go func() {
		<-closeChan
		log.Println("shutting down...")
		_ = listener.Close() // 停止监听，listener.Accept()会立即返回 io.EOF
		_ = handler.Close()  // 关闭应用层服务器
	}()

	// 在异常退出后释放资源
	defer func() {
		// close during unexpected error
		_ = listener.Close()
		_ = handler.Close()
	}()
	// 返回一个空的Context
	ctx := context.Background()
	var waitDone sync.WaitGroup
	for {
		// 监听端口, 阻塞直到收到新连接或者出现错误
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("accept err", err)
			break
		}
		// 开启 goroutine 来处理新连接
		fmt.Println("accept link")
		waitDone.Add(1)
		go func() {
			defer func() {
				waitDone.Done()
			}()
			handler.Handle(ctx, conn)
		}()
	}
	waitDone.Wait()
}

// 客户端连接的抽象
type ServeClient struct {
	// tcp 连接
	Conn net.Conn
	// 当服务端开始发送数据时进入waiting, 阻止其它goroutine关闭连接
	// wait.Wait是作者编写的带有最大等待时间的封装:
	// https://github.com/HDT3213/godis/blob/master/src/lib/sync/wait/wait.go
	Waiting wait.Wait
}

// 客户端处理函数
type ServeHandler struct {
	activeConn sync.Map
	closing    atomic.Boolean
}

// 处理客户端发来的消息
func (h *ServeHandler) Handle(ctx context.Context, conn net.Conn) {
	// 关闭中的 handler 不会处理新连接
	if h.closing.Get() {
		_ = conn.Close()
	}

	client := &ServeClient{
		Conn: conn,
	}
	// 记住仍然存活的连接
	h.activeConn.Store(client, struct{}{})

	reader := bufio.NewReader(conn)
	for {
		// may occurs: client EOF, client timeout, server early close
		msg, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				fmt.Println("connection close")
				h.activeConn.Delete(client)
			} else {
				fmt.Println(err)
			}
			return
		}
		// 发送数据前先置为waiting状态，阻止连接被关闭
		client.Waiting.Add(1)
		// 模拟关闭时未完成发送的情况
		//logger.Info("sleeping")
		//time.Sleep(10 * time.Second

		b := []byte(msg)
		_, _ = conn.Write(b)
		// 发送完毕, 结束waiting
		client.Waiting.Done()
	}
}

// 关闭客户端连接
func (c *ServeClient) Close() error {
	// 等待数据发送完成或超时
	c.Waiting.WaitWithTimeout(10 * time.Second)
	c.Conn.Close()
	return nil
}

// 关闭服务器
func (h *ServeHandler) Close() error {
	fmt.Println("handler shutting down...")
	h.closing.Set(true)
	// 逐个关闭连接
	h.activeConn.Range(func(key interface{}, val interface{}) bool {
		client := key.(*ServeClient)
		_ = client.Close()
		return true
	})
	return nil
}
