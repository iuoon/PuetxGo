package PtDB

/**
 * Title:mysql连接池
 * User: iuoon
 * Date: 2016-9-23
 * Version: 1.0
 */
import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

//MysqlPool mysql连接池
var MysqlPool *sql.DB

//InitMysql mysql连接池初始化函数，在程序启动时候调用
func InitMysql(user string, pwd string, host string, port int, dbs string, charset string) error {
	var err error
	connInfo := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s", user, pwd, host, port, dbs, charset)
	MysqlPool, err = sql.Open("mysql", connInfo)
	if err != nil {
		return err
	}
	MysqlPool.SetMaxOpenConns(128)
	MysqlPool.SetMaxIdleConns(64)
	MysqlPool.Ping()

	return nil
}
