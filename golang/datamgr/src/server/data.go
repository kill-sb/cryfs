package main

import (
	"net/http"
	"encoding/json"
	"log"
//	"os"
	_"dbop"
	api "apiv1"
)

func NewDataFunc(w http.ResponseWriter, r *http.Request){
	if r.Method=="POST"{
		encack:=api.NewDataAck()
		w.Header().Set("Content-Type","application/json")
		var encreq api.EncDataReq
		err:=json.NewDecoder(r.Body).Decode(&encreq)
		if err!=nil{
			log.Println("Decode json error:",err)
			json.NewEncoder(w).Encode(encack)
			return
		}
		luinfo,err:=GetLoginUserInfo(encreq.Token)
		if err!=nil{
			encack.Code=1
			encack.Msg=err.Error()
			json.NewEncoder(w).Encode(encack)
			return
		}
		// user info checked ok
		// reference crypt.go:dbop.SaveMeta
		log.Println(luinfo.Name)
	}else{
		http.NotFound(w,r)
	}
}

func GetDataInfoFunc(w http.ResponseWriter, r *http.Request){
}


func ShareDataFunc(w http.ResponseWriter, r *http.Request){
}

