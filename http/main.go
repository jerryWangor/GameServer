package main

import (
	"fmt"
	"gameserver/model"
	"gameserver/utils"
	"github.com/gin-gonic/gin"
	"strconv"
	"time"
)

// Register check 结构体
type RegisterC struct {
	Account  string `form:"account" binding:"required,len=11"`
	Password string `form:"password" binding:"required,min=6,nefield=Account"`
	Sex      int    `gorm:"default:1" form:"sex" :"min=1,max=2"`
}

type LoginC struct {
	Account  string `form:"account" binding:"required,len=11"`
	Password string `form:"password" binding:"required,min=6,nefield=Account"`
}

// go的http服务器，用于玩家登录，获取jwt
func main() {
	r := gin.Default()

	// 注册中间件
	r.Use(MiddleWare())
	r.GET("/check_account", CheckAccountFunc)
	r.GET("/register", RegisterFunc)
	r.GET("/login", LoginFunc)
	r.GET("/get_token", GetTokenFunc)

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
	fmt.Println("返回数据", gin.H{
		"code": code,
		"msg":  msg,
		"data": data,
	})
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
		accountinfo, err := model.GetAccountInfo(account)
		// 如果没有查到
		if err != nil {
			ReturnJson(c, 200, 102, "mysql select error", "")
		} else {
			if len(accountinfo) > 0 {
				ReturnJson(c, 200, 103, "account exists", "")
			} else {
				ReturnJson(c, 200, 200, "success", "")
			}
		}

	}
}

func RegisterFunc(c *gin.Context) {
	var registerc RegisterC
	if err := c.ShouldBindQuery(&registerc); err != nil {
		ReturnJson(c, 200, 101, "params error", "")
		return
	}
	// 判断是否注册过了
	accinfo, err := model.GetAccountInfo(registerc.Account)
	if err == nil {
		if len(accinfo) > 0 {
			ReturnJson(c, 200, 102, "account is exists", "")
			return
		}
	} else {
		ReturnJson(c, 200, 102, "get account error", "")
		return
	}

	// 注册成功写入数据库
	var account model.Account
	account.Account = registerc.Account
	account.Password = utils.GetMd5String([]byte(registerc.Password)) // md5加密
	account.Sex = registerc.Sex
	account.Sign_time = time.Now().Format("2006:01:02 15:04:05")
	fmt.Println("account信息：", account)
	lastid, err := model.InsertAccount(account)
	if err != nil {
		ReturnJson(c, 200, 103, "mysql insert error", "")
	} else {
		ReturnJson(c, 200, 200, "success", lastid)
	}
}

func LoginFunc(c *gin.Context) {
	// 获取参数，进行判断
	var loginc LoginC
	if err := c.ShouldBindQuery(&loginc); err != nil {
		ReturnJson(c, 200, 101, "params error", "")
		return
	}
	// 判断是否注册过了
	accinfo, err := model.GetAccountInfo(loginc.Account)
	if err == nil {
		if accinfo == nil {
			ReturnJson(c, 200, 102, "account is not register", "")
			return
		}
	} else {
		ReturnJson(c, 200, 103, "get account error", "")
		return
	}

	// 判断密码
	if utils.GetMd5String([]byte(loginc.Password)) != accinfo[0].Password {
		ReturnJson(c, 200, 104, "password error", "")
		return
	}

	// 登录成功写入日志
	var logininfo model.AccountLogin
	logininfo.Accid = accinfo[0].Accid
	logininfo.Login_time = time.Now().Format("2006:01:02 15:04:05")
	_, err = model.InsertLogin(logininfo)

	// 判断是否登录过了，有token
	var tokenname = "token_" + accinfo[0].Account
	//_, rerr := model.GetRedisString(tokenname)
	//if rerr != false {
	//	ReturnJson(c, 200, 105, "login repeat", "")
	//	return
	//}

	// 登录成功连接redis，生成一个token保存起来
	// 用accid和当前时间生成token
	var token = accinfo[0].Account + strconv.Itoa(int(time.Now().Unix()))
	result := model.SetRedis(tokenname, token)
	if result == false {
		fmt.Println("token 设置失败")
		ReturnJson(c, 200, 106, "login error", "")
		return
	}

	// 根据分服或者其他的，返回当前请求账号需要连接的TCP服务器信息
	var data = struct {
		Host  string
		Port  string
		Accid int
	}{
		Host:  "127.0.0.1",
		Port:  "20001",
		Accid: accinfo[0].Accid,
	}
	ReturnJson(c, 200, 200, "success", data)
}

func GetTokenFunc(c *gin.Context) {
	account := c.Query("account")
	if account == "" {
		ReturnJson(c, 200, 101, "account is null", "")
	} else {
		// 从redis查询token
		var tokenname = "token_" + account
		result, err := model.GetRedisString(tokenname)
		if err == false {
			ReturnJson(c, 200, 102, "get token error", "")
		} else {
			ReturnJson(c, 200, 200, "success", result)
		}
	}
}
