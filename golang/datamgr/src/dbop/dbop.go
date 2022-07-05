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

const COMMENT_INIT="Pending..."

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
	query:=fmt.Sprintf("insert into runcontext (userid,os,baseimg,crtime,ipaddr,outputuuid,detime) values (%d,'%s','%s','%s','%s','%s','%s')",rc.UserId,rc.OS,rc.BaseImg,rc.StartTime,rc.IPAddr,rc.OutputUuid,rc.EndTime)
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
	query:=fmt.Sprintf("select userid,os,baseimg,ipaddr,outputuuid,crtime,detime from runcontext where id=%d",rcid)
	res,err:=db.Query(query)
	if err!=nil{
		return nil,err
	}
	info:=new (api.RCInfo)
	if res.Next(){
		err=res.Scan(&info.UserId,&info.OS,&info.BaseImg,&info.IPAddr,&info.OutputUuid,&info.StartTime,&info.EndTime)
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
	if pdata.OrgName==""{
		pdata.OrgName=pdata.Uuid
	}
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

func GetDataOwner(obj *api.DataObj)(int32,error){
	if obj==nil{
		return -1,errors.New("Null DataObj pointer")
	}
	tbname:=""
	var ownerid int32=-1
	switch obj.Type{
	case core.ENCDATA:
		tbname="efilemeta"
	case core.CSDFILE:
		tbname="sharetags"
	default:
		return ownerid,errors.New("Invalid Data type")
	}
	db:=GetDB()
	query:=fmt.Sprintf("select ownerid from %s where uuid='%s'",tbname,obj.Obj)
	res,err:=db.Query(query)
	if err!=nil{
		fmt.Printf("select from %s error: %s\n",tbname,err.Error())
		return ownerid,err
	}
	if res.Next(){
		err=res.Scan(&ownerid)
		if err!=nil{
			return ownerid,err
		}
		return ownerid,nil
	}
	return ownerid,errors.New("Data not found")
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
	query:=fmt.Sprintf("select sha256, ownerid,descr, receivers,expire,maxuse,datauuid,perm,fromtype, crtime,orgname,isdir from sharetags where uuid='%s'",uuid)
   res,err:=db.Query(query)
    if err!=nil{
        return nil,err
    }
	if res.Next(){
		info:=new (api.ShareInfoData)
	// info.FileUri will be filled outside
		var recv string
		info.Uuid=uuid
        if err=res.Scan(&info.Sha256,&info.OwnerId, &info.Descr,&recv,&info.Expire,&info.MaxUse,&info.FromUuid,&info.Perm,&info.FromType,&info.CrTime,&info.OrgName,&info.IsDir);err!=nil{
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
	query:=fmt.Sprintf("select sha256,ownerid,descr,receivers,expire,maxuse,keycryptkey,datauuid,perm,fromtype, crtime,orgname,isdir from sharetags where uuid='%s'",uuid)
   res,err:=db.Query(query)
    if err!=nil{
        return nil,err
    }
    if res.Next(){
		info:=new (api.ShareInfoData)
		// info.FileUri will be filled outside
		var recv string

		info.Uuid=uuid
        if err=res.Scan(&info.Sha256,&info.OwnerId, &info.Descr, &recv,&info.Expire,&info.MaxUse,&info.EncKey,&info.FromUuid,&info.Perm,&info.FromType,&info.CrTime,&info.OrgName,&info.IsDir);err!=nil{
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
	var err error
	sinfo.Receivers,err=GetUserNames(sinfo.RcvrIds)
	if err!=nil{
		return err
	}
	for i,user:=range sinfo.Receivers{
		if recvlist!=""{
			recvlist+=","
		}
		recvlist+=user
		query=fmt.Sprintf("insert into shareusers (taguuid,userid,leftuse) values ('%s',%d,%d)",sinfo.Uuid,sinfo.RcvrIds[i],sinfo.MaxUse)
		if _,err=db.Exec(query);err!=nil{
			fmt.Println("Insert into shareusers error",err)
			return err
		}
	}
	recvlist=strings.TrimSpace(recvlist)
	keystr:=sinfo.EncKey
	if sinfo.OrgName==""{
		sinfo.OrgName=sinfo.Uuid
	}
	query=fmt.Sprintf("insert into sharetags (uuid,sha256,ownerid,descr,receivers,expire,maxuse,keycryptkey,datauuid,perm,fromtype,crtime,orgname,isdir) values ('%s','%s',%d,'%s','%s','%s',%d,'%s','%s',%d,%d,'%s','%s',%d)",sinfo.Uuid,sinfo.Sha256,sinfo.OwnerId,sinfo.Descr,recvlist,sinfo.Expire,sinfo.MaxUse,keystr,sinfo.FromUuid,sinfo.Perm,sinfo.FromType,sinfo.CrTime,sinfo.OrgName,sinfo.IsDir)
	if _, err= db.Exec(query); err != nil {
		fmt.Println("Insert shareinfo into db error:",query, err,"expire=",sinfo.Expire)
		return err
	}

	return nil
}

func GetUserNames(uids []int32)([]string,error){
	n:=len(uids)
	if n<1{
		return nil,errors.New("empty user list in GetUserNames")
	}
	db:=GetDB()

	ret:=make([]string,n)
	query:=fmt.Sprintf("select name from users where id='%d'",uids[0])
	for i:=1;i<n;i++{
		query+=fmt.Sprintf(" or id='%d'",uids[i])
	}
	res,err:=db.Query(query)
	if err!=nil{
		return ret,err
	}
	i:=0
	for res.Next(){
		err=res.Scan(&ret[i])
		if err!=nil{
			return nil,err
		}
		i++
	}
	if i!=n{
		return nil,errors.New("GetUserNames error in dbop, check your id list")
	}
	return ret,nil

}

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

func SearchEncData(req *api.SearchEncDataReq)([]*api.EncDataNode,error){
	db:=GetDB()
	query:=fmt.Sprintf("select uuid, crtime from efilemeta where ownerid=%d",req.UserId)
	if req.Start!=""{
		query+=fmt.Sprintf(" and crtime >= '%s' ",req.Start+" 00:00:00")
	}
	if req.End!=""{
		query+=fmt.Sprintf(" and crtime <= '%s' ",req.End+" 23.59:59")
	}
	if req.Latest==1{
		query+=" order by crtime desc"
	}else{
		query+=" order by crtime asc"
	}
	res,err:=db.Query(query)
	if err!=nil{
		log.Println("select from db error:",err)
		return nil,err
	}
	ret:=make([]*api.EncDataNode,0,50)
	for res.Next(){
		node:=new(api.EncDataNode)
		node.UserId=req.UserId
		err=res.Scan(&node.Uuid,&node.Crtime)
		if err!=nil{
			return nil,err
		}
		ret=append(ret,node)
	}
	return ret,nil
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
	if req.Latest==1{
		query+="order by sharetags.crtime desc"
	}else{
		query+="order by sharetags.crtime asc"
	}
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

func TraceBack(obj *api.DataObj)([]*api.DataObj,error){
	if obj.Type<0{
		return nil,errors.New("wrong data type")
	}
	objmap:=make(map[string]*api.DataObj)
// ...
	queue:=make([]*api.DataObj,1,100)
	queue[0]=obj
	cur:=0
	for{
		if cur==len(queue){
			break
		}
		parents,err:=GetDataParents(queue[cur])
		if err!=nil{
			return nil,err
		}
		if parents==nil{
			cur++
			continue
		}
		for _,v:=range parents{
			if _,ok:=objmap[v.Obj];ok{
				continue // already found before
			}else{
				objmap[v.Obj]=v
				if v.Type==core.ENCDATA || v.Type==core.CSDFILE{
					queue=append(queue,v)
				}
			}
		}
		cur++
	}

	retobj:=make([]*api.DataObj,len(objmap))
	i:=0
	for _,v:=range objmap{
		retobj[i]=v
		i++
	}
	return retobj,nil
}

func GetDataParents(obj *api.DataObj)([]*api.DataObj,error){
	db:=GetDB()
	if obj.Type==core.CSDFILE{
		cobj:=new (api.DataObj)
		retobj:=make([]*api.DataObj,1)
		query:=fmt.Sprintf("select fromtype,datauuid from sharetags where uuid='%s'",obj.Obj)
		res,err:=db.Query(query)
		if err!=nil{
			return nil,err
		}
		if res.Next(){
			err=res.Scan(&cobj.Type,&cobj.Obj)
			if err!=nil{
				return nil,err
			}
			retobj[0]=cobj
			return retobj,nil
		}else{
			return nil,nil
			//return nil,errors.New("csd data not found")
		}
	}else if obj.Type==core.ENCDATA{
		//query:=fmt.Sprintf("select rcinputdata.srcuuid,rcinputdata.srctype from rcinputdata, efilemeta where efilemeta.uuid='%s' and rcinputdata.rcid=efilemeta.fromrcid", obj.Obj)
		query:=fmt.Sprintf("select fromrcid,orgname from efilemeta where uuid='%s'",obj.Obj)
		res,err:=db.Query(query)
		if err!=nil{
			return nil,err
		}
		var rcid int64
		var orgname string
		if !res.Next(){
			return nil, err
		}
		err=res.Scan(&rcid,&orgname)
		if err!=nil{
			return nil,err
		}
		if rcid<=0{ // root plain file
			cobj:=new (api.DataObj)
			retobj:=make([]*api.DataObj,1)
			cobj.Type=-1
			cobj.Obj=orgname
			retobj[0]=cobj
			return retobj,nil
		}

		query=fmt.Sprintf("select srcuuid,srctype from rcinputdata where rcid=%d", rcid)
		res,err=db.Query(query)
		if err!=nil{
			return nil,err
		}
		retobj:=make([]*api.DataObj,0,10)
		for res.Next(){
			nobj:=new (api.DataObj)
			err=res.Scan(&nobj.Obj,&nobj.Type)
			if err!=nil{
				return nil,err
			}
			retobj=append(retobj,nobj)
		}
		return retobj,nil
	}else{
		return nil,errors.New("wrong data type")
	}
}

func TraceForward(obj *api.DataObj)([]*api.DataObj,error){
	if obj.Type<0{
		return nil,errors.New("wrong data type")
	}
	objmap:=make(map[string]*api.DataObj)

	queue:=make([]*api.DataObj,1,100)
	queue[0]=obj
	cur:=0
	for{
		if cur==len(queue){
			break
		}
		children,err:=GetDataChildren(queue[cur])
		if err!=nil{
			return nil,err
		}
		if children==nil{
			cur++
			continue
		}
		for _,v:=range children{
			if _,ok:=objmap[v.Obj];ok{
				continue // already found before
			}else{
				objmap[v.Obj]=v
				if v.Type==core.ENCDATA || v.Type==core.CSDFILE{
					queue=append(queue,v)
				}else{
					fmt.Println("Find invalid obj type in dbop.TraceForwrd:",*v)
					return nil,errors.New("Invalid child data type");
				}
			}
		}
		cur++
	}

	retobj:=make([]*api.DataObj,len(objmap))
	i:=0
	for _,v:=range objmap{
		retobj[i]=v
		i++
	}
	return retobj,nil
}

func GetDataChildren(obj *api.DataObj)([]*api.DataObj,error){
	if obj.Type!=core.CSDFILE && obj.Type!=core.ENCDATA{
		return nil,errors.New("Invalid data type")
	}
	db:=GetDB()
	// search  generated encdata
	query:=fmt.Sprintf("select efilemeta.uuid from efilemeta, rcinputdata  where rcinputdata.srcuuid='%s' and rcinputdata.srctype=%d and efilemeta.fromrcid=rcinputdata.rcid",obj.Obj,obj.Type)
	res,err:=db.Query(query)
	if err!=nil{
		return nil,err
	}

	retobj:=make([]*api.DataObj,0,20)
	for res.Next(){
		nobj:=new (api.DataObj)
		err=res.Scan(&nobj.Obj)
		if err!=nil{
			fmt.Println("scan data error in dbop.GetDataChild")
			return nil,err
		}
		nobj.Type=core.ENCDATA
		retobj=append(retobj,nobj)
	}

	// search generated csd files
	query=fmt.Sprintf("select uuid from sharetags where datauuid='%s' and fromtype=%d",obj.Obj,obj.Type)
	res,err=db.Query(query)
	if err!=nil{
		return nil,err
	}
	for res.Next(){
		nobj:=new (api.DataObj)
		err=res.Scan(&nobj.Obj)
		if err!=nil{
			return nil,err
		}
		nobj.Type=core.CSDFILE
		retobj=append(retobj,nobj)
	}
	return retobj,nil
}

func NewNotify(info *api.NotifyInfo)error{
    db:=GetDB()
    query:=fmt.Sprintf("insert into notifies (type,content,descr,fromuid,touid) values (%d,'%s','%s',%d,%d)",info.Type,info.Content,info.Comment,info.FromUid,info.ToUid)
    if result, err := db.Exec(query); err == nil {
		info.Id, _ = result.LastInsertId()
		return nil
	}else{
		return err
	}
}

func SetNotifyStat(id int64, isnew int32)error{
	db:=GetDB()
	query:=fmt.Sprintf("update notifies set isnew=%d where id=%d",isnew,id)
	if _, err := db.Exec(query); err != nil {
		fmt.Println("db exec error:",query)
		return err
	}
	return nil
}

func DelNotifies(ids []int64)error{
	if len(ids)<1{
		return nil
	}
	db:=GetDB()
	query:=fmt.Sprintf("delete from notifies where id=%d",ids[0])
	for _,v:=range ids[1:]{
		query+=fmt.Sprintf(" or id=%d",v)
	}
	if _, err := db.Exec(query); err != nil {
		fmt.Println("query error:",query)
		return err
	} /*else {
		if row, _ := res.RowsAffected(); row ==l {
			full=true
		}
	}*/
	return nil
}

func GetNotifyInfo(id int64)(*api.NotifyInfo,error){
	db:=GetDB()
	ninfo:=new (api.NotifyInfo)
	ninfo.Id=id
	query:=fmt.Sprintf("select type,content,descr,crtime,fromuid,touid,isnew from notifies where id=%d",id)
	res,err:=db.Query(query)
	if err!=nil{
		fmt.Println("select from notifies error:",err)
		return nil,err
	}
	if res.Next(){
		err=res.Scan(&ninfo.Type,&ninfo.Content,&ninfo.Comment,&ninfo.CrTime,&ninfo.FromUid,&ninfo.ToUid,&ninfo.IsNew)
		if err!=nil{
			return nil,err
		}
		return ninfo,nil
	}else{
		fmt.Println("Can't find ",id,"in db")
		return nil,errors.New("Cant find notify data in db")
	}
}

func SearchNotifies(req *api.SearchNotifiesReq)([]*api.NotifyInfo,error){
	if req.FromUid==0 && req.ToUid==0{
		return nil,errors.New("'fromid' and 'toid' should be assigned at least one")
	}
	db:=GetDB()
	query:="select id,type,content,descr,crtime,fromuid,touid,isnew from notifies "
	if req.FromUid!=0 && req.ToUid!=0{
		query+=fmt.Sprintf("where fromuid=%d and touid=%d ",req.FromUid,req.ToUid)
	}else if req.FromUid!=0{
		query+=fmt.Sprintf("where  fromuid=%d ",req.FromUid)
	}else{
		query+=fmt.Sprintf("where touid=%d ",req.ToUid)
	}
	if req.Type!=0{
		query+=fmt.Sprintf(" and type=%d ",req.Type)
	}
	if req.IsNew!=-1{
		query+=fmt.Sprintf(" and isnew=%d ",req.IsNew)
	}
	if req.Latest==1{
		query+=" order by crtime desc"
	}else{
		query+=" order by crtime asc"
	}
	res,err:=db.Query(query)
	if err!=nil{
		log.Println("select from db error:",err)
		return nil,err
	}
	ret:=make([]*api.NotifyInfo,0,50)
	for res.Next(){
		node:=new(api.NotifyInfo)
		err=res.Scan(&node.Id,&node.Type,&node.Content,&node.Comment,&node.CrTime,&node.FromUid,&node.ToUid,&node.IsNew)
		if err!=nil{
			return nil,err
		}
		ret=append(ret,node)
	}
	return ret,nil
}

func RecordProcQueue(expid int64, queue []*api.ExProcNode) error{
	var err error=nil
	if len(queue)==0{
		return err
	}
	db:=GetDB()
	defer func(){
		if err!=nil{
			db.Exec(fmt.Sprintf("delete from exprocque where expid=%d",expid))
			db.Exec(fmt.Sprintf("delete from expinvolvedata where expid=%d",expid))
		}
	}()
	for _,v:=range queue{
		query:=fmt.Sprintf("insert into exprocque (expid,procuid,status,comment) values (%d,%d,%d,'%s')",expid,v.ProcUid,v.Status,COMMENT_INIT)
		var result sql.Result
	    if result, err = db.Exec(query); err == nil{
			nodeid, _:= result.LastInsertId()
			for _,usrdata:=range v.SrcData{
				_,err=db.Exec(fmt.Sprintf("insert into expinvolvedata (expid,nodeid,datauuid,datatype,dataowner) values (%d,%d,'%s',%d,%d)",expid,nodeid,usrdata.Uuid,usrdata.Type,usrdata.UserId))
				if err!=nil{
					return err
				}
			}
		}else{
			return err
		}
	}
	return nil
}

func NotifyExportReq(userid int32,expid int64, queue []*api.ExProcNode,comment *string)error{
	for _,author:=range queue{
		ni:=&api.NotifyInfo{Type:api.EXPORTDATA,FromUid:userid,ToUid:author.ProcUid,Content:fmt.Sprintf("%d",expid),Comment:*comment}
		if err:=NewNotify(ni);err!=nil{
			return err
		}
	}
	return nil
}


func CreateProcQueue(selfid int32,data *api.DataObj)([]*api.ExProcNode,error){
	sources,err:=TraceBack(data)
	if err!=nil{
		return nil,err
	}
	nodes:=make([]*api.ExProcNode,0,50)
	authors:=make(map[int32] *api.ExProcNode) // just for search author of data
	for _,obj:=range sources{
		if obj.Type!=core.ENCDATA{
			continue
		}
		var owner int32
		owner,err=GetDataOwner(obj)
		if err!=nil{
			return nil,err
		}
		if owner==selfid{
			continue
		}
		curobj:=&api.ProcDataObj{Uuid:obj.Obj,Type:obj.Type,UserId:owner}
		user,find:=authors[owner]
		if !find{
			user=new (api.ExProcNode)
			authors[owner]=user
			user.ProcUid=owner
			user.Status=api.WAITING
			user.Comment=COMMENT_INIT
			user.SrcData=make([]*api.ProcDataObj,1,20)
			user.SrcData[0]=curobj
			nodes=append(nodes,user)
		}else{
			user.SrcData=append(user.SrcData,curobj)
		}
	}
	return nodes,nil
}

func NewExport(data *api.DataObj,userid int32, comment *string)(*api.ExportProcInfo,error){
	if data==nil{
		return nil,errors.New("Empty DataObj pointer")
	}
	epinfo:=new (api.ExportProcInfo)
	epinfo.DstData=&api.ProcDataObj{Uuid:data.Obj,Type:data.Type,UserId:userid}
	var err error
	epinfo.ProcQueue,err=CreateProcQueue(userid,data)
	if err!=nil || epinfo.ProcQueue==nil{
		return nil,err
	}
	epinfo.Comment=*comment
	epinfo.CrTime=core.GetCurTime()
	if len(epinfo.ProcQueue)==0{ // from raw data or from data ownered by self
		epinfo.Status=api.AGREE
	}
	epinfo.Status=api.WAITING
	db:=GetDB()
	query:=fmt.Sprintf("insert into exports (requid,status,datatype,datauuid,crtime,comment) values (%d,%d,%d,'%s','%s','%s')",userid,epinfo.Status,epinfo.DstData.Type,epinfo.DstData.Uuid,epinfo.CrTime,epinfo.Comment)
    if result, err := db.Exec(query); err == nil{
		epinfo.ExpId, _ = result.LastInsertId()
		if err=RecordProcQueue(epinfo.ExpId,epinfo.ProcQueue);err!=nil{
			db.Exec(fmt.Sprintf("delete from exports where expid=%d",epinfo.ExpId))
			return nil,err
		}
		NotifyExportReq(userid, epinfo.ExpId,epinfo.ProcQueue,comment)
		return epinfo,nil
	}else{
		return nil,err
	}
}

func GetExportInfo(expid int64)(*api.ExportProcInfo,error){
	epinfo:=new (api.ExportProcInfo)
	epinfo.DstData=new (api.ProcDataObj)
	epinfo.ExpId=expid
	db:=GetDB()
	query:=fmt.Sprintf("select requid,status,datatype,datauuid,crtime,comment from exports where expid=%d",expid)
	res,err:=db.Query(query)
	if err!=nil{
		fmt.Println("select from exports error:",err)
		return nil,err
	}
	if res.Next(){
		err=res.Scan(&epinfo.DstData.UserId,&epinfo.Status,&epinfo.DstData.Type,&epinfo.DstData.Uuid,&epinfo.CrTime,&epinfo.Comment)
		if err!=nil{
			return nil,err
		}
	}
	return epinfo,nil
}

func LoadProcQueue(epinfo *api.ExportProcInfo)error{
	db:=GetDB()
	query:=fmt.Sprintf("select status,procuid,comment,proctime,nodeid from exprocque where expid=%d",epinfo.ExpId)
	res,err:=db.Query(query)
	epinfo.ProcQueue=make([]*api.ExProcNode,0,50)
	for res.Next(){
		node:=new(api.ExProcNode)
		var nodeid int64
		err=res.Scan(&node.Status,&node.ProcUid,&node.Comment,&node.ProcTime,&nodeid)
		if err!=nil{
			epinfo.ProcQueue=nil
			return err
		}
		node.SrcData=make([]*api.ProcDataObj,0,20)
		query:=fmt.Sprintf("select datauuid,datatype,dataowner from expinvolvedata where nodeid=%d",nodeid)
		resdata,err:=db.Query(query)
		for resdata.Next(){
			srcnode:=new (api.ProcDataObj)
			err=resdata.Scan(&srcnode.Uuid,&srcnode.Type,&srcnode.UserId)
			if err!=nil{
				epinfo.ProcQueue=nil
				return err
			}
			node.SrcData=append(node.SrcData,srcnode)
		}
		epinfo.ProcQueue=append(epinfo.ProcQueue,node)
	}
	return nil
}

func SearchExpProc(req *api.SearchExpReq)([]*api.ExportProcInfo,error){
	if req.FromUid<=0 && req.ToUid<=0{
		return nil,errors.New("'fromid' and 'toid' should be assigned at least one")
	}
	db:=GetDB()
	query:=""
	if req.ToUid>0{
		query=fmt.Sprintf("select exports.expid, exports.requid,exports.status,exports.datatype,exports.datauuid,exports.crtime,exports.comment from exports,exprocque where (exprocque.procuid=%d and exprocque.expid=exports.expid) ", req.ToUid)
		if req.FromUid>0{
			query+=fmt.Sprintf("and exports.requid=%d ",req.FromUid)
		}
	}else{ //no ToUid, FromUid must be >0
		query=fmt.Sprintf("select exports.expid,exports.requid,exports.status,exports.datatype,exports.datauuid,exports.crtime,exports.comment from exports where exports.requid=%d ",req.FromUid)
	}
	if req.Status!=0{
		query+=fmt.Sprintf(" and exports.Status=%d ",req.Status)
	}
    if req.Start!=""{
		query+=fmt.Sprintf(" and exports.crtime >= '%s' ",req.Start+" 00:00:00")
	}
	if req.End!=""{
		query+=fmt.Sprintf(" and exports.crtime <= '%s' ",req.End+" 23.59:59")
	}
	res,err:=db.Query(query)
	if err!=nil{
		log.Println("select from db error:",err)
		return nil,err
	}
	ret:=make([]*api.ExportProcInfo,0,50)
	for res.Next(){
		info:=new(api.ExportProcInfo)
// exports.expid,exports.requid,exports.status,exports.datatype,exports.datauuid,exports.crtime 
		info.DstData=new (api.ProcDataObj)
		err=res.Scan(&info.ExpId,&info.DstData.UserId,&info.Status,&info.DstData.Type,&info.DstData.Uuid,&info.CrTime,&info.Comment)
		if err!=nil{
			return nil,err
		}
		if err=LoadProcQueue(info);err!=nil{
			return nil,err
		}
		ret=append(ret,info)
	}
	return ret,nil
}

func RespExportReq(uid int32,req *api.RespExpReq) error{
	info,err:=GetExportInfo(req.ExpId)
	if err!=nil{
		return err
	}
	err=LoadProcQueue(info)
	if err!=nil{
		return err
	}
	var node *api.ExProcNode=nil
	nlist:=len(info.ProcQueue)
	refuse:=false
	agree:=0
	for i:=0;i<nlist;i++{
		if info.ProcQueue[i].ProcUid==uid{
			if info.ProcQueue[i].Status!=api.WAITING{
				return errors.New("The request is processed already")
			}
			node=info.ProcQueue[i]
		}else if info.ProcQueue[i].Status==api.AGREE{
			agree++
		}else if info.ProcQueue[i].Status==api.REFUSE{
			refuse=true
		}
	}
	if node==nil{
		return errors.New("User not in author list")
	}

	db:=GetDB()
	query:=fmt.Sprintf("update exprocque set status=%d,comment='%s',proctime='%s' where expid=%d and procuid=%d", req.Status,req.Comment,core.GetCurTime(), req.ExpId, uid)
	if _,err=db.Exec(query);err!=nil{
		return err
	}

	if (!refuse && agree==nlist-1 && req.Status!=api.WAITING ) || req.Status==api.REFUSE{
		query=fmt.Sprintf("update exports set status=%d where expid=%d",req.Status,req.ExpId)
		_,err=db.Exec(query)
		if err!=nil{
			return err
		}
	}
	return nil
}

func NewContact(uid, cid int32)error{
    db:=GetDB()
	query:=fmt.Sprintf("select count(*) from contacts where userid=%d and contactuserid=%d",uid,cid)
	res,err:=db.Query(query)
    if err!=nil{
        fmt.Println("select from contacts error:",err)
        return err
    }
    if res.Next(){
        var count int64
        err=res.Scan(&count)
        if err!=nil{
            return err
        }
		if count>0{
			return nil //errors.New("contact exists already")
		}
	}
    query=fmt.Sprintf("insert into contacts (userid, contactuserid) values (%d,%d)",uid,cid)
    if _, err= db.Exec(query); err == nil {
		return nil
	}else{
		return err
	}
}

func ListContacts(uid int32)([]*api.ContactInfo,error){
	db:=GetDB()
	query:=fmt.Sprintf("select contacts.contactuserid,users.name from contacts,users where contacts.userid=%d and users.id=contacts.contactuserid",uid)
	res, err := db.Query(query)
	if err != nil {
		fmt.Println("query contacts error:",query)
		return nil,err
	}
	clist:=make([]*api.ContactInfo,0,50)
	for res.Next(){
		cinfo:=new(api.ContactInfo)
		err=res.Scan(&cinfo.UserId,&cinfo.Name)
		if err!=nil{
			return nil,err
		}
		clist=append(clist,cinfo)
	}
	return clist,nil
}

func FuzzySearch(uid int32, keyword string)([]*api.ContactInfo,error){
	db:=GetDB()
	query:=fmt.Sprintf("select contacts.contactuserid,users.name from contacts,users where contacts.userid=%d and contacts.contactuserid=users.id and users.name like '%s'",uid,"%"+keyword+"%")
	res,err:=db.Query(query)
	if err!=nil{
		fmt.Println("query error:",query)
		return nil,err
	}
	clist:=make([]*api.ContactInfo,0,50)
	for res.Next(){
		cinfo:=new(api.ContactInfo)
		err=res.Scan(&cinfo.UserId,&cinfo.Name)
		if err!=nil{
			return nil,err
		}
		clist=append(clist,cinfo)
	}
	return clist,nil
}

func DelContact(uid,cid int32 )error{
	db:=GetDB()
	query:=fmt.Sprintf("delete from contacts where userid=%d and contactuserid=%d",uid,cid)
	if _,err:= db.Exec(query);err != nil{
		fmt.Println("DelContact error:",query)
		return err
	}
	return nil
}

