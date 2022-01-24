package main

import (
	"net/http"
	"time"
	"encoding/json"
	"fmt"
	"errors"
	"crypto/sha256"
	"crypto/md5"
	"dbop"
	"log"
	core "coredata"
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
/*
type AuthInfo struct{
	Name string `json:"name"`
	Passwd string `json:"passwd"`
	PriMask int32 `json:"primask"`
}

type TokenInfo struct{
	Id int32 `json:"id"`
	Token string `json:"token"`
	Key	string `json:"key"`
	Status int32 `json:"retval"`
	ErrInfo	string `json:"errinfo"`
}

type LoginUserInfo struct{
	Name string
	Id int32
	Keylocalkey []byte
	PriMask int32
	LogExpire time.Time
	Email string
	Descr string
	RegTime time.Time
	Mobile string
}*/

func GenUserInfo(ainfo* core.AuthInfo, tinfo *core.TokenInfo)*core.LoginUserInfo{
	luinfo:=new (core.LoginUserInfo)
	luinfo.Name=ainfo.Name
	luinfo.Id=tinfo.Id
	luinfo.Keylocalkey=core.StringToBinkey(tinfo.Key)
	luinfo.PriMask=ainfo.PriMask
	luinfo.UpdateToken()
	return luinfo
}

func NewToken()*core.TokenInfo{
	token:=&core.TokenInfo{Id:-1,Token:"",Key:"",Status:-1,ErrInfo:"Error Parameter"}
	return token
}

func GetLoginUserInfo(token string)(*core.LoginUserInfo,error){
	if ti,ok:=tokenmap[token];ok{
		// not expired 
		if time.Now().After(ti.LogExpire){
			return nil,errors.New("Token expired")
		}else{
			ti.UpdateToken() // should update here?
			return ti,nil
		}
	}else{
		return nil,errors.New("Token not found")
	}
}

/*
func (token *TokenInfo)GetUserInfo()(info* AuthInfo)error{
	// if return nil, means authoration succeeded,all fields should be fill correctly according to AuthInfo
	return nil
}*/

func makeToken(id int32)string{
	result:=id^XOR_BASE
	data:=core.UIntToBytes(uint64(result))
	sum:=md5.Sum(data)
	return fmt.Sprintf("%x",sum)
}

func LoginFunc(w http.ResponseWriter, r *http.Request){
	if r.Method=="POST"{
		token:=NewToken()
		w.Header().Set("Content-Type","application/json")
		var ainfo core.AuthInfo
		err:= json.NewDecoder(r.Body).Decode(&ainfo)
		if err!=nil{
			fmt.Println("Decode json error:",r.Body,"-",err)
			return
		}
		log.Println(ainfo)
//		token.GetUserInfo(&ainfo)
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
		if shastr==shasum{// password check ok
			token.Id=id
			token.Key=key
			token.Token=makeToken(token.Id)
			luinfo:=GenUserInfo(&ainfo,token)
			tokenmap[token.Token]=luinfo
		//	(time.Now().Unix()<<32)|(int64(id<<8))|(int64(ainfo.PriMask&0xff))
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
