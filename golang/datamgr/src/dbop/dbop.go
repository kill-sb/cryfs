package dbop
// todo: use map to cache db operate result

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

var useridcache map[int32]string
var usernamecache map[string]int32
var curdb *sql.DB

func init() {
	useridcache=make(map[int32]string)
	usernamecache=make(map[string]int32)
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
	query:=fmt.Sprintf("insert into efilemeta (uuid,descr,fromtype,fromobj,ownerid,hashmd5,orgname,isdir) values ('%s','%s',%d,'%s',%d,'%s','%s',%d)",pdata.Uuid,pdata.Descr,pdata.FromType,pdata.FromObj,pdata.OwnerId,pdata.HashMd5,pdata.OrgName,pdata.IsDir)
	if _, err := db.Exec(query); err != nil {
		fmt.Println("Insert encrypted data info db error:", err)
		return err
	}
	return nil
}

func UpdateMeta(pdata *core.EncryptedData) error{
	db:=GetDB()
	query:=fmt.Sprintf("update efilemeta set hashmd5='%s' where uuid='%s'",pdata.HashMd5,pdata.Uuid)
	if _, err := db.Exec(query); err != nil {
		fmt.Println("Update encrypted data info db error:", err)
		return err
	}
	return nil
}

func GetEncDataInfo(uuid string)(*core.EncryptedData,error){
	db:=GetDB()

	data:=new (core.EncryptedData)
	data.Uuid=uuid
	query:=fmt.Sprintf("select descr,fromtype,fromobj,ownerid,hashmd5,isdir,orgname,crtime from efilemeta where uuid='%s'",uuid)
	res,err:=db.Query(query)
	if err!=nil{
		fmt.Println("select from efilemeta error:",err)
		return nil,err
	}
	if res.Next(){
		err=res.Scan(&data.Descr,&data.FromType,&data.FromObj,&data.OwnerId,&data.HashMd5,&data.IsDir,&data.OrgName,&data.CrTime)
		if err!=nil{
			return nil,err
		}
		data.OwnerName,err=GetUserName(data.OwnerId)
		if err!=nil{
			return nil,err
		}
		return data,nil
	}else{
		fmt.Println("Can't find ",data.Uuid,"in db")
		return nil,errors.New("Cant find raw data in db")
	}
}

func UpdateOpenTimes(sinfo *core.ShareInfo)error{
	db:=GetDB()
	if sinfo.LeftUse<=0{
		fmt.Printf("Impossible here, while MaxUse=%d and LeftUse=%d",sinfo.MaxUse,sinfo.LeftUse)
		return errors.New("Invalid LeftTime")
	}
	sinfo.LeftUse--
	query:=fmt.Sprintf("update sharetags set leftuse=%d where uuid='%s'",sinfo.LeftUse,sinfo.Uuid)
	if _,err:=db.Exec(query);err!=nil{
		fmt.Println("Update lefttime error:",err)
		return err
	}

	return nil
}

func GetBriefShareInfo(uuid string)(*core.ShareInfo,error){
	db:=GetDB()
	query:=fmt.Sprintf("select ownerid,receivers,expire,maxuse,leftuse,datauuid,perm,fromtype,crtime, orgname from sharetags where uuid='%s'",uuid)
	res,err:=db.Query(query)
    if err!=nil{
        return nil,err
    }
    if res.Next(){
		info:=new (core.ShareInfo)
		// info.FileUri will be filled outside
		info.Uuid=uuid
		var recv string
        if err=res.Scan(&info.OwnerId, &recv,&info.Expire,&info.MaxUse,&info.LeftUse,&info.FromUuid,&info.Perm,&info.FromType,&info.CrTime,&info.OrgName);err!=nil{
			fmt.Println("query",query,"error:",err)
			return nil,err
		}
		info.OwnerName,err=GetUserName(info.OwnerId)
		if err!=nil{
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
	query:=fmt.Sprintf("select ownerid,receivers,expire,maxuse,leftuse,keycryptkey,datauuid,perm,fromtype, crtime,orgname from sharetags where uuid='%s'",uuid)
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

        if err=res.Scan(&info.OwnerId, &recv,&info.Expire,&info.MaxUse,&info.LeftUse,&randkey,&info.FromUuid,&info.Perm,&info.FromType,&info.CrTime,&info.OrgName);err!=nil{
			fmt.Println("query",query,"error:",err)
			return nil,err
		}
		info.OwnerName,err=GetUserName(info.OwnerId)
		if err!=nil{
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
	return sinfo.OrgName,nil
/*	from:=sinfo.FromUuid
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
	return "",errors.New("Can't find org filename")*/
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
	query=fmt.Sprintf("insert into sharetags (uuid,ownerid,receivers,expire,maxuse,leftuse,keycryptkey,datauuid,perm,fromtype,crtime,orgname) values ('%s',%d,'%s','%s',%d,%d,'%s','%s',%d,%d,'%s','%s')",sinfo.Uuid,sinfo.OwnerId,recvlist,sinfo.Expire,sinfo.MaxUse,sinfo.LeftUse,keystr,sinfo.FromUuid,sinfo.Perm,sinfo.FromType,sinfo.CrTime,sinfo.OrgName)
	if _, err := db.Exec(query); err != nil {
		fmt.Println("Insert shareinfo into db error:",query, err,"expire=",sinfo.Expire)
		return err
	}

	return nil
}

func GetUserName(uid int32)(string,error){
	ret,ok:=useridcache[uid]
	if ok{
		return ret,nil
	}
	db:=GetDB()
	query:=fmt.Sprintf("select name from users where id='%d'",uid)
	res,err:=db.Query(query)
	if err!=nil{
		return ret,err
	}
	if !res.Next(){
		return ret,errors.New("No such user ")
	}else{
		res.Scan(&ret)
	}
	useridcache[uid]=ret
	usernamecache[ret]=uid
	return ret,nil

}

func IsValidUser(user string)(int32,error){
	ret,ok:=usernamecache[user]
	if ok{
		return ret,nil
	}
	db:=GetDB()
	query:=fmt.Sprintf("select id from users where name='%s'",user)
	res,err:=db.Query(query)
	if err!=nil{
		return ret,err
	}
	if !res.Next(){
		return ret,errors.New("No such user ")
	}else{
		res.Scan(&ret)
	}
	usernamecache[user]=ret
	useridcache[ret]=user
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



*/
