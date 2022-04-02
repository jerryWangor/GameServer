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
	"gameserver/utils"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// 定义主命令常量
const (
	LOGIN_AUTH = 1001 // 登录验证
	SIGN_DAY   = 1002 // 每日签到
)

// Config stores tcp server properties
type Config struct {
	Address    string        `yaml:"address"`     // 监听地址
	MaxConnect uint32        `yaml:"max-connect"` // 最大连接数
	Timeout    time.Duration `yaml:"timeout"`     // 超时时间
}

// Handler 是应用层服务器的抽象接口
type Handler interface {
	Handle(ctx context.Context, conn net.Conn)
	Close() error
	NormalClose(client *ServeClient) error
}

// ServeClient 客户端连接的抽象
type ServeClient struct {
	Conn      net.Conn  // tcp 连接
	AuthState bool      // 认证状态，连接成功后必须在规定时间内认证，不然就主动断开
	Waiting   wait.Wait // 当服务端开始发送数据时进入waiting, 阻止其它goroutine关闭连接
}

// ServeHandler 服务端处理函数
type ServeHandler struct {
	activeConn sync.Map       // 所有活跃连接，存的是上面的ServeClient，为什么用sync.map呢，是因为在协程里面不会被锁报错
	closing    atomic.Boolean // 关闭状态
}

// ListenAndServeWithSignal 监听中断信号并通过 closeChan 通知服务器关闭
func ListenAndServeWithSignal(cfg *Config, handler Handler) error {
	closeChan := make(chan struct{})
	sigCh := make(chan os.Signal)

	// 我理解是创建一个通道，用于接收发来的信号，如果收到退出信号就发给closeChan通道，执行退出操作
	// 比如我们ctrl+c主动关闭，就会触发，或者在linux服务器上面杀进程
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		// 这里该协程被通道阻塞了，等下次收到命令后继续执行
		sig := <-sigCh
		switch sig {
		case syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			closeChan <- struct{}{}
		}
	}()
	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		log.Println("tcp服务器监听失败，", err)
		return err
	}
	log.Println(fmt.Sprintf("bind: %s, start listening...", cfg.Address))
	ListenAndServe(listener, handler, closeChan)
	return nil
}

// ListenAndServe 监听并提供服务，并在收到 closeChan 发来的关闭通知后关闭
func ListenAndServe(listener net.Listener, handler Handler, closeChan <-chan struct{}) {
	// 监听关闭通知
	go func() {
		<-closeChan
		log.Println("服务器主动关闭...")
		_ = listener.Close() // 停止监听，listener.Accept()会立即返回 io.EOF
		_ = handler.Close()  // 关闭应用层服务器
	}()

	// 在异常退出后释放资源
	defer func() {
		log.Println("服务器defer关闭...")
		_ = listener.Close()
		_ = handler.Close()
	}()

	// 返回一个空的Context
	ctx := context.Background()
	// 创建一个协程计数器
	var wg sync.WaitGroup
	for {
		// 监听端口, 阻塞直到收到新连接或者出现错误
		conn, err := listener.Accept()
		if err != nil {
			log.Println("accept err: ", err)
			break
		}
		// 开启 goroutine 来处理新连接
		log.Println("客户端连接，来自:", conn.RemoteAddr().String())
		wg.Add(1) // 计数器+1
		go func() {
			defer func() {
				wg.Done() // 计数器-1
			}()
			handler.Handle(ctx, conn)
		}()
	}
	wg.Wait() // 等待，直到计数为0，意思就是等待所有子协程执行完成后再结束主协程
}

// Handle 处理客户端发来的消息
func (h *ServeHandler) Handle(ctx context.Context, conn net.Conn) {
	// 关闭中的 handler 不会处理新连接的消息
	if h.closing.Get() {
		_ = conn.Close()
	}

	// 创建客户端结构体
	client := &ServeClient{
		Conn:      conn,
		AuthState: false,
	}
	// 保存存活的连接到sync.map中
	h.activeConn.Store(client, struct{}{})

	// 检查认证
	go client.CheckAuth(h)

	// 从缓存中读取数据，if has
	reader := bufio.NewReader(conn)
	for {
		// 这里以\n为分隔，后期考虑用header+body来处黏包拆包
		//msg, err := reader.ReadString('\n')
		msg, err := reader.ReadBytes('\n')
		if err != nil {
			// 当在Read时，收到一个IO.EOF，代表的就是对端已经关闭了发送的通道，通常来说是发起了FIN
			if err == io.EOF {
				log.Println("客户端主动关闭")
				h.NormalClose(client)
			} else {
				log.Println("read err: ", err)
			}
			return
		}
		// 发送数据前先置为waiting状态，阻止连接被关闭
		client.Waiting.Add(1)

		// 根据接收到的消息执行不同的操作
		UnPackageBytes(msg, client, h)

		// 发送完毕, 结束waiting
		client.Waiting.Done()
	}
}

// Close 关闭服务器处理函数
func (h *ServeHandler) Close() error {
	log.Println("ServeHandler shutting down...")
	h.closing.Set(true)
	// 逐个关闭客户端连接
	h.activeConn.Range(func(key interface{}, val interface{}) bool {
		client := key.(*ServeClient)
		_ = client.Close()
		h.activeConn.Delete(key) // 这里要记住从连接池里面移除
		return true
	})
	return nil
}

// Close 关闭客户端连接
func (c *ServeClient) Close() error {
	// 等待数据发送完成或超时10秒后
	c.Waiting.WaitWithTimeout(10 * time.Second)
	log.Println("主动关闭客户端:", c.Conn)
	c.Conn.Close()
	return nil
}

// NormalClose 关闭正常通信的客户端连接，一般是收到客户端关闭消息，或者验证失败，或者收到不合法消息的时候
func (h *ServeHandler) NormalClose(c *ServeClient) error {
	c.Waiting.WaitWithTimeout(10 * time.Second)
	c.Close()
	h.activeConn.Delete(c)
	return nil
}

// CheckAuth 检查认证
func (c *ServeClient) CheckAuth(h *ServeHandler) {
	log.Println("开始检查auth")
	select {
	case <-time.After(time.Second * 10):
		if c.AuthState == false {
			log.Println("auth认证失败")
			h.NormalClose(c)
		}
	}
}

// UnPackageBytes
// 解析消息包
// 消息检验码 4字节
// 消息长度	4字节
// 身份		8字节
// 主命令	4字节
// 子命令	4字节
// 加密方式	4字节
// 消息体	N字节（字节数组：消息长度+消息+消息长度+消息）
// 分隔符	1字节
func UnPackageBytes(bs []byte, c *ServeClient, h Handler) {

	log.Println("bytes", bs)
	log.Println("bytes len", len(bs))
	// 开始解析bytes
	// 消息检验码
	crccode := append([]byte{0, 0, 0, 0}, bs[0:4]...)
	crccode_i := utils.BytesToInt(crccode)
	log.Println("消息检验码", crccode_i)
	if crccode_i != 65433 {
		log.Println("协议错误", c)
		h.NormalClose(c)
		return
	}
	// 消息长度
	msglen := append([]byte{0, 0, 0, 0}, bs[4:8]...)
	msglen_i := utils.BytesToInt(msglen)
	log.Println("消息长度", msglen_i)

	// 身份（账号ID或者其他）
	accid := bs[8:16]
	accid_i := utils.BytesToInt(accid)
	log.Println("身份", accid_i)

	// 主命令
	mcommand := append([]byte{0, 0, 0, 0}, bs[16:20]...)
	mcommand_i := utils.BytesToInt(mcommand)
	log.Println("主命令", mcommand_i)

	// 子命令
	ccommand := append([]byte{0, 0, 0, 0}, bs[20:24]...)
	ccommand_i := utils.BytesToInt(ccommand)
	log.Println("子命令", ccommand_i)

	// 加密方式
	encrypt := append([]byte{0, 0, 0, 0}, bs[24:28]...)
	encrypt_i := utils.BytesToInt(encrypt)
	log.Println("加密方式", encrypt_i)

	// 消息体
	msgbody := bs[28 : len(bs)-4]
	log.Println("消息体", msgbody)
	// 判断
	if len(msgbody)%8 != 0 {
		log.Println("消息体结构错误")
		return
	}
	num := len(msgbody) / 8
	ms := 0
	for i := 0; i < num; i++ {
		// 每个消息前面都有个是这段消息的长度，这里先不做处理
		ms++
		if ms == 1 {
			continue
		}
		// 取到消息体
		tmsg := msgbody[(i * 8):((i + 1) * 8)]
		log.Println("消息", string(tmsg))
		ms = 0
	}

	// 剩下的是分隔符len(bs)-4
	log.Println("auth检查通过")
	c.AuthState = true
}
