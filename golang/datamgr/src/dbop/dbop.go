package dbop

import (
	_ "MySQL"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"
	"strings"
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

func ParseVisitors(recvlist string) ([]string,[]int32,error){
    strret:=strings.Split(recvlist,",")
    intret:=make([]int32,0,len(strret))
    for _,user:=range strret{
        user=strings.TrimSpace(user)
        id,err:=IsValidUser(user) // should fix to asking server later
        if err!=nil{
            return nil,nil,err
        }
        intret=append(intret,id)
    }
    return strret,intret,nil
}

func SaveMeta(pdata *core.EncryptedData) error{
	db:=GetDB()
	query:=fmt.Sprintf("insert into efilemeta (uuid,descr,fromtype,fromobj,ownerid,hashmd5) values ('%s','%s','%d','%s','%d','%s')",pdata.Uuid,pdata.Descr,pdata.FromType,pdata.FromObj,pdata.OwnerId,pdata.HashMd5)
	if _, err := db.Exec(query); err != nil {
		fmt.Println("Insert encrypted data info db error:", err)
		return err
	}
	return nil
}

func GetEncDataInfo(uuid string)(*core.EncryptedData,error){
	db:=GetDB()

	data:=new (core.EncryptedData)
	query:=fmt.Sprintf("select descr,fromtype,fromobj,ownerid,hashmd5,isdir from efilemeta where uuid='%s'",uuid)
	res,err:=db.Query(query)
	if err!=nil{
		fmt.Println("select from efilemeta error:",err)
		return nil,err
	}
	if res.Next(){
		err=res.Scan(&data.Descr,&data.FromType,&data.FromObj,&data.OwnerId,&data.HashMd5,&data.IsDir)
		if err!=nil{
			return nil,err
		}
		return data,nil
	}else{
		fmt.Println("Can't find ",data.Uuid,"in db")
		return nil,errors.New("Cant find raw data in db")
	}
}

func GetBriefShareInfo(uuid string)(*core.ShareInfo,error){
	db:=GetDB()
	query:=fmt.Sprintf("select ownerid,receivers,expire,maxuse,leftuse,datauuid,perm,fromtype,crtime from sharetags where uuid='%s'",uuid)
	res,err:=db.Query(query)
    if err!=nil{
        return nil,err
    }
    if res.Next(){
		info:=new (core.ShareInfo)
		// info.FileUri will be filled outside
		info.Uuid=uuid
		var recv string
        if err=res.Scan(&info.OwnerId, &recv,&info.Expire,&info.MaxUse,&info.LeftUse,&info.FromUuid,&info.Perm,&info.FromType,&info.CrTime);err!=nil{
			fmt.Println("query",query,"error:",err)
			return nil,err
		}
		info.Receivers,info.RcvrIds,err=ParseVisitors(recv)
		if err!=nil{
			fmt.Println("Parse visitor from db error",err)
			return nil,err
		}
		return info,nil

	}else{
		return nil,errors.New("No shared info found in server")
	}


}

func LoadShareInfo(head *core.ShareInfoHeader)(*core.ShareInfo,error){
	db:=GetDB()
	uuid:=string(head.Uuid[:])
	query:=fmt.Sprintf("select ownerid,receivers,expire,maxuse,leftuse,keycryptkey,datauuid,perm,fromtype, crtime from sharetags where uuid='%s'",uuid)
   res,err:=db.Query(query)
    if err!=nil{
        return nil,err
    }
    if res.Next(){
		info:=new (core.ShareInfo)
		// info.FileUri will be filled outside
		info.Uuid=uuid
		info.IsDir=head.IsDir
		info.ContentType=int(head.ContentType)
		var recv,randkey string
		info.EncryptedKey=make([]byte,16)
		copy(info.EncryptedKey,head.EncryptedKey[:])

        if err=res.Scan(&info.OwnerId, &recv,&info.Expire,&info.MaxUse,&info.LeftUse,&randkey,&info.FromUuid,&info.Perm,&info.FromType,&info.CrTime);err!=nil{
			fmt.Println("query",query,"error:",err)
			return nil,err
		}
		info.RandKey=core.StringToBinkey(randkey)
		info.Receivers,info.RcvrIds,err=ParseVisitors(recv)
		if err!=nil{
			fmt.Println("Parse visitor from db error",err)
			return nil,err
		}
		return info,nil

	}else{
		return nil,errors.New("No shared info found in server")
	}

}

func GetOrgFileName(sinfo *core.ShareInfo)(string,error){
	from:=sinfo.FromUuid
	db:=GetDB()
	for target:=sinfo.FromType;target!=core.RAWDATA;{
		// referenced from another csd
		query:=fmt.Sprintf("select datauuid,fromtype from sharetags where uuid='%s'",from)
		res,err:=db.Query(query)
		if err!=nil{
			fmt.Println("GetOrgFileName query ",query,"error:",err)
			return "",err
		}
		if !res.Next(){
			fmt.Println("GetOrgFileName can't find uuid",from)
			return "",err
		}
		res.Scan(&from,&target)
	}
	query:=fmt.Sprintf("select fromobj from efilemeta where uuid='%s'",from)
	res,err:=db.Query(query)
	if err!=nil{
		fmt.Println("GetOrgFileName query", query,"error:",err)
		return "",err
	}
	if res.Next(){
		res.Scan(&from)
		return from,nil
	}
	return "",errors.New("Can't find org filename")
}

func WriteShareInfo(sinfo *core.ShareInfo) error{
	db:=GetDB()
	recvlist:=""
	query:=""
	for i,user:=range sinfo.Receivers{
		if recvlist!=""{
			recvlist+=","
		}
		recvlist+=user
		query=fmt.Sprintf("insert into shareusers (taguuid,userid) values ('%s',%d)",sinfo.Uuid,sinfo.RcvrIds[i])
		if _,err:=db.Exec(query);err!=nil{
			fmt.Println("Insert into shareusers error",err)
			return err
		}
	}
	recvlist=strings.TrimSpace(recvlist)
	keystr:=core.BinkeyToString(sinfo.RandKey)
	query=fmt.Sprintf("insert into sharetags (uuid,ownerid,receivers,expire,maxuse,leftuse,keycryptkey,datauuid,perm,fromtype,crtime) values ('%s',%d,'%s','%s',%d,%d,'%s','%s',%d,%d,'%s')",sinfo.Uuid,sinfo.OwnerId,recvlist,sinfo.Expire,sinfo.MaxUse,sinfo.LeftUse,keystr,sinfo.FromUuid,sinfo.Perm,sinfo.FromType,sinfo.CrTime)
	if _, err := db.Exec(query); err != nil {
		fmt.Println("Insert shareinfo into db error:",query, err,"expire=",sinfo.Expire)
		return err
	}

	return nil
}

func GetUserName(uid int32)(string,error){
	db:=GetDB()
	query:=fmt.Sprintf("select name from users where id='%d'",uid)
	res,err:=db.Query(query)
	var ret string =""
	if err!=nil{
		return ret,err
	}
	if !res.Next(){
		return ret,errors.New("No such user ")
	}else{
		res.Scan(&ret)
	}
	return ret,nil

}

func IsValidUser(user string)(int32,error){
	db:=GetDB()
	query:=fmt.Sprintf("select id from users where name='%s'",user)
	res,err:=db.Query(query)
	var ret int32=-1
	if err!=nil{
		return ret,err
	}
	if !res.Next(){
		return ret,errors.New("No such user ")
	}else{
		res.Scan(&ret)
	}
	return ret,nil
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
