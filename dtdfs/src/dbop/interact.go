package dbop
// todo: use map to cache db operate result

import (
	_ "MySQL"
	"database/sql"
	"errors"
	"fmt"
	"log"
	api "apiv1"
	core "coredata"
)

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
	if res!=nil{
		defer res.Close()
	}
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
	var maxcnt int32 =50
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
    if req.MaxCount!=0{
        query+=fmt.Sprintf(" limit %d,%d",req.StartItem,req.MaxCount)
		maxcnt=req.MaxCount
    }

	res,err:=db.Query(query)
	if res!=nil{
		defer res.Close()
	}
	if err!=nil{
		log.Println("select from db error:",err)
		return nil,err
	}
	ret:=make([]*api.NotifyInfo,0,maxcnt)
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
		return nil
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

func NotifyShareReq(ownerid int32,recvrs []int32, uuid string)error{
	for _,rcuid:=range recvrs{
		ni:=&api.NotifyInfo{Type:api.SHAREDATA,FromUid:ownerid,ToUid:rcuid,Content:uuid,Comment:"New Shared Data"}
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
	}else{
		epinfo.Status=api.WAITING
	}
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
	if res!=nil{
		defer res.Close()
	}
	if err!=nil{
		fmt.Println("select from exports error:",err)
		return nil,err
	}
	if res.Next(){
		err=res.Scan(&epinfo.DstData.UserId,&epinfo.Status,&epinfo.DstData.Type,&epinfo.DstData.Uuid,&epinfo.CrTime,&epinfo.Comment)
		if err!=nil{
			return nil,err
		}
	}else{
		return nil, errors.New("No such export id")
	}
	return epinfo,nil
}

func LoadProcQueue(epinfo *api.ExportProcInfo)error{
	db:=GetDB()
	query:=fmt.Sprintf("select status,procuid,comment,proctime,nodeid from exprocque where expid=%d",epinfo.ExpId)
	res,err:=db.Query(query)
	if res!=nil{
		defer res.Close()
	}
	if err!=nil{
		return err
	}
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
		if resdata!=nil{
			defer resdata.Close()
		}
		if err!=nil{
			epinfo.ProcQueue=nil
			return err
		}
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
	var maxcnt int32=50
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
    if req.Latest==1{
        query+=" order by exports.crtime desc"
    }else{
        query+=" order by exports.crtime asc"
    }
    if req.MaxCount!=0{
        query+=fmt.Sprintf(" limit %d,%d",req.StartItem,req.MaxCount)
		maxcnt=req.MaxCount
    }

	res,err:=db.Query(query)
	if res!=nil{
		defer res.Close()
	}
	if err!=nil{
		log.Println("select from db error:",err)
		return nil,err
	}
	ret:=make([]*api.ExportProcInfo,0,maxcnt)
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


