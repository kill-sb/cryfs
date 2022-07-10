package dbop

import (
    _ "MySQL"
    "database/sql"
	"os"
//    "errors"
	"sync"
	"time"
    "log"
)

var POOL_SIZE int =50   // max 20 connection in pool
const CONN_TIMEOUT  = 10*60 // 10 minutes
var ConnPool chan *ConnInst

type ConnInst struct{
	dbconn *sql.DB
	timer *time.Timer
	inuse bool
	lock *sync.Mutex
}
/*
func GetDB(conn *ConnInst)
{
	return conn.dbconn
}*/

func GetDB() *ConnInst{
	conn:=<-ConnPool
	conn.lock.Lock()
	conn.inuse=true
	conn.lock.Unlock()
	return conn
}

func PutDB(conn *ConnInst){
	conn.lock.Lock()
	conn.inuse=false
	conn.lock.Unlock()
	ConnPool<-conn
}

func Reconnect(conn *ConnInst){
	for{
		select{
		case <-conn.timer.C:
			conn.lock.Lock()
			if !conn.inuse{
				conn.dbconn.Close()
				conn.dbconn, _= sql.Open(dbname, dbconfig)
			}
			conn.lock.Unlock()
			conn.timer.Reset(time.Second*CONN_TIMEOUT)
		}
	}
}

func NewConn()(*ConnInst,error){
	conn:=new (ConnInst)
	var err error
	conn.dbconn,err=sql.Open(dbname, dbconfig)
    if err != nil {
	    log.Println("Open database error:", err)
		return nil,err
    }
	conn.lock=new (sync.Mutex)
	conn.inuse=false
	conn.timer=time.NewTimer(time.Second*CONN_TIMEOUT)
	go Reconnect(conn)
	return conn,nil
}

func InitPool(){
	ConnPool=make(chan *ConnInst,POOL_SIZE)
	for i:=0;i<POOL_SIZE;i++{
		conn,err:=NewConn()
		if err!=nil{
			log.Println("Init database error:",err)
			os.Exit(1)
		}
		conn.dbconn.SetMaxOpenConns(0)
		conn.dbconn.SetMaxIdleConns(1000)
		conn.dbconn.SetConnMaxLifetime(time.Second * 60*30)
		ConnPool<-conn
	}
}
