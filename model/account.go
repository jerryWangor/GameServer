package model

import (
	"fmt"
	"gameserver/utils"
	"github.com/davecgh/go-spew/spew"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var Db *sqlx.DB

type Person struct {
	UserId   int    `db:"user_id"`
	Username string `db:"username"`
	Sex      string `db:"sex"`
	Email    string `db:"email"`
}

// 账号表结构
type Account struct {
	Accid     int 		`db:"accid"`
	Account   string	`db:"account"`
	Password  string	`db:"password"`
	Sex       int		`db:"sex"`
	Sign_time string	`db:"sign_time"`
}

func init() {
	Db = GetDb()
	fmt.Println("init")
}

// 获取数据库连接实例-单例
func GetDb() *sqlx.DB {
	db, err := sqlx.Open("mysql", "jerry:jerry123@tcp(42.193.50.38:3306)/golang")
	if err != nil {
		// 数据库连接失败
		fmt.Println("open mysql failed,", err)
		return db
	} else {
		return db
	}
}

// 获取账号信息
func GetAccountInfo(acc string) interface{} {
	var where = []string{}
	if acc != "" {
		where = append(where, spew.Sprintf("account='%s'", acc))
	}
	wheres := ""
	if len(where) > 0 {
		wheres = "where " + utils.GetWheres(where)
	}

	var account []Person
	err := Db.Select(&account, "select accid,account,password,sex,sign_time from person " + wheres)
	if err != nil {
		fmt.Println("exec failed, ", err)
		return ""
	}

	if acc != "" {
		return account[0]
	} else {
		return account
	}
}

// 注册账号，写入
func InsertAccount() {

}
