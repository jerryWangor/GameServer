package main

import (
	"fmt"
	"gameserver/model"
	"github.com/gin-gonic/gin"
)

// Register check 结构体
type RegisterC struct {
	Account  string `form:"account" binding:"required,len=11"`
	Password string `form:"password binding:"required" validate:"min=6,nefield=Account"`
	Sex int `form:"sex" validate:"min=1,max=2"`
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

// 统一返回json数据
func ReturnJson(c *gin.Context, httpcode int, code int, msg string, data interface{}) {
	c.JSON(httpcode, gin.H{
		"code": code,
		"msg":  msg,
		"data": data,
	})
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
		ReturnJson(c, 200, 101, "account is null", data)
	} else {
		// 从数据库查询是否存在相通的account，这里可以进行缓存
		fmt.Println("account：", account)
		var accountinfo = model.GetAccountInfo(account)
		if accountinfo != nil {

		}
		ReturnJson(c, 200, 200, "success", accountinfo)
	}
}

func RegisterFunc(c *gin.Context) {
	var registerc RegisterC
	if err := c.ShouldBindQuery(&registerc); err != nil {
		ReturnJson(c, 200, 101, "params error", "")
	}
	// 判断是否注册过了
	var accountinfo = model.GetAccountInfo(registerc.Account)
	if accountinfo != "" {
		ReturnJson(c, 200, 102, "account is exists", "")
	}
	// 注册成功写入数据库

}

func LoginFunc(c *gin.Context) {

	// 获取参数，进行判断
}
