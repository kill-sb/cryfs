package dbop
// todo: use map to cache db operate result

import (
	_ "MySQL"
	"database/sql"
	"time"
	"os"
	"fmt"
	api "apiv1"
)

//var useridcache map[int32]string
var usernamecache map[string] *api.UserInfoData
var userinfocache map[int32] *api.UserInfoData
var curdb *sql.DB =nil

const dbname="mysql"
const dbconfig="cmit:123456@tcp(mysqlsvr:3306)/cmit"
const COMMENT_INIT="Pending..."
const CHECK_TIME=60*2

func init() {
	//useridcache=make(map[int32]string)
	userinfocache=make(map[int32] *api.UserInfoData)
	usernamecache=make(map[string] *api.UserInfoData)
	ConnDB()
}

func CheckConnect(timer *time.Timer){
   for{
        select{
        case <-timer.C:
			err:=curdb.Ping()
            if err!=nil{
                curdb.Close()
                curdb, _= sql.Open(dbname, dbconfig)
                timer.Reset(time.Second*CHECK_TIME)
            }
        }
    }
}

func ConnDB() {
	var err error
	//curdb, err = sql.Open("mysql", "cmit:123456@tcp(mysqlsvr:3306)/cmit")
	if curdb==nil{
		curdb, err = sql.Open(dbname, dbconfig)
		if err != nil {
			fmt.Println("Open database error:", err)
			os.Exit(1)
		}
		curdb.SetMaxOpenConns(0)
		curdb.SetMaxIdleConns(1000)
		curdb.SetConnMaxLifetime(time.Second *5)
		go CheckConnect(time.NewTimer(time.Second*CHECK_TIME))
	}
}

func GetDB() *sql.DB {
/*	if err := curdb.Ping(); err != nil {
		curdb.Close()
		ConnDB()
	}*/
	if curdb==nil{
		ConnDB()
	}
	return curdb
}

