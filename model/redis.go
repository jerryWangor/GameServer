package model

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
)

// 连接池的使用
var pool *redis.Pool // 定义

func init() {
	// 初始化
	setpass := redis.DialPassword("jerry123")
	pool = &redis.Pool{
		MaxIdle:     8,   // 最大空闲链接数
		MaxActive:   0,   // 和数据库的最大链接数，0表示没有限制。（当数据有并发问题的时候需要考虑）
		IdleTimeout: 100, // 最大空闲时间
		Dial: func() (redis.Conn, error) { // 初始化链接代码，指明要连接的协议，IP，端口号
			return redis.Dial("tcp", "42.193.50.38:9001", setpass)
		},
	}
}

func SetRedis(name string, value interface{}) bool {
	// 从连接池中取出一个链接
	c := pool.Get()
	defer c.Close() // 应用程序必须关闭返回的连接：回收方法是activeConn中的Close
	_, err := c.Do("Set", name, value, 0)
	if err != nil {
		fmt.Println(err)
		return false
	}

	return true
}

func GetRedisInt(name string) (int, bool) {
	c := pool.Get()
	defer c.Close()

	value, err := redis.Int(c.Do("Get", name))
	if err != nil {
		fmt.Println(err)
		return 0, false
	}

	return value, true
}

func GetRedisString(name string) (string, bool) {
	c := pool.Get()
	defer c.Close()

	value, err := redis.String(c.Do("Get", name))
	if err != nil {
		fmt.Println(err)
		return "", false
	}

	return value, true
}