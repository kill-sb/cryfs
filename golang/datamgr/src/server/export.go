package main

import (
	"net/http"
	"encoding/json"
	"log"
//	core "coredata"
//	"os"
	"dbop"
	api "apiv1"
)

func ExportDataFunc(w http.ResponseWriter, r *http.Request){
	if r.Method=="POST"{
		epack:=api.NewExProcAck()
		w.Header().Set("Content-Type","application/json")
		var epreq api.NewExportReq
		err:=json.NewDecoder(r.Body).Decode(&epreq)
		if err!=nil{
			Debug("Decode json error:",err)
			json.NewEncoder(w).Encode(epack)
			return
		}
        if g_config.Debug{
            DebugJson("Request:",&epreq)
            defer DebugJson("Response:",epack)
        }

		luinfo,err:=GetLoginUserInfo(epreq.Token)
		if err!=nil{
			epack.Code=1
			epack.Msg=err.Error()
			json.NewEncoder(w).Encode(epack)
			return
		}
/*		ownerid,err:=dbop.GetDataOwner(epreq.Data)
		if err!=nil{
            epack.Code=1
            epack.Msg=err.Error()
            json.NewEncoder(w).Encode(epack)
            return
        }

		if luinfo.Id!=ownerid{
			epack.Code=2
			epack.Msg="Invalid user"
			json.NewEncoder(w).Encode(epack)
			return
		}*/
		// user info checked ok
		if epack.Data,err=dbop.NewExport(epreq.Data,luinfo.Id,&epreq.Comment);err!=nil{
			epack.Code=1
			epack.Msg=err.Error()
			epack.Data=nil
		}else{
			epack.Code=0
			epack.Msg="OK"
		}
		json.NewEncoder(w).Encode(epack)
	}else{
		http.NotFound(w,r)
	}
}

func GetExProgressFunc(w http.ResponseWriter, r *http.Request){
	if r.Method=="POST"{
		ifack:=api.NewDataInfoAck()
		w.Header().Set("Content-Type","application/json")
		var difreq api.GetDataInfoReq
		err:=json.NewDecoder(r.Body).Decode(&difreq)
		if err!=nil{
			Debug("Decode json error:",err)
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
			ifack.Code=2
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

// check token/ownerid later, valid only when ownerid==fromid || owner==toid
func SearchExpReqFunc(w http.ResponseWriter, r *http.Request){
	if r.Method=="POST"{
		w.Header().Set("Content-Type","application/json")
		var ssdreq api.SearchShareDataReq
		ssdack:=api.NewSearchDataAck(make([]*api.ShareDataNode,0,0))
		err:=json.NewDecoder(r.Body).Decode(&ssdreq)
		if err!=nil{
			log.Println("Decode json error:",err)
			json.NewEncoder(w).Encode(ssdack)
			return
		}

		uinfo,err:=GetLoginUserInfo(ssdreq.Token)
        if err!=nil{
            ssdack.Code=1
            ssdack.Msg="You should login first"
            json.NewEncoder(w).Encode(ssdack)
            return
        }
		if uinfo.Id!=ssdreq.ToId && uinfo.Id!=ssdreq.FromId{
			ssdack.Code=3
			ssdack.Msg="You should be either OWNER or RECEIVER of the data"
			json.NewEncoder(w).Encode(ssdack)
			return
		}

		objs,err:=dbop.SearchShareData(&ssdreq)
		if err!=nil{
			ssdack.Code=2
			ssdack.Msg=err.Error()
		}else{
			ssdack=api.NewSearchDataAck(objs)
			ssdack.Code=0
			ssdack.Msg="OK"
		}
		json.NewEncoder(w).Encode(ssdack)
        if g_config.Debug{
            DebugJson("Request:",&ssdreq)
            DebugJson("Response:",ssdack)
        }
		/*
		var linfo *LoginUserInfo
		linfo,err=GetLoginUserInfo(sifreq.Token)
		*/
	}
}

func GetExpReqInfoFunc(w http.ResponseWriter, r *http.Request){
	//if r.Method=="GET"{
	if r.Method=="POST"{
		sifack:=api.NewShareInfoAck()
		w.Header().Set("Content-Type","application/json")
		var sifreq api.ShareInfoReq
		err:=json.NewDecoder(r.Body).Decode(&sifreq)
		if err!=nil{
			log.Println("Decode json error:",err)
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
			sifack.Code=1
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
		sifack.Code=2
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
			sifack.Code=3
			sifack.Msg="user not in share list"
			json.NewEncoder(w).Encode(sifack)
			return
		}
		if retdata.LeftUse==0{
			sifack.Code=4
			sifack.Msg="open times exhausted"
	        json.NewEncoder(w).Encode(sifack)
	        return
		}

		if retdata.LeftUse>0{
			err=dbop.DecreaseOpenTimes(retdata,linfo.Id)
			if err!=nil{
				sifack.Code=5
				sifack.Msg=err.Error()
                json.NewEncoder(w).Encode(sifack)
	            return
			}
			// check expired time later
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


func RespExpReqFunc(w http.ResponseWriter, r *http.Request){
	if r.Method=="POST"{
		shrack:=api.NewShareAck()
		w.Header().Set("Content-Type","application/json")
		var shrreq api.ShareDataReq
		err:=json.NewDecoder(r.Body).Decode(&shrreq)
		if err!=nil{
			log.Println("Decode json error:",err)
			json.NewEncoder(w).Encode(shrack)
			return
		}
        if g_config.Debug{
            DebugJson("Request:",&shrreq)
            defer DebugJson("Response:",shrack)
        }
		luinfo,err:=GetLoginUserInfo(shrreq.Token)
		if err!=nil{
			shrack.Code=1
			shrack.Msg=err.Error()
			json.NewEncoder(w).Encode(shrack)
			return
		}
		if luinfo.Id!=shrreq.Data.OwnerId{
			shrack.Code=2
			shrack.Msg="Invalid user"
			json.NewEncoder(w).Encode(shrack)
			return
		}

		// user info checked ok
		// reference crypt.go:dbop.SaveMeta
		if err=dbop.WriteShareInfo(shrreq.Data);err!=nil{
			shrack.Code=1
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

