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
	return nil
}

func LookupPasswdSHA(user string)(int,string,string,error){
	db:=GetDB()
	query:=fmt.Sprintf("select id,pwdsha256,enclocalkey from users where name='%s'",user)
	res,err:=db.Query(query)
	if err!=nil{
		return -1,"","",err
	}
	if res.Next(){
		var key string
		var shasum string
		var id int
		if err:=res.Scan(&id,&shasum,&key);err!=nil{
			return -1,"","",err
		}else{
			return id,shasum,key,nil
		}
	}
	return -1,"","",errors.New("No such user")
}
/*
func Lookup(year, term int) (*Info, error) {
	db := GetDB()
	var id int
	query := fmt.Sprintf("select * from records where year='%d' and term='%d'", year, term)
	res, err := db.Query(query)
	if err != nil {
		fmt.Println("Lookup in database error:", err)
		return nil, err
	}
	if res.Next() {
		info := new(Info)
		if err := res.Scan(&info.Year, &info.Term, &info.RedBalls[0], &info.RedBalls[1], &info.RedBalls[2], &info.RedBalls[3], &info.RedBalls[4], &info.RedBalls[5], &info.BlueBall, &info.Date, &id); err == nil {
			return info, nil
		} else {
			fmt.Println("Scan err", err)
		}
	}
	return nil, nil
}

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

func findLeastID() (int, error) {
	db := GetDB()
	for i := 1; ; i++ {
		query := fmt.Sprintf("select bb from mysel where id=%d", i)
		if res, err := db.Query(query); err != nil {
			return 0, err
		} else {
			if res.Next() {
				continue
			} else {
				return i, nil
			}
		}
	}
	return 0, errors.New("Impossible error")
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

func InsertSel(info *MySelInfo) {
	db := GetDB()
	var err error
	if info.Id, err = findLeastID(); err != nil {
		fmt.Println("Find id error:", err.Error())
		return
	}
	query := fmt.Sprintf("insert into mysel(id,rb1,rb2,rb3,rb4,rb5,rb6,bb,date) values ('%d','%d','%d','%d','%d','%d','%d','%d','%s')", info.Id, info.RedBalls[0], info.RedBalls[1], info.RedBalls[2], info.RedBalls[3], info.RedBalls[4], info.RedBalls[5], info.BlueBall, info.Date)
	if _, err := db.Exec(query); err != nil {
		fmt.Println("insert failed:", err.Error())
	}
}

func GetSelected() []*MySelInfo {
	db := GetDB()
	query := fmt.Sprintf("Select * from mysel order by id asc")
	if res, err := db.Query(query); err != nil {
		fmt.Println("select in db error", err.Error())
		return nil
	} else {
		list := make([]*MySelInfo, 0, 100) // should <10
		for res.Next() {
			info := new(MySelInfo)
			info.RedBalls = make([]int, 6)
			if err := res.Scan(&info.Id, &info.RedBalls[0], &info.RedBalls[1], &info.RedBalls[2], &info.RedBalls[3], &info.RedBalls[4], &info.RedBalls[5], &info.BlueBall, &info.Date); err != nil {
				fmt.Println("Scan query error in GetSeled:", err.Error())
				return nil
			} else {
				list = append(list, info)
			}
		}
		return list
	}
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

func (info *Info) AddInfo() error {
	db := GetDB()
	query := fmt.Sprintf("insert into records (year,term,rb1,rb2,rb3,rb4,rb5,rb6,bb,runtime) values ('%d','%d','%d','%d','%d','%d','%d','%d','%d','%s')", info.Year, info.Term, info.RedBalls[0], info.RedBalls[1], info.RedBalls[2], info.RedBalls[3], info.RedBalls[4], info.RedBalls[5], info.BlueBall, info.Date)
	if _, err := db.Exec(query); err != nil {
		fmt.Println("Insert into db error:", err)
		return err
	}
	return nil
}*/
