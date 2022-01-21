package main

import (
	"net/http"
	"time"
	"encoding/json"
	"fmt"
	"crypto/sha256"
	"dbop"
	"log"
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

type AuthInfo struct{
	Name string `json:"name"`
	Passwd string `json:"passwd"`
	PriMask int32 `json:"primask"`
}

type TokenInfo struct{
	Id int32 `json:"id"`
	Token int64 `json:"token"`
	Key	string `json:"key"`
	Status int32 `json:"retval"`
	ErrInfo	string `json:"errinfo"`
}

func NewToken()*TokenInfo{
	token:=&TokenInfo{Id:-1,Token:-1,Key:"",Status:-1,ErrInfo:"Error Parameter"}
	return token
}

func (token *TokenInfo)GetUserInfo(info* AuthInfo)error{
	// if return nil, means authoration succeeded,all fields should be fill correctly according to AuthInfo
	return nil
}

func LoginFunc(w http.ResponseWriter, r *http.Request){
	if r.Method=="POST"{
		token:=NewToken()
		w.Header().Set("Content-Type","application/json")
		var ainfo AuthInfo
		err:= json.NewDecoder(r.Body).Decode(&ainfo)
		if err!=nil{
			fmt.Println("Decode json error:",r.Body,"-",err)
			return
		}
	//	log.Println(ainfo)
		token.GetUserInfo(&ainfo)
		// check user/passwd
		id,shasum,key,err:=dbop.LookupPasswdSHA(ainfo.Name)
		if err!=nil{
			return
		}
		sharet:=sha256.Sum256([]byte(ainfo.Passwd))
		shastr:=""
		for _,ch:=range sharet{
			shastr=fmt.Sprintf("%s%02x",shastr,ch)
		}
		if shastr==shasum{
			token.Id=id
			token.Key=key
			token.Token=(time.Now().Unix()<<32)|(int64(id<<8))|(int64(ainfo.PriMask&0xff))
			token.Status=0
			token.ErrInfo=""
			json.NewEncoder(w).Encode(token)
		}

	}else{
//		http.Redirect(w,r,"/",http.StatusFound)
		http.NotFound(w,r)
	}
}

func GetUserFunc(w http.ResponseWriter, r *http.Request){
}
