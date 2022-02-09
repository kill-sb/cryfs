package main

import (
	"net/http"
	"encoding/json"
	"fmt"
	"crypto/sha256"
	"sync"
	"time"
	"dbop"
	"log"
	core "coredata"
	api "apiv1"
/*	"os"
	_"dbop"
	*/
)

/*
{
	"name":user_name,
	"passwd":password,
	"primask",0 // default
}
*/

const (
	XOR_BASE=0x34345789
)

type LoginUserInfo struct{
    Name string
    Id int32
    Keylocalkey []byte
    PriMask int32
    LogExpire time.Time
	Lock sync.RWMutex
/*  Email string
    Descr string
    RegTime time.Time
    Mobile string
    */

}

func GenUserInfo(ainfo* api.AuthInfo, tinfo *api.TokenInfo)*LoginUserInfo{
	luinfo:=new (LoginUserInfo)
	luinfo.Name=ainfo.Name
	luinfo.Id=tinfo.Id
	luinfo.Keylocalkey=core.StringToBinkey(tinfo.Key)
	luinfo.PriMask=ainfo.PriMask
	luinfo.UpdateToken()
	return luinfo
}

func LoginFunc(w http.ResponseWriter, r *http.Request){
	if r.Method=="POST"{
		token:=api.NewToken()
		w.Header().Set("Content-Type","application/json")
		var ainfo api.AuthInfo
		err:= json.NewDecoder(r.Body).Decode(&ainfo)
		if err!=nil{
			fmt.Println("Decode json error:",r.Body,"-",err)
			json.NewEncoder(w).Encode(token)
			return
		}
		log.Print("")
//		log.Println("login:",ainfo)
//		token.GetUserInfo(&ainfo)
		// check user/passwd
		id,shasum,key,err:=dbop.LookupPasswdSHA(ainfo.Name)
		if err!=nil{
			json.NewEncoder(w).Encode(token)
			return
		}
		sharet:=sha256.Sum256([]byte(ainfo.Passwd))
		shastr:=""
		for _,ch:=range sharet{
			shastr=fmt.Sprintf("%s%02x",shastr,ch)
		}
		if shastr==shasum{// password check ok
			token.Data.Id=id
			token.Data.Key=key
			token.Data.Token=makeToken(token.Data.Id)
			luinfo:=GenUserInfo(&ainfo,token.Data)
			AddToken(token.Data.Token,luinfo)
		//	log.Println(*luinfo)
			token.Code=0
			token.Msg="OK"
		}else{
			token.Code=1
			token.Msg="Invalid user/password"
		}
		json.NewEncoder(w).Encode(token)
	}else{
//		http.Redirect(w,r,"/",http.StatusFound)
		http.NotFound(w,r)
	}
}

func GetUserFunc(w http.ResponseWriter, r *http.Request){
	if r.Method=="POST"{
		usrack:=api.NewUserInfoAck()
		w.Header().Set("Content-Type","application/json")
		var usrreq api.GetUserReq
		err:=json.NewDecoder(r.Body).Decode(&usrreq)
		if err!=nil{
			log.Println("Decode json error:",err)
			json.NewEncoder(w).Encode(usrack)
			return
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
				usrack.Code=3
				usrack.Msg=fmt.Sprintf("search userid=%d error: %s",v,err.Error())
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
			log.Println("Decode json error:",err)
			json.NewEncoder(w).Encode(usrack)
			return
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
				usrack.Code=3
				usrack.Msg=fmt.Sprintf("search user %s error: %s",v,err.Error())
				break
			}else{
				log.Println(usr.Name,usr.Id)
				usrack.Data=append(usrack.Data,*usr)
			}
		}
		json.NewEncoder(w).Encode(usrack)
	}else{
		http.NotFound(w,r)
	}

}
