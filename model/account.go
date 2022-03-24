package model

import (
	"fmt"
	"gameserver/utils"
	"github.com/davecgh/go-spew/spew"
	"github.com/jmoiron/sqlx"
)

var Db *sqlx.DB

// 账号表结构
type Account struct {
	Accid     int
	Account   string
	Password  string
	Sex       int
	Sign_time string
}

func init() {
	Db = GetDb()
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

func GetAccountInfo(acc string) interface{} {
	var account []Account
	var where = []string{}
	if acc != "" {
		where = append(where, spew.Sprintf("account='%s'", acc))
	}
	wheres := ""
	if len(where) > 0 {
		wheres = "where " + utils.GetWheres(where)
	}

	err := Db.Select(&account, "select * from person "+wheres)
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
