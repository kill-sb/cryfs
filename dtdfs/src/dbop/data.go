package dbop
// todo: use map to cache db operate result

import (
	_ "MySQL"
	"errors"
	"fmt"
	"log"
	"strings"
	api "apiv1"
	core "coredata"
)

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
	defer res.Close()
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
	defer res.Close()
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

func GetSrcObjs(rcid int64)([]*api.SourceObj,error){
	db:=GetDB()
	query:=fmt.Sprintf("select srctype,srcuuid from rcinputdata where rcid=%d",rcid)
	res,err:=db.Query(query)
	if err!=nil{
		return nil,err
	}
	defer res.Close()
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
	defer res.Close()
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
	defer res.Close()
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
	defer res.Close()
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
	defer res.Close()
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
	defer res.Close()
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
		if userid==-1{ // export data use only
			return info,nil
		}

		query=fmt.Sprintf("select leftuse from shareusers where taguuid='%s' and userid=%d", uuid,userid)
		res1,err:=db.Query(query)
		if res1!=nil{
			defer res1.Close()
		}
		if err!=nil || !res1.Next(){
			info.LeftUse=0
		}else{
			err=res1.Scan(&info.LeftUse)
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
	NotifyShareReq(sinfo.OwnerId,sinfo.RcvrIds,sinfo.Uuid);
	return nil
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
	defer res.Close()
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
	defer res.Close()
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
		defer res.Close()
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
		defer res.Close()
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
		res1,err:=db.Query(query)
		if err!=nil{
			return nil,err
		}
		defer res1.Close()
		retobj:=make([]*api.DataObj,0,10)
		for res1.Next(){
			nobj:=new (api.DataObj)
			err=res1.Scan(&nobj.Obj,&nobj.Type)
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
	defer res.Close()
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
	res1,err:=db.Query(query)
	if err!=nil{
		return nil,err
	}
	defer res1.Close()
	for res1.Next(){
		nobj:=new (api.DataObj)
		err=res1.Scan(&nobj.Obj)
		if err!=nil{
			return nil,err
		}
		nobj.Type=core.CSDFILE
		retobj=append(retobj,nobj)
	}
	return retobj,nil
}

