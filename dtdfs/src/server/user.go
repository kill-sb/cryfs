package main

import (
    "net/http"
    "encoding/json"
    "fmt"
    "dbop"
    //core "coredata"
    api "apiv1"
)

func GetUserFunc(w http.ResponseWriter, r *http.Request){
	if r.Method=="POST"{
		usrack:=api.NewUserInfoAck()
		w.Header().Set("Content-Type","application/json")
		var usrreq api.GetUserReq
		err:=json.NewDecoder(r.Body).Decode(&usrreq)
		if err!=nil{
			Debug("Decode json error:",err)
			usrack.Data=nil
			usrack.Code=api.ERR_BADPARAM
			json.NewEncoder(w).Encode(usrack)
			return
		}
		if g_config.Debug{
			DebugJson("Request:",&usrreq)
			defer DebugJson("Response:",usrack)
		}
		/*
		_,err=GetLoginUserInfo(sifreq.Token)
        if err!=nil{
            sifack.Code=1
            sifack.Msg="You should login first"
            json.NewEncoder(w).Encode(sifack)
            return
        }*/
		usrack.Code=0
		usrack.Msg="OK"
		for _,v:=range usrreq.Id{
			usr,err:=dbop.GetUserInfo(v)
			if err!=nil{
				usrack.Data=nil
				usrack.Code=api.ERR_INVDATA
				usrack.Msg=fmt.Sprintf("search userid=%d error: %s",v,err.Error())
				usrack.Data=[]api.UserInfoData{}
				break
			}else{
				usrack.Data=append(usrack.Data,*usr)
			}
		}
		json.NewEncoder(w).Encode(usrack)
	}else{
		http.NotFound(w,r)
	}
}

func FindUserNameFunc(w http.ResponseWriter, r *http.Request){
	if r.Method=="POST"{
		usrack:=api.NewUserInfoAck()
		w.Header().Set("Content-Type","application/json")
		var usrreq api.FindUserNameReq
		err:=json.NewDecoder(r.Body).Decode(&usrreq)
		if err!=nil{
			Debug("Decode json error:",err)
			usrack.Data=nil
			usrack.Code=api.ERR_BADPARAM
			json.NewEncoder(w).Encode(usrack)
			return
		}
        if g_config.Debug{
            DebugJson("Request:",&usrreq)
            defer DebugJson("Response:",usrack)
        }

		/*
		_,err=GetLoginUserInfo(sifreq.Token)
        if err!=nil{
            sifack.Code=1
            sifack.Msg="You should login first"
            json.NewEncoder(w).Encode(sifack)
            return
        }*/
		usrack.Code=0
		usrack.Msg="OK"
		for _,v:=range usrreq.Name{
			usr,err:=dbop.GetUserInfoByName(v)
			if err!=nil{
				usrack.Data=nil
				usrack.Code=api.ERR_INVDATA
				usrack.Msg=err.Error()
				break
			}else{
//				Debug(usr.Name,usr.Id)
				usrack.Data=append(usrack.Data,*usr)
			}
		}
		json.NewEncoder(w).Encode(usrack)
	}else{
		http.NotFound(w,r)
	}

}

func AddContactsFunc(w http.ResponseWriter, r *http.Request){
	if r.Method=="POST"{
		acack:=api.NewSimpleAck()
		w.Header().Set("Content-Type","application/json")
		var acreq api.AddContactReq
		err:=json.NewDecoder(r.Body).Decode(&acreq)
		if err!=nil{
			Debug("Decode json error:",err)
			acack.Code=api.ERR_BADPARAM
			json.NewEncoder(w).Encode(acack)
			return
		}
		if g_config.Debug{
			DebugJson("Request:",&acreq)
			defer DebugJson("Response:",acack)
		}
		uinfo,err:=GetLoginUserInfo(acreq.Token)
        if err!=nil{
            acack.Code=api.ERR_INVDATA
            acack.Msg="You should login first"
            json.NewEncoder(w).Encode(acack)
            return
        }
		for _,id:=range acreq.Ids{
			_,err=dbop.GetUserInfo(id)
			if err!=nil{
				acack.Code=api.ERR_INVDATA
				acack.Msg=err.Error()
				json.NewEncoder(w).Encode(acack)
				return
			}
		}
		for _,id:=range acreq.Ids{
			err=dbop.NewContact(uinfo.Id,id)
			if err!=nil{
				acack.Code=api.ERR_INTERNAL
				acack.Msg=err.Error()
				json.NewEncoder(w).Encode(acack)
				return
			}
		}
		acack.Code=0
		acack.Msg="OK"
		json.NewEncoder(w).Encode(acack)
	}else{
		http.NotFound(w,r)
	}
}

func DelContactsFunc(w http.ResponseWriter, r *http.Request){
	if r.Method=="POST"{
		dcack:=api.NewSimpleAck()
		w.Header().Set("Content-Type","application/json")
		var dcreq api.DelContactReq
		err:=json.NewDecoder(r.Body).Decode(&dcreq)
		if err!=nil{
			Debug("Decode json error:",err)
			dcack.Code=api.ERR_BADPARAM
			json.NewEncoder(w).Encode(dcack)
			return
		}
		if g_config.Debug{
			DebugJson("Request:",&dcreq)
			defer DebugJson("Response:",dcack)
		}
		uinfo,err:=GetLoginUserInfo(dcreq.Token)
        if err!=nil{
            dcack.Code=api.ERR_INVDATA
            dcack.Msg="You should login first"
            json.NewEncoder(w).Encode(dcack)
            return
        }
		for _,id:=range dcreq.Ids{
			err=dbop.DelContact(uinfo.Id,id)
			if err!=nil{
				dcack.Code=api.ERR_INTERNAL
				dcack.Msg=err.Error()
				json.NewEncoder(w).Encode(dcack)
				return
			}
		}
		dcack.Code=0
		dcack.Msg="OK"
		json.NewEncoder(w).Encode(dcack)
	}else{
		http.NotFound(w,r)
	}

}

func GetContactsFunc(w http.ResponseWriter, r *http.Request){
	if r.Method=="POST"{
		gcack:=api.NewGetContactsAck()
		w.Header().Set("Content-Type","application/json")
		var gcreq api.GetContactReq
		err:=json.NewDecoder(r.Body).Decode(&gcreq)
		if err!=nil{
			Debug("Decode json error:",err)
			gcack.Data=nil
			gcack.Code=api.ERR_BADPARAM
			json.NewEncoder(w).Encode(gcack)
			return
		}
		if g_config.Debug{
			DebugJson("Request:",&gcreq)
			defer DebugJson("Response:",gcack)
		}
		uinfo,err:=GetLoginUserInfo(gcreq.Token)
        if err!=nil{
			gcack.Data=nil
            gcack.Code=api.ERR_INVDATA
            gcack.Msg="You should login first"
            json.NewEncoder(w).Encode(gcack)
            return
        }
		gcack.Data,err=dbop.ListContacts(uinfo.Id)
		if err!=nil{
			gcack.Data=nil
			gcack.Code=api.ERR_INVDATA
			gcack.Msg=err.Error()
		}else{
			gcack.Code=0
			gcack.Msg="OK"
		}
		json.NewEncoder(w).Encode(gcack)
	}else{
		http.NotFound(w,r)
	}
}

func FuzzySearchFunc(w http.ResponseWriter, r *http.Request){
	if r.Method=="POST"{
		fsack:=api.NewGetContactsAck()
		w.Header().Set("Content-Type","application/json")
		var fsreq api.FzSearchReq
		err:=json.NewDecoder(r.Body).Decode(&fsreq)
		if err!=nil{
			Debug("Decode json error:",err)
			fsack.Data=nil
			fsack.Code=api.ERR_BADPARAM
			json.NewEncoder(w).Encode(fsack)
			return
		}
		if g_config.Debug{
			DebugJson("Request:",&fsreq)
			defer DebugJson("Response:",fsack)
		}
		uinfo,err:=GetLoginUserInfo(fsreq.Token)
        if err!=nil{
			fsack.Data=nil
            fsack.Code=api.ERR_INVDATA
            fsack.Msg="You should login first"
            json.NewEncoder(w).Encode(fsack)
            return
        }
		fsack.Data,err=dbop.FuzzySearch(uinfo.Id,fsreq.Keyword)
		if err!=nil{
			fsack.Data=nil
			fsack.Code=api.ERR_INTERNAL
			fsack.Msg=err.Error()
		}else{
			fsack.Code=0
			fsack.Msg="OK"
		}
		json.NewEncoder(w).Encode(fsack)
	}else{
		http.NotFound(w,r)
	}
}
