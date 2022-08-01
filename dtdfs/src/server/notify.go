package main

import (
    "net/http"
    "encoding/json"
    "log"
//  core "coredata"
//  "os"
	"fmt"
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
			snack.Data=0
			snack.Code=api.ERR_BADPARAM
			json.NewEncoder(w).Encode(snack)
			return
		}
        if g_config.Debug{
            DebugJson("Request:",&snreq)
            defer DebugJson("Response:",snack)
        }
		luinfo,err:=GetLoginUserInfo(snreq.Token)
		if err!=nil{
			snack.Data=0
			snack.Code=api.ERR_INVDATA
			snack.Msg=err.Error()
			json.NewEncoder(w).Encode(snack)
			return
		}
		if luinfo.Id!=snreq.Data.FromUid{
			snack.Data=0
			snack.Code=api.ERR_ACCESS
			snack.Msg="Send user is different from login user"
			json.NewEncoder(w).Encode(snack)
			return
		}
		_,err=dbop.GetUserInfo(snreq.Data.ToUid)
		if err!=nil{
			snack.Data=0
			snack.Code=api.ERR_INVDATA
			snack.Msg="Invalid receive user"
			json.NewEncoder(w).Encode(snack)
			return
		}

		// user info checked ok
		if err=dbop.NewNotify(snreq.Data);err!=nil{
			snack.Data=0
			snack.Code=api.ERR_INTERNAL
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

func SetNotifyStatFunc(w http.ResponseWriter, r *http.Request){
	if r.Method=="POST"{
		snsack:=api.NewSimpleAck()
		w.Header().Set("Content-Type","application/json")
		var snsreq api.SetNotifyStatReq
		err:=json.NewDecoder(r.Body).Decode(&snsreq)
		if err!=nil{
			log.Println("Decode json error:",err)
			snsack.Code=api.ERR_BADPARAM
			json.NewEncoder(w).Encode(snsack)
			return
		}
        if g_config.Debug{
            DebugJson("Request:",&snsreq)
            defer DebugJson("Response:",snsack)
        }
		luinfo,err:=GetLoginUserInfo(snsreq.Token)
		if err!=nil{
			snsack.Code=api.ERR_INVDATA
			snsack.Msg=err.Error()
			json.NewEncoder(w).Encode(snsack)
			return
		}
		alen:=len(snsreq.Ids)
		if alen!=len(snsreq.Stats){
			snsack.Code=api.ERR_INVDATA
			snsack.Msg="Invalid parameters"
			json.NewEncoder(w).Encode(snsack)
			return
		}
		for i:=0;i<alen;i++{
			ninfo,err:=dbop.GetNotifyInfo(snsreq.Ids[i])
			if err!=nil{
				snsack.Code=api.ERR_INVDATA
				snsack.Msg=err.Error()
				json.NewEncoder(w).Encode(snsack)
				return
			}
			if luinfo.Id!=ninfo.ToUid {
				snsack.Code=api.ERR_ACCESS
				snsack.Msg="Login user is not receive user."
				json.NewEncoder(w).Encode(snsack)
				return
			}
			if err=dbop.SetNotifyStat(snsreq.Ids[i],snsreq.Stats[i]);err!=nil{
				snsack.Code=api.ERR_INVDATA
				snsack.Msg=err.Error()
                json.NewEncoder(w).Encode(snsack)
                return
			}
		}
		snsack.Code=0
		snsack.Msg="OK"
		json.NewEncoder(w).Encode(snsack)
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
			dnack.Code=api.ERR_BADPARAM
			json.NewEncoder(w).Encode(dnack)
			return
		}
        if g_config.Debug{
            DebugJson("Request:",&dnreq)
            defer DebugJson("Response:",dnack)
        }
		luinfo,err:=GetLoginUserInfo(dnreq.Token)
		if err!=nil{
			dnack.Code=api.ERR_INVDATA
			dnack.Msg=err.Error()
			json.NewEncoder(w).Encode(dnack)
			return
		}
		for _,id:=range dnreq.Ids{
			ninfo,err:=dbop.GetNotifyInfo(id)
			if err!=nil{
				dnack.Code=api.ERR_INVDATA
				dnack.Msg=err.Error()
				json.NewEncoder(w).Encode(dnack)
				return
			}
			if luinfo.Id!=ninfo.ToUid && luinfo.Id!=ninfo.FromUid{
				dnack.Code=api.ERR_ACCESS
				dnack.Msg="Login user is neither send user nor receive user."
				json.NewEncoder(w).Encode(dnack)
				return
			}
		}
		if err=dbop.DelNotifies(dnreq.Ids);err!=nil{
			dnack.Code=api.ERR_INTERNAL
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

func GetNotifyInfoFunc(w http.ResponseWriter, r *http.Request){
	if r.Method=="POST"{
		gniack:=api.NewSearchNotifiesAck()
		w.Header().Set("Content-Type","application/json")
		var gnireq api.GetNotifyInfoReq
		err:=json.NewDecoder(r.Body).Decode(&gnireq)
		if err!=nil{
			log.Println("Decode json error:",err)
			gniack.Data=nil
			gniack.Code=api.ERR_BADPARAM
			json.NewEncoder(w).Encode(gniack)
			return
		}
        if g_config.Debug{
            DebugJson("Request:",&gnireq)
            defer DebugJson("Response:",gniack)
        }
		luinfo,err:=GetLoginUserInfo(gnireq.Token)
		if err!=nil{
			gniack.Data=nil
			gniack.Code=api.ERR_INVDATA
			gniack.Msg=err.Error()
			json.NewEncoder(w).Encode(gniack)
			return
		}

		nIds:=len(gnireq.Ids)
		gniack.Data=make([]*api.NotifyInfo,0,nIds)
		for _,id:=range gnireq.Ids{
			ninfo,err:=dbop.GetNotifyInfo(id)
			if err!=nil{
				gniack.Code=api.ERR_INVDATA
				gniack.Msg=err.Error()
				gniack.Data=nil
				break
			}
			if luinfo.Id!=ninfo.FromUid && luinfo.Id!=ninfo.ToUid{
				gniack.Code=api.ERR_ACCESS
				gniack.Msg=fmt.Sprintf("Current user is neither sender nor receiver of notify-%d",ninfo.Id)
				gniack.Data=nil
				break
			}
			gniack.Data=append(gniack.Data,ninfo)
		}
		if gniack.Data!=nil && nIds==len(gniack.Data){ // all data are ok
			gniack.Code=0
			gniack.Msg="OK"
		}
		json.NewEncoder(w).Encode(gniack)
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
			snack.Data=nil
			snack.Code=api.ERR_BADPARAM
			json.NewEncoder(w).Encode(snack)
			return
		}
        if g_config.Debug{
            DebugJson("Request:",&snreq)
            defer DebugJson("Response:",snack)
        }

        if snreq.StartItem<0 || snreq.MaxCount<0{
            snack.Data=nil
            snack.Code=api.ERR_INVDATA
            snack.Msg="Invalid search parameter"
            json.NewEncoder(w).Encode(snack)
            return
        }

		luinfo,err:=GetLoginUserInfo(snreq.Token)
		if err!=nil{
			snack.Data=nil
			snack.Code=api.ERR_INVDATA
			snack.Msg=err.Error()
			json.NewEncoder(w).Encode(snack)
			return
		}
		if luinfo.Id!=snreq.FromUid && luinfo.Id!=snreq.ToUid{
			snack.Data=nil
			snack.Code=api.ERR_ACCESS
			snack.Msg="Login user is neither send user nor receive user"
			json.NewEncoder(w).Encode(snack)
			return
		}

		// user info checked ok
		if snack.Data,err=dbop.SearchNotifies(&snreq);err!=nil{
			snack.Code=api.ERR_INVDATA
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
