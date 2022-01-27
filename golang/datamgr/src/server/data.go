package main

import (
	"net/http"
	"encoding/json"
	"log"
	core "coredata"
//	"os"
	"dbop"
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
		if luinfo.Id!=encreq.OwnerId{
			encack.Code=2
			encack.Msg="Invalid user"
			json.NewEncoder(w).Encode(encack)
			return
		}
		// user info checked ok
		// reference crypt.go:dbop.SaveMeta
		log.Println("newdata:",encreq)
		pdata:=new(core.EncryptedData)
	    pdata.Uuid=encreq.Uuid
	    pdata.Descr=encreq.Descr
	    pdata.FromType=encreq.FromType
	    pdata.FromObj=encreq.FromObj
	    pdata.OrgName=encreq.OrgName
	    pdata.OwnerId=encreq.OwnerId
	    pdata.EncryptingKey=core.StringToBinkey(encreq.EncKey)
	    pdata.IsDir=encreq.IsDir
		if err:=dbop.SaveMeta(pdata);err!=nil{
			encack.Code=2
			encack.Msg=err.Error()
		}else{
			encack.Code=0
			encack.Msg="OK"
		}
		json.NewEncoder(w).Encode(encack)
	}else{
		http.NotFound(w,r)
	}
}

func GetDataInfoFunc(w http.ResponseWriter, r *http.Request){
}


func ShareDataFunc(w http.ResponseWriter, r *http.Request){
}

