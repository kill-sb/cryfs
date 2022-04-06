package dbop
// todo: use map to cache db operate result

import (
	_ "MySQL"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"
	"strings"
	api "apiv1"
	core "coredata"
)

//var useridcache map[int32]string
var usernamecache map[string] *api.UserInfoData
var userinfocache map[int32] *api.UserInfoData
var curdb *sql.DB

func init() {
	//useridcache=make(map[int32]string)
	userinfocache=make(map[int32] *api.UserInfoData)
	usernamecache=make(map[string] *api.UserInfoData)
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

func NewRunContext(rc *api.RCInfo) error{
	db:=GetDB()
	query:=fmt.Sprintf("insert into runcontext (userid,os,baseimg,crtime) values (%d,'%s','%s','%s')",rc.UserId,rc.OS,rc.BaseImg,rc.StartTime)
	if result, err := db.Exec(query); err == nil {
        rc.RCId, _ = result.LastInsertId()
		// create other info in rcinputdata & rcimport
		for _,data:=range rc.InputData{
			query=fmt.Sprintf("insert into rcinputdata (rcid, srcuuid, srctype) values (%d,'%s',%d)",rc.RCId, data.DataUuid,data.DataType)
			_,err=db.Exec(query)
			if err!=nil{
				return err
			}
		}
		for _,tool:=range rc.ImportPlain{
			query=fmt.Sprintf("insert into rcimport (rcid,relname,filedesc,sha256,size) values (%d,'%s','%s','%s',%d)",rc.RCId,tool.RelName,tool.FileDesc,tool.Sha256,tool.Size)
			_,err=db.Exec(query)
			if err!=nil{
				return err
			}
		}
		return nil
	}else {
		return err
	}
}

func UpdateRunContext(userid int32, rcid int64, datauuid string, endtime string) error{
	db:=GetDB()
	// TODO : check userid in RCInfo from rcid
	query:=fmt.Sprintf("select userid from runcontext where id=%d",rcid)
	res,err:=db.Query(query)
    if err!=nil{
        fmt.Println("select from runcontext error:",err)
        return err
    }
    if res.Next(){
		var uid int32
        err=res.Scan(&uid)
        if err!=nil{
            return err
        }
        if uid!=userid{
			return errors.New("invalid user")
		}
    }
	query=fmt.Sprintf("update runcontext set outputuuid='%s', detime='%s' where id=%d",datauuid, endtime,rcid)
	_,err=db.Exec(query)
	return err
}

func GetRCInfo(rcid int64)(*api.RCInfo,error){
	db:=GetDB()
	query:=fmt.Sprintf("select userid,os,baseimg,outputuuid,crtime,detime from runcontext where id=%d",rcid)
	res,err:=db.Query(query)
	if err!=nil{
		return nil,err
	}
	info:=new (api.RCInfo)
	if res.Next(){
		err=res.Scan(&info.UserId,&info.OS,&info.BaseImg,&info.OutputUuid,&info.StartTime,&info.EndTime)
		if err!=nil{
			return nil,err
		}
		info.InputData,err=GetSrcObjs(rcid)
		if err!=nil{
			return nil,err
		}
		info.ImportPlain,err=GetImportInfo(rcid)
		if err!=nil{
			return nil,err
		}
		info.RCId=rcid
		return info,nil
	}
	return nil,errors.New("runcontext id not found")
}

func SaveEncMeta(pdata *api.EncDataReq) error{
	db:=GetDB()
	query:=fmt.Sprintf("insert into efilemeta (uuid,descr,fromrcid,ownerid,orgname,isdir) values ('%s','%s',%d, %d,'%s',%d)",pdata.Uuid,pdata.Descr,pdata.FromRCId,pdata.OwnerId,pdata.OrgName,pdata.IsDir)
	if _, err := db.Exec(query); err != nil {
		fmt.Println("Insert encrypted data info db error:", err)
		return err
	}
	return nil
}
/*
func UpdateMeta(pdata *api.UpdateDataInfoReq) error{
	db:=GetDB()
	query:=fmt.Sprintf("update efilemeta set hashmd5='%s' where uuid='%s'",pdata.Hash256,pdata.Uuid)
	if _, err := db.Exec(query); err != nil {
		fmt.Println("Update encrypted data info db error:", err)
		return err
	}
	return nil
}*/

func GetSrcObjs(rcid int64)([]*api.SourceObj,error){
	db:=GetDB()
	query:=fmt.Sprintf("select srctype,srcuuid from rcinputdata where rcid=%d",rcid)
	res,err:=db.Query(query)
	if err!=nil{
		return nil,err
	}
	objs:=make([]*api.SourceObj,0,10)
	for res.Next(){
		obj:=new(api.SourceObj)
		err=res.Scan(&obj.DataType,&obj.DataUuid)
		if err!=nil{
			return nil,err
		}
		objs=append(objs,obj)
	}
	return objs,nil
}

func GetImportInfo(rcid int64)([]*api.ImportFile,error){
    db:=GetDB()
    query:=fmt.Sprintf("select relname,filedesc,sha256,size from rcimport where rcid=%d",rcid)
    res,err:=db.Query(query)
    if err!=nil{
        return nil,err
    }
    objs:=make([]*api.ImportFile,0,10)
    for res.Next(){
        obj:=new(api.ImportFile)
        err=res.Scan(&obj.RelName,&obj.FileDesc,&obj.Sha256,&obj.Size)
        if err!=nil{
            return nil,err
        }
        objs=append(objs,obj)
    }
    return objs,nil
}

func GetEncDataInfo(uuid string)(*api.EncDataInfo,error){
	db:=GetDB()
	data:=new (api.EncDataInfo)
	data.Uuid=uuid
	query:=fmt.Sprintf("select descr,fromrcid,ownerid,isdir,orgname,crtime from efilemeta where uuid='%s'",uuid)
	res,err:=db.Query(query)
	if err!=nil{
		fmt.Println("select from efilemeta error:",err)
		return nil,err
	}
	if res.Next(){
		err=res.Scan(&data.Descr,&data.FromRCId,&data.OwnerId,&data.IsDir,&data.OrgName,&data.CrTime)
		if err!=nil{
			return nil,err
		}
		// search srcobj
		data.SrcObj,err=GetSrcObjs(data.FromRCId)
		if err!=nil{
			return nil,err
		}
		return data,nil
	}else{
		fmt.Println("Can't find ",data.Uuid,"in db")
		return nil,errors.New("Cant find raw data in db")
	}
}

func DecreaseOpenTimes(sinfo *api.ShareInfoData, userid int32) error{
	db:=GetDB()
	if sinfo.LeftUse<=0{
		fmt.Printf("Impossible here, while MaxUse=%d and LeftUse=%d",sinfo.MaxUse,sinfo.LeftUse)
		return errors.New("Invalid LeftTime")
	}
	sinfo.LeftUse--
	query:=fmt.Sprintf("update shareusers set leftuse=%d where taguuid='%s' and userid=%d",sinfo.LeftUse,sinfo.Uuid,userid)
	if _,err:=db.Exec(query);err!=nil{
		fmt.Println("Update lefttime error:",err)
		return err
	}
	return nil
}

func GetShareInfoData(uuid string)(*api.ShareInfoData,error){
	db:=GetDB()
	query:=fmt.Sprintf("select ownerid,descr, receivers,expire,maxuse,datauuid,perm,fromtype, crtime,orgname,isdir from sharetags where uuid='%s'",uuid)
   res,err:=db.Query(query)
    if err!=nil{
        return nil,err
    }
	if res.Next(){
		info:=new (api.ShareInfoData)
	// info.FileUri will be filled outside
		var recv string
		info.Uuid=uuid
        if err=res.Scan(&info.OwnerId, &info.Descr,&recv,&info.Expire,&info.MaxUse,&info.FromUuid,&info.Perm,&info.FromType,&info.CrTime,&info.OrgName,&info.IsDir);err!=nil{
			fmt.Println("query",query,"error:",err)
			return nil,err
		}
		info.Receivers,info.RcvrIds,err=ParseVisitors(recv)
		if err!=nil{
			fmt.Println("Parse visitor from db error",err)
			return nil,err
		}
		info.EncKey=""
		info.LeftUse=0
		return info,nil
	}else{
		return nil,errors.New("No shared info found in server")
	}
}

func GetUserShareInfoData(uuid string, userid int32)(*api.ShareInfoData,error){
	db:=GetDB()
	query:=fmt.Sprintf("select ownerid,descr,receivers,expire,maxuse,keycryptkey,datauuid,perm,fromtype, crtime,orgname,isdir from sharetags where uuid='%s'",uuid)
   res,err:=db.Query(query)
    if err!=nil{
        return nil,err
    }
    if res.Next(){
		info:=new (api.ShareInfoData)
		// info.FileUri will be filled outside
		var recv string

		info.Uuid=uuid
        if err=res.Scan(&info.OwnerId, &info.Descr, &recv,&info.Expire,&info.MaxUse,&info.EncKey,&info.FromUuid,&info.Perm,&info.FromType,&info.CrTime,&info.OrgName,&info.IsDir);err!=nil{
			fmt.Println("query",query,"error:",err)
			return nil,err
		}
		info.Receivers,info.RcvrIds,err=ParseVisitors(recv)
		if err!=nil{
			fmt.Println("Parse visitor from db error",err)
			return nil,err
		}
		query=fmt.Sprintf("select leftuse from shareusers where taguuid='%s' and userid=%d", uuid,userid)
		res,err=db.Query(query)
		if err!=nil || !res.Next(){
			info.LeftUse=0
		}else{
			err=res.Scan(&info.LeftUse)
			if err!=nil{
				return nil,err
			}
		}
		return info,nil
	}else{
		return nil,errors.New("No shared info found in server")
	}
}

func GetOrgFileName(sinfo *core.ShareInfo)(string,error){
	return sinfo.OrgName,nil
}

func WriteShareInfo(sinfo *api.ShareInfoData) error{
	db:=GetDB()
	recvlist:=""
	query:=""
	for i,user:=range sinfo.Receivers{
		if recvlist!=""{
			recvlist+=","
		}
		recvlist+=user
		query=fmt.Sprintf("insert into shareusers (taguuid,userid,leftuse) values ('%s',%d,%d)",sinfo.Uuid,sinfo.RcvrIds[i],sinfo.MaxUse)
		if _,err:=db.Exec(query);err!=nil{
			fmt.Println("Insert into shareusers error",err)
			return err
		}
	}
	recvlist=strings.TrimSpace(recvlist)
	keystr:=sinfo.EncKey
	query=fmt.Sprintf("insert into sharetags (uuid,ownerid,descr,receivers,expire,maxuse,keycryptkey,datauuid,perm,fromtype,crtime,orgname,isdir) values ('%s',%d,'%s','%s','%s',%d,'%s','%s',%d,%d,'%s','%s',%d)",sinfo.Uuid,sinfo.OwnerId,sinfo.Descr,recvlist,sinfo.Expire,sinfo.MaxUse,keystr,sinfo.FromUuid,sinfo.Perm,sinfo.FromType,sinfo.CrTime,sinfo.OrgName,sinfo.IsDir)
	if _, err := db.Exec(query); err != nil {
		fmt.Println("Insert shareinfo into db error:",query, err,"expire=",sinfo.Expire)
		return err
	}

	return nil
}
/*
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
*/
func IsValidUser(user string)(int32,error){
	var ret int32 =-1
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
	return ret,nil
}

func GetUserInfoByName(name string)(*api.UserInfoData,error){
	ret,ok:=usernamecache[name]
	if ok{
		return ret,nil
	}
	ret=new (api.UserInfoData)
	ret.Name=name
	db:=GetDB()
	query:=fmt.Sprintf("select descr,id,mobile,email from users where name='%s'",name)
	res,err:=db.Query(query)
	if err!=nil{
		fmt.Println("Query error:",err)
		return ret,err
	}
	if !res.Next(){
		ret.Id=-1
		fmt.Println("error",err)
		return ret,nil
	}else{
		res.Scan(&ret.Descr,&ret.Id,&ret.Mobile,&ret.Email)
	}
	usernamecache[name]=ret
	return ret,nil
}

func GetUserInfo(id int32)(*api.UserInfoData,error){
	ret,ok:=userinfocache[id]
	if ok{
		return ret,nil
	}
	ret=new (api.UserInfoData)
	ret.Id=id
	db:=GetDB()
	query:=fmt.Sprintf("select descr,name,mobile,email from users where id=%d",id)
	res,err:=db.Query(query)
	if err!=nil{
		return ret,err
	}
	if !res.Next(){
		return ret,errors.New("No such user ")
	}else{
		res.Scan(&ret.Descr,&ret.Name,&ret.Mobile,&ret.Email)
	}
	userinfocache[id]=ret
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

func SearchShareData(req *api.SearchShareDataReq)([]*api.ShareDataNode,error){
	if req.FromId<=0 && req.ToId<=0{
		return nil,errors.New("'fromid' and 'toid' should be assigned at least one")
	}
	db:=GetDB()
	query:="select sharetags.uuid, sharetags.ownerid,sharetags.crtime, shareusers.userid, shareusers.leftuse from sharetags,shareusers "
	if req.FromId>0 && req.ToId>0{
		query+=fmt.Sprintf("where sharetags.ownerid=%d and shareusers.userid=%d ",req.FromId,req.ToId)
	}else if req.FromId>0{
		query+=fmt.Sprintf("where sharetags.ownerid=%d ",req.FromId)
	}else{
		query+=fmt.Sprintf("where shareusers.userid=%d ",req.ToId)
	}
	query+=" and sharetags.uuid=shareusers.taguuid "
	if req.Start!=""{
		query+=fmt.Sprintf("and sharetags.crtime >= '%s' ",req.Start+" 00:00:00")
	}
	if req.End!=""{
		query+=fmt.Sprintf("and sharetags.crtime <= '%s' ",req.End+" 23.59:59")
	}
	log.Println("search query:",query)
	res,err:=db.Query(query)
	if err!=nil{
		log.Println("select from db error:",err)
		return nil,err
	}
	ret:=make([]*api.ShareDataNode,0,50)
	for res.Next(){
		node:=new(api.ShareDataNode)
		err=res.Scan(&node.Uuid,&node.FromId,&node.Crtime,&node.ToId,&node.LeftTimes)
		if err!=nil{
			return nil,err
		}
		ret=append(ret,node)
	}
	return ret,nil
}

func SingleTrace(obj *api.DataObj)([]*api.DataObj,error){
	// TODO : need to be implemented with multi source supported
//	db:=GetDB()
	if obj.Type<0{
		return nil,errors.New("wrong data type")
	}
	retobj:=make([]*api.DataObj,0,10) // current only 1 parent, later will be multiple
/*	curtype:=obj.Type
	curobj:=obj.Obj
	ptype:=-1
	i:=0
	for {
	var query string
	if curtype==core.RAWDATA{
		query=fmt.Sprintf("select fromtype,fromobj from efilemeta where uuid='%s'",curobj)
	}else if curtype==core.CSDFILE{
		query=fmt.Sprintf("select fromtype,datauuid from sharetags where uuid='%s'",curobj)
	}
	res,err:=db.Query(query)
	if err!=nil{
		log.Println("select from db error:",err)
		return nil,err
	}
	if res.Next(){
		newobj:=new (api.DataObj)
		err=res.Scan(&newobj.Type,&newobj.Obj)
		if err!=nil{
			return nil,err
		}
		retobj=append(retobj,newobj)
		ptype=newobj.Type
		i++
		if curtype==core.RAWDATA && ptype==core.RAWDATA{
			newobj.Type=-1
			break
		}
		curtype=newobj.Type
		curobj=newobj.Obj
	}else{
		log.Println("Can't find ",obj.Obj,"in db")
		return nil,errors.New(fmt.Sprintf("Cant find %s with type %d in db",curobj,curtype))
	}
	}*/
	return retobj,nil
}

func GetDataParent(obj *api.DataObj)([]api.DataObj,error){
	// TODO: need to be reimplemented
	return nil,nil
/*
	db:=GetDB()
	if obj.Type<0{
		return nil,errors.New("wrong data type")
	}
	retobj:=make([]api.DataObj,1,10) // current only 1 parent, later will be multiple
	var query string
	if obj.Type==core.RAWDATA{
		query=fmt.Sprintf("select fromtype,fromobj from efilemeta where uuid='%s'",obj.Obj)
	}else if obj.Type==core.CSDFILE{
		query=fmt.Sprintf("select fromtype,datauuid from sharetags where uuid='%s'",obj.Obj)
	}
	res,err:=db.Query(query)
	if err!=nil{
		fmt.Println("select from db error:",err)
		return nil,err
	}
	if res.Next(){
		err=res.Scan(&retobj[0].Type,&retobj[0].Obj)
		if err!=nil{
			return nil,err
		}
		return retobj,nil
	}else{
		fmt.Println("Can't find ",obj.Obj,"in db")
		return nil,errors.New("Cant find"+obj.Obj+"data in db")
	}
	*/
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
