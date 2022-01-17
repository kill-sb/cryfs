package main

import (
	"net/http"
	"encoding/json"
	"fmt"
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
	Status int32 `json:"retval"`
	ErrInfo	string `json:"errinfo"`
}

func NewToken()*TokenInfo{
	token:=&TokenInfo{Id:-1,Token:-1,Status:-1,ErrInfo:"Error Parameter"}
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
		token.GetUserInfo(&ainfo)
		// check user/passwd
	}else{
//		http.Redirect(w,r,"/",http.StatusFound)
		http.NotFound(w,r)
	}
}

func GetUserFunc(w http.ResponseWriter, r *http.Request){
}
