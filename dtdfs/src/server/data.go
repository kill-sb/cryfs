package main

import (
	"net/http"
	"encoding/json"
	"log"
	"time"
	"strings"
	core "coredata"
//	"os"
	"dbop"
	api "apiv1"
)

func NewDataFunc(w http.ResponseWriter, r *http.Request){
	if r.Method=="POST"{
		encack:=api.NewSimpleAck()
		w.Header().Set("Content-Type","application/json")
		var encreq api.EncDataReq
		err:=json.NewDecoder(r.Body).Decode(&encreq)
		if err!=nil{
			Debug("Decode json error:",err)
			encack.Code=api.ERR_BADPARAM
			json.NewEncoder(w).Encode(encack)
			return
		}
        if g_config.Debug{
            DebugJson("Request:",&encreq)
            defer DebugJson("Response:",encack)
        }
		luinfo,err:=GetLoginUserInfo(encreq.Token)
		if err!=nil{
			encack.Code=api.ERR_INVDATA
			encack.Msg=err.Error()
			json.NewEncoder(w).Encode(encack)
			return
		}
		if luinfo.Id!=encreq.OwnerId{
			encack.Code=api.ERR_ACCESS
			encack.Msg="Invalid user"
			json.NewEncoder(w).Encode(encack)
			return
		}
		// user info checked ok
		// reference crypt.go:dbop.SaveMeta
		log.Println("newdata:",encreq)
/*		pdata:=new(core.EncryptedData)
	    pdata.Uuid=encreq.Uuid
	    pdata.Descr=encreq.Descr
	    pdata.FromRCId=encreq.FromRCId // TODO FromContext should be created and filled in dbop.SaveMeta
	    pdata.OrgName=encreq.OrgName
	    pdata.OwnerId=encreq.OwnerId
	    pdata.EncryptingKey=nil //""core.StringToBinkey(encreq.EncKey)
	    pdata.IsDir=encreq.IsDir
		*/
		if err:=dbop.SaveEncMeta(&encreq);err!=nil{
			encack.Code=api.ERR_INTERNAL
			encack.Msg=err.Error()
		}else{
			encack.Code=0
			encack.Msg="OK"
		}
		json.NewEncoder(w).Encode(encack)
	}else{
		http.NotFound(w,r)
	}
}

func GetDataInfoFunc(w http.ResponseWriter, r *http.Request){
	if r.Method=="POST"{
		ifack:=api.NewDataInfoAck()
		w.Header().Set("Content-Type","application/json")
		var difreq api.GetDataInfoReq
		err:=json.NewDecoder(r.Body).Decode(&difreq)
		if err!=nil{
			Debug("Decode json error:",err)
			ifack.Data=nil
			ifack.Code=api.ERR_BADPARAM
			json.NewEncoder(w).Encode(ifack)
			return
		}
        if g_config.Debug{
            DebugJson("Request:",&difreq)
            defer DebugJson("Response:",ifack)
        }
		/*
		_,err=GetLoginUserInfo(sifreq.Token)
        if err!=nil{
            sifack.Code=1
            sifack.Msg="You should login first"
            json.NewEncoder(w).Encode(sifack)
            return
        }*/

		retdata,err:=dbop.GetEncDataInfo(difreq.Uuid)
		if err!=nil{
			ifack.Data=nil
			ifack.Code=api.ERR_INVDATA
			ifack.Msg=err.Error()
		}else{
			ifack.Code=0
			ifack.Msg="OK"
			ifack.Data=retdata
		}
        json.NewEncoder(w).Encode(ifack)
	}else{
		http.NotFound(w,r)
	}
}

func SearchEncDataFunc(w http.ResponseWriter, r *http.Request){
	if r.Method=="POST"{
		w.Header().Set("Content-Type","application/json")
		var sedreq api.SearchEncDataReq
		sedack:=api.NewSearchEncAck(make([]*api.EncDataNode,0,0))
		err:=json.NewDecoder(r.Body).Decode(&sedreq)
		if err!=nil{
			log.Println("Decode json error:",err)
			sedack.Data=nil
			sedack.Code=api.ERR_BADPARAM
			json.NewEncoder(w).Encode(sedack)
			return
		}
        if g_config.Debug{
            DebugJson("Request:",&sedreq)
            defer DebugJson("Response:",sedack)
        }
		if sedreq.StartItem<0 || sedreq.MaxCount<0{
			sedack.Data=nil
			sedack.Code=api.ERR_INVDATA
			sedack.Msg="Invalid search parameter"
			json.NewEncoder(w).Encode(sedack)
			return
		}

		uinfo,err:=GetLoginUserInfo(sedreq.Token)
        if err!=nil{
			sedack.Data=nil
            sedack.Code=api.ERR_INVDATA
            sedack.Msg="You should login first"
            json.NewEncoder(w).Encode(sedack)
            return
        }
		if uinfo.Id!=sedreq.UserId {
			sedack.Data=nil
			sedack.Code=api.ERR_ACCESS
			sedack.Msg="You should login as the userid you searched for"
			json.NewEncoder(w).Encode(sedack)
			return
		}
		objs,err:=dbop.SearchEncData(&sedreq)
		if err!=nil{
			sedack.Data=nil
			sedack.Code=api.ERR_INVDATA
			sedack.Msg=err.Error()
		}else{
			sedack=api.NewSearchEncAck(objs)
			sedack.Code=0
			sedack.Msg="OK"
		}
		json.NewEncoder(w).Encode(sedack)

		/*
		var linfo *LoginUserInfo
		linfo,err=GetLoginUserInfo(sifreq.Token)
		*/
	}
}

// check token/ownerid later, valid only when ownerid==fromid || owner==toid
func SearchShareDataFunc(w http.ResponseWriter, r *http.Request){
	if r.Method=="POST"{
		w.Header().Set("Content-Type","application/json")
		var ssdreq api.SearchShareDataReq
		ssdack:=api.NewSearchDataAck(make([]*api.ShareDataNode,0,0))
		err:=json.NewDecoder(r.Body).Decode(&ssdreq)
		if err!=nil{
			log.Println("Decode json error:",err)
			ssdack.Data=nil
			ssdack.Code=api.ERR_BADPARAM
			json.NewEncoder(w).Encode(ssdack)
			return
		}
        if g_config.Debug{
            DebugJson("Request:",&ssdreq)
            defer DebugJson("Response:",ssdack)
        }

		if ssdreq.StartItem<0 || ssdreq.MaxCount<0{
			ssdack.Data=nil
			ssdack.Code=api.ERR_INVDATA
			ssdack.Msg="Invalid search parameter"
			json.NewEncoder(w).Encode(ssdack)
			return
		}

		uinfo,err:=GetLoginUserInfo(ssdreq.Token)
        if err!=nil{
			ssdack.Data=nil
            ssdack.Code=api.ERR_INVDATA
            ssdack.Msg="You should login first"
            json.NewEncoder(w).Encode(ssdack)
            return
        }
		if uinfo.Id!=ssdreq.ToId && uinfo.Id!=ssdreq.FromId{
			ssdack.Data=nil
			ssdack.Code=api.ERR_ACCESS
			ssdack.Msg="You should be either OWNER or RECEIVER of the data"
			json.NewEncoder(w).Encode(ssdack)
			return
		}

		objs,err:=dbop.SearchShareData(&ssdreq)
		if err!=nil{
			ssdack.Data=nil
			ssdack.Code=api.ERR_INVDATA
			ssdack.Msg=err.Error()
		}else{
			ssdack=api.NewSearchDataAck(objs)
			ssdack.Code=0
			ssdack.Msg="OK"
		}
		json.NewEncoder(w).Encode(ssdack)

		/*
		var linfo *LoginUserInfo
		linfo,err=GetLoginUserInfo(sifreq.Token)
		*/
	}
}

func GetShareInfoFunc(w http.ResponseWriter, r *http.Request){
	//if r.Method=="GET"{
	if r.Method=="POST"{
		sifack:=api.NewShareInfoAck()
		sifack.Data=nil
		w.Header().Set("Content-Type","application/json")
		var sifreq api.ShareInfoReq
		err:=json.NewDecoder(r.Body).Decode(&sifreq)
		if err!=nil{
			log.Println("Decode json error:",err)
			sifack.Data=nil
			sifack.Code=api.ERR_BADPARAM
			json.NewEncoder(w).Encode(sifack)
			return
		}
        if g_config.Debug{
            DebugJson("Request:",&sifreq)
            defer DebugJson("Response:",sifack)
        }

	var retdata *api.ShareInfoData
	var linfo *LoginUserInfo
	linfo,err=GetLoginUserInfo(sifreq.Token)
	if err!=nil{ // not a valid token
		if sifreq.NeedKey==1{
			sifack.Data=nil
			sifack.Code=api.ERR_INVDATA
			sifack.Msg="You should login first"
			json.NewEncoder(w).Encode(sifack)
			return
		}else{ // if token is not valid , LeftUse may be incorrect(0)
			retdata,err=dbop.GetShareInfoData(sifreq.Uuid)
		}
	}else{ // NeedKey should be checked later, if NeedKey==1 will cause LeftUse--
		retdata,err=dbop.GetUserShareInfoData(sifreq.Uuid,linfo.Id)
	}
	if err!=nil{ // get share info error in db
		sifack.Data=nil
		sifack.Code=api.ERR_INVDATA
		sifack.Msg=err.Error()
		json.NewEncoder(w).Encode(sifack)
		return
	}

	if sifreq.NeedKey==0{
			retdata.EncKey=""
	}else{
		inlist:=false
		for _,id:=range retdata.RcvrIds{
			if linfo.Id==id{
				inlist=true
				break
			}
		}
		if !inlist{
			sifack.Data=nil
			sifack.Code=api.ERR_ACCESS
			sifack.Msg="user not in share list"
			json.NewEncoder(w).Encode(sifack)
			return
		}
		if retdata.LeftUse==0{
			sifack.Data=nil
			sifack.Code=api.ERR_ACCESS
			sifack.Msg="open times exhausted"
	        json.NewEncoder(w).Encode(sifack)
	        return
		}
        strexp:=strings.Replace(retdata.Expire," ","T",1)+"+08:00"
        tmexp,err:=time.Parse(time.RFC3339,strexp)
        if err!=nil{
			sifack.Data=nil
			sifack.Code=api.ERR_INVDATA
			sifack.Msg="bad expire time"
			json.NewEncoder(w).Encode(sifack)
			return
        }
        tmnow:=time.Now()
        if tmnow.After(tmexp){
			sifack.Data=nil
			sifack.Code=api.ERR_ACCESS
			sifack.Msg="shared data has expired at:"+retdata.Expire
			json.NewEncoder(w).Encode(sifack)
            return
        }

		if retdata.LeftUse>0{
			err=dbop.DecreaseOpenTimes(retdata,linfo.Id)
			if err!=nil{
				sifack.Data=nil
				sifack.Code=api.ERR_INTERNAL
				sifack.Msg=err.Error()
                json.NewEncoder(w).Encode(sifack)
	            return
			}
		}// otherwise, LeftUse==-1, ulimited
	}
	sifack.Code=0
	sifack.Msg="OK"
	sifack.Data=retdata
	json.NewEncoder(w).Encode(sifack)
	} else{
		http.NotFound(w,r)
	}
}


func ShareDataFunc(w http.ResponseWriter, r *http.Request){
	if r.Method=="POST"{
		shrack:=api.NewSimpleAck()
		w.Header().Set("Content-Type","application/json")
		var shrreq api.ShareDataReq
		err:=json.NewDecoder(r.Body).Decode(&shrreq)
		if err!=nil{
			log.Println("Decode json error:",err)
			shrack.Code=api.ERR_BADPARAM
			json.NewEncoder(w).Encode(shrack)
			return
		}
        if g_config.Debug{
            DebugJson("Request:",&shrreq)
            defer DebugJson("Response:",shrack)
        }
		luinfo,err:=GetLoginUserInfo(shrreq.Token)
		if err!=nil{
			shrack.Code=api.ERR_INVDATA
			shrack.Msg=err.Error()
			json.NewEncoder(w).Encode(shrack)
			return
		}
		if luinfo.Id!=shrreq.Data.OwnerId{
			shrack.Code=api.ERR_ACCESS
			shrack.Msg="Invalid user"
			json.NewEncoder(w).Encode(shrack)
			return
		}

		// user info checked ok
		// reference crypt.go:dbop.SaveMeta
		if err=dbop.WriteShareInfo(shrreq.Data);err!=nil{
			shrack.Code=api.ERR_INTERNAL
			shrack.Msg=err.Error()
		}else{
			shrack.Code=0
			shrack.Msg="OK"
		}
		json.NewEncoder(w).Encode(shrack)
	}else{
		http.NotFound(w,r)
	}
}

func CreateRCFunc(w http.ResponseWriter, r *http.Request){
	if r.Method=="POST"{
		rcack:=api.NewRCInfoAck()
		w.Header().Set("Content-Type","application/json")
		var rcreq api.CreateRCReq
		err:=json.NewDecoder(r.Body).Decode(&rcreq)
		if err!=nil{
			Debug("Decode json error:",err)
			rcack.Data=nil
			rcack.Code=api.ERR_BADPARAM
			json.NewEncoder(w).Encode(rcack)
			return
		}
        if g_config.Debug{
            DebugJson("Request:",&rcreq)
            defer DebugJson("Response:",rcack)
        }
		luinfo,err:=GetLoginUserInfo(rcreq.Token)
		if err!=nil{
			rcack.Data=nil
			rcack.Code=api.ERR_INVDATA
			rcack.Msg=err.Error()
			json.NewEncoder(w).Encode(rcack)
			return
		}
/*		if luinfo.Id!=rcreq.Data.UserId{
			rcack.Code=2
			rcack.Msg="Invalid user"
			json.NewEncoder(w).Encode(rcack)
			return
		}*/
		rcreq.Data.UserId=luinfo.Id
		rcreq.Data.IPAddr=GetClientPublicIP(r)
		if rcreq.Data.EndTime==""{
			rcreq.Data.EndTime=core.GetCurTime()
		}
		if err:=dbop.NewRunContext(rcreq.Data);err!=nil{
			rcack.Data=nil
			rcack.Code=api.ERR_INTERNAL
			rcack.Msg=err.Error()
		}else{
			rcack.Code=0
			rcack.Msg="OK"
			rcack.Data=rcreq.Data
		}
		json.NewEncoder(w).Encode(rcack)
	}else{
		http.NotFound(w,r)
	}
}

func UpdateRCFunc(w http.ResponseWriter, r *http.Request){
	if r.Method=="POST"{
		ack:=api.NewSimpleAck()
		w.Header().Set("Content-Type","application/json")
		var urcreq api.UpdateRCReq
		err:=json.NewDecoder(r.Body).Decode(&urcreq)
		if err!=nil{
			Debug("Decode json error:",err)
			ack.Code=api.ERR_BADPARAM
			json.NewEncoder(w).Encode(ack)
			return
		}
        if g_config.Debug{
            DebugJson("Request:",&urcreq)
            defer DebugJson("Response:",ack)
        }
		luinfo,err:=GetLoginUserInfo(urcreq.Token)
		if err!=nil{
			ack.Code=api.ERR_INVDATA
			ack.Msg=err.Error()
			json.NewEncoder(w).Encode(ack)
			return
		}

		if err:=dbop.UpdateRunContext(luinfo.Id,urcreq.RCId,urcreq.OutputUuid,urcreq.EndTime);err!=nil{
			ack.Code=api.ERR_INVDATA
			ack.Msg=err.Error()
		}else{
			ack.Code=0
			ack.Msg="OK"
		}
		json.NewEncoder(w).Encode(ack)
	}else{
		http.NotFound(w,r)
	}
}

func GetRCInfoFunc(w http.ResponseWriter, r *http.Request){
	if r.Method=="POST"{
		rcack:=api.NewRCInfoAck()
		w.Header().Set("Content-Type","application/json")
		var grireq api.GetRCInfoReq
		err:=json.NewDecoder(r.Body).Decode(&grireq)
		if err!=nil{
			Debug("Decode json error:",err)
			rcack.Data=nil
			rcack.Code=api.ERR_BADPARAM
			json.NewEncoder(w).Encode(rcack)
			return
		}
        if g_config.Debug{
            DebugJson("Request:",&grireq)
            defer DebugJson("Response:",rcack)
        }
		/*
		luinfo,err:=GetLoginUserInfo(grireq.Token)
		if err!=nil{
			rcack.Code=1
			rcack.Msg=err.Error()
			json.NewEncoder(w).Encode(rcack)
			return
		}
		if luinfo.Id!=rcreq.Data.UserId{
			rcack.Code=2
			rcack.Msg="Invalid user"
			json.NewEncoder(w).Encode(rcack)
			return
		}*/
		if rcack.Data,err=dbop.GetRCInfo(grireq.RCId);err!=nil{
			rcack.Data=nil
			rcack.Code=api.ERR_INVDATA
			rcack.Msg=err.Error()
		}else{
			rcack.Code=0
			rcack.Msg="OK"
		}
		json.NewEncoder(w).Encode(rcack)
	}else{
		http.NotFound(w,r)
	}
}
