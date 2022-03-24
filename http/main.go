package main

import (
	"fmt"
	"gameserver/model"
	"github.com/gin-gonic/gin"
)

// Register 结构体
type Register struct {
	Account  string `form:"account" binding:"required"`
	Password string `form:"password binding:"required"`
}

// Login 结构体
type Login struct {
	Account  string `form:"account" binding:"required"`
	Password string `form:"password binding:"required"`
}

// go的http服务器，用于玩家登录，获取jwt
func main() {
	r := gin.Default()

	// 注册中间件
	r.Use(MiddleWare())
	r.GET("/check_account", CheckAccountFunc)
	r.GET("/Register", RegisterFunc)
	r.GET("/login", LoginFunc)

	r.Run(":8080")
}

// 定义中间件
func MiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 打印访问来源信息
		fp := c.FullPath()
		fmt.Println(fp)
		c.Next()
	}
}

func CheckAccountFunc(c *gin.Context) {
	account := c.Query("account")
	if account == "" {
		data := struct {
			Name string
			Age  int
		}{
			Name: "Jerry",
			Age:  28,
		}
		c.JSON(200, gin.H{
			"code": 101,
			"msg":  "account is null",
			"data": data,
		})
	} else {
		// 从数据库查询是否存在相通的account，这里可以进行缓存
		fmt.Println("account：", account)
		var accountinfo = model.GetAccountInfo(account)
		c.JSON(200, gin.H{
			"code": 0,
			"msg":  "success",
			"data": accountinfo,
		})
	}
}

func RegisterFunc(c *gin.Context) {

}

func LoginFunc(c *gin.Context) {

	// 获取参数，进行判断
}
