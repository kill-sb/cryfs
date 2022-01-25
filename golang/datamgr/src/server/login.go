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
		log.Println(ainfo)
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
			token.Id=id
			token.Key=key
			token.Token=makeToken(token.Id)
			luinfo:=GenUserInfo(&ainfo,token)
			AddToken(token.Token,luinfo)
		//	log.Println(*luinfo)
			token.Status=0
			token.ErrInfo=""
		}else{
			token.Status=1
			token.ErrInfo="Invalid user/password"
		}
		json.NewEncoder(w).Encode(token)
	}else{
//		http.Redirect(w,r,"/",http.StatusFound)
		http.NotFound(w,r)
	}
}

func GetUserFunc(w http.ResponseWriter, r *http.Request){
}
