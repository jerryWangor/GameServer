package model

import (
	"fmt"
	"gameserver/utils"
	"github.com/davecgh/go-spew/spew"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var Db *sqlx.DB

// 账号表结构
type Account struct {
	Accid     int 		`db:"accid"`
	Account   string	`db:"account"`
	Password  string	`db:"password"`
	Sex       int		`db:"sex"`
	Sign_time string	`db:"sign_time"`
}

func init() {
	var err error
	Db, err = sqlx.Connect("mysql", "jerry:jerry123@tcp(42.193.50.38:3306)/golang?charset=utf8")
	if err != nil {
		// 数据库连接失败
		fmt.Println("open mysql failed,", err)
		return
	}
	Db.SetMaxOpenConns(100) // 设置数据库连接池的最大连接数
	Db.SetMaxIdleConns(50)  // 设置最大空闲连接数
}

// 获取账号信息
func GetAccountInfo(acc string) ([]Account, error) {
	var where = []string{}
	if acc != "" {
		where = append(where, spew.Sprintf("account='%s'", acc))
	}
	wheres := ""
	if len(where) > 0 {
		wheres = "where " + utils.GetWheres(where)
	}

	var account []Account
	err := Db.Select(&account, "select accid,account,password,sex,sign_time from account " + wheres)
	return account, err
}

// 注册账号，写入
func InsertAccount(accinfo Account) (int64, error) {
		conn, err := Db.Begin()
		if err != nil {
			return 0, err
		}

		// 当前时间
		r, err := conn.Exec("insert into account(accid, account, password, sex, sign_time)values(null, ?, ?, ?, ?)", accinfo.Account, accinfo.Password, accinfo.Sex, accinfo.Sign_time)
		fmt.Println(err)
		if err != nil {
			conn.Rollback()
			return 0, err
		}
		id, err := r.LastInsertId()
		if err != nil {
			conn.Rollback()
			return 0, err
		}
		conn.Commit()

		return id, nil
}
