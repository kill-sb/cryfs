package main

import (
    "net/http"
    "encoding/json"
    "log"
//  core "coredata"
//  "os"
    "dbop"
    api "apiv1"
)

func CreateNotifyFunc(w http.ResponseWriter, r *http.Request){
	if r.Method=="POST"{
		snack:=api.NewSendNotifyAck()
		w.Header().Set("Content-Type","application/json")
		var snreq api.SendNotifyReq
		err:=json.NewDecoder(r.Body).Decode(&snreq)
		if err!=nil{
			log.Println("Decode json error:",err)
			json.NewEncoder(w).Encode(snack)
			return
		}
        if g_config.Debug{
            DebugJson("Request:",&snreq)
            defer DebugJson("Response:",snack)
        }
		luinfo,err:=GetLoginUserInfo(snreq.Token)
		if err!=nil{
			snack.Code=1
			snack.Msg=err.Error()
			json.NewEncoder(w).Encode(snack)
			return
		}
		if luinfo.Id!=snreq.Data.FromUid{
			snack.Code=2
			snack.Msg="Send user is different from login user"
			json.NewEncoder(w).Encode(snack)
			return
		}
		_,err=dbop.GetUserInfo(snreq.Data.ToUid)
		if err!=nil{
			snack.Code=3
			snack.Msg="Invalid receive user"
			json.NewEncoder(w).Encode(snack)
			return
		}

		// user info checked ok
		if err=dbop.NewNotify(snreq.Data);err!=nil{
			snack.Code=4
			snack.Msg=err.Error()
		}else{
			snack.Code=0
			snack.Msg="OK"
			snack.Data=snreq.Data.Id
		}
		json.NewEncoder(w).Encode(snack)
	}else{
		http.NotFound(w,r)
	}
}

func DelNotifyFunc(w http.ResponseWriter, r *http.Request){
	if r.Method=="POST"{
		dnack:=api.NewSimpleAck()
		w.Header().Set("Content-Type","application/json")
		var dnreq api.DelNotifyReq
		err:=json.NewDecoder(r.Body).Decode(&dnreq)
		if err!=nil{
			log.Println("Decode json error:",err)
			json.NewEncoder(w).Encode(dnack)
			return
		}
        if g_config.Debug{
            DebugJson("Request:",&dnreq)
            defer DebugJson("Response:",dnack)
        }
		luinfo,err:=GetLoginUserInfo(dnreq.Token)
		if err!=nil{
			dnack.Code=1
			dnack.Msg=err.Error()
			json.NewEncoder(w).Encode(dnack)
			return
		}
		for _,id:=range dnreq.Ids{
			ninfo,err:=dbop.GetNotifyInfo(id)
			if err!=nil{
				dnack.Code=1
				dnack.Msg=err.Error()
				json.NewEncoder(w).Encode(dnack)
				return
			}
			if luinfo.Id!=ninfo.ToUid && luinfo.Id!=ninfo.FromUid{
				dnack.Code=2
				dnack.Msg="Login user is neither send user nor receive user."
				json.NewEncoder(w).Encode(dnack)
				return
			}
		}
		if err=dbop.DelNotifies(dnreq.Ids);err!=nil{
			dnack.Code=3
			dnack.Msg=err.Error()
		}else{
			dnack.Code=0
			dnack.Msg="OK"
		}
		json.NewEncoder(w).Encode(dnack)
	}else{
		http.NotFound(w,r)
	}
}

func SearchNotifiesFunc(w http.ResponseWriter, r *http.Request){
	if r.Method=="POST"{
		snack:=api.NewSearchNotifiesAck()
		w.Header().Set("Content-Type","application/json")
		var snreq api.SearchNotifiesReq
		err:=json.NewDecoder(r.Body).Decode(&snreq)
		if err!=nil{
			log.Println("Decode json error:",err)
			json.NewEncoder(w).Encode(snack)
			return
		}
        if g_config.Debug{
            DebugJson("Request:",&snreq)
            defer DebugJson("Response:",snack)
        }
		luinfo,err:=GetLoginUserInfo(snreq.Token)
		if err!=nil{
			snack.Code=1
			snack.Msg=err.Error()
			json.NewEncoder(w).Encode(snack)
			return
		}
		if luinfo.Id!=snreq.FromUid && luinfo.Id!=snreq.ToUid{
			snack.Code=2
			snack.Msg="Login user is neither send user nor receive user"
			json.NewEncoder(w).Encode(snack)
			return
		}

		// user info checked ok
		if snack.Data,err=dbop.SearchNotifies(snreq.FromUid,snreq.ToUid);err!=nil{
			snack.Code=1
			snack.Msg=err.Error()
			snack.Data=nil
		}else{
			snack.Code=0
			snack.Msg="OK"
		}
		json.NewEncoder(w).Encode(snack)
	}else{
		http.NotFound(w,r)
	}
}
