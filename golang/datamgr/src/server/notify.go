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
			snack.Code=1
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
