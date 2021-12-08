package dbop

import (
	_ "MySQL"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"
	core "coredata"
)

var curdb *sql.DB

func init() {
	ConnDB()
}

func ConnDB() {
	var err error
	curdb, err = sql.Open("mysql", "cmit:123456@tcp(mysqlsvr:3306)/cmit")
	if err != nil {
		fmt.Println("Open database error:", err)
		os.Exit(1)
	}
	curdb.SetConnMaxLifetime(time.Second * 500)
}

func GetDB() *sql.DB {
	if err := curdb.Ping(); err != nil {
		curdb.Close()
		ConnDB()
	}
	return curdb
}

func SaveMeta(pdata *core.EncryptedData) error{
	db:=GetDB()
	query:=fmt.Sprintf("insert into efilemeta (uuid,descr,fromtype,fromobj,ownerid,hashmd5) values ('%s','%s','%d','%s','%d','%s')",pdata.Uuid,pdata.Descr,pdata.FromType,pdata.FromObj,pdata.OwnerId,pdata.HashMd5)
	if _, err := db.Exec(query); err != nil {
		fmt.Println("Insert into db error:", err)
		return err
	}
	return nil
}

func LookupPasswdSHA(user string)(int32,string,string,error){
	db:=GetDB()
	query:=fmt.Sprintf("select id,pwdsha256,enclocalkey from users where name='%s'",user)
	res,err:=db.Query(query)
	if err!=nil{
		return -1,"","",err
	}
	if res.Next(){
		var key string
		var shasum string
		var id int32
		if err:=res.Scan(&id,&shasum,&key);err!=nil{
			return -1,"","",err
		}else{
			return id,shasum,key,nil
		}
	}
	return -1,"","",errors.New("No such user")
}
/*

func DelSel(id int) bool {
	db := GetDB()
	query := fmt.Sprintf("delete from mysel where id=%d", id)
	if res, err := db.Exec(query); err != nil {
		fmt.Println("db exec error:", err.Error())
	} else {
		if row, _ := res.RowsAffected(); row > 0 {
			return true
		}
	}
	return false
}


func (info *MySelInfo) UpdateInfo() (bool, error) {
	db := GetDB()
	ret := false
	query := fmt.Sprintf("update mysel set rb1=%d,rb2=%d,rb3=%d,rb4=%d,rb5=%d,rb6=%d,bb=%d,date='%s' where id=%d", info.RedBalls[0], info.RedBalls[1], info.RedBalls[2], info.RedBalls[3], info.RedBalls[4], info.RedBalls[5], info.BlueBall, info.Date, info.Id)

	if res, err := db.Exec(query); err != nil {
		return false, err
	} else if rows, _ := res.RowsAffected(); rows > 0 {
		ret = true
	}
	return ret, nil
}

func EnumAll(startyear int, limit int64, proc func(info *Info)) {
    db := GetDB()
    query := fmt.Sprintf("select count(*) as value from records")
    var rows int64
    if err := db.QueryRow(query).Scan(&rows); err != nil {
        fmt.Println("Query rows error")
        return
    }
    if rows < limit || limit < 0 {
        limit = rows
    }

    query = fmt.Sprintf("select * from records where year>=%d limit %d,%d", startyear, rows-limit, limit)
    if res, err := db.Query(query); err != nil {
        fmt.Println("slect in db error")
        return
    } else {
        info := new(Info)
        var id int
        for res.Next() {
            if err := res.Scan(&info.Year, &info.Term, &info.RedBalls[0], &info.RedBalls[1], &info.RedBalls[2], &info.RedBalls[3], &info.RedBalls[4], &info.RedBalls[5], &info.BlueBall, &info.Date, &id); err == nil {
                proc(info)
            } else {
                fmt.Println("Scan query data error", err)
                break
            }
        }
    }
}

*/
