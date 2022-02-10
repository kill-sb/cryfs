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
	if r.Method=="POST"{
		ifack:=api.NewDataInfoAck()
		w.Header().Set("Content-Type","application/json")
		var difreq api.GetDataInfoReq
		err:=json.NewDecoder(r.Body).Decode(&difreq)
		if err!=nil{
			log.Println("Decode json error:",err)
			json.NewEncoder(w).Encode(ifack)
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

		retdata,err:=dbop.GetEncDataInfo(difreq.Uuid)
		if err!=nil{
			ifack.Code=2
			ifack.Msg=err.Error()
		}else{
			ifack.Code=0
			ifack.Msg="OK"
			ifack.Data=retdata
		}
        json.NewEncoder(w).Encode(ifack)
	}else{
		http.NotFound(w,r)
	}
}


func GetShareInfoFunc(w http.ResponseWriter, r *http.Request){
	//if r.Method=="GET"{
	if r.Method=="POST"{
		sifack:=api.NewShareInfoAck()
		w.Header().Set("Content-Type","application/json")
		var sifreq api.ShareInfoReq
		err:=json.NewDecoder(r.Body).Decode(&sifreq)
		if err!=nil{
			log.Println("Decode json error:",err)
			json.NewEncoder(w).Encode(sifack)
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

		retdata,err:=dbop.GetShareInfoData(sifreq.Uuid)
		if err!=nil{
			sifack.Code=2
			sifack.Msg=err.Error()
		}else{
			sifack.Code=0
			sifack.Msg="OK"
			sifack.Data=retdata
		}
        json.NewEncoder(w).Encode(sifack)
	}else{
		http.NotFound(w,r)
	}
}


func ShareDataFunc(w http.ResponseWriter, r *http.Request){
	if r.Method=="POST"{
		shrack:=api.NewShareAck()
		w.Header().Set("Content-Type","application/json")
		var shrreq api.ShareDataReq
		err:=json.NewDecoder(r.Body).Decode(&shrreq)
		if err!=nil{
			log.Println("Decode json error:",err)
			json.NewEncoder(w).Encode(shrack)
			return
		}
		luinfo,err:=GetLoginUserInfo(shrreq.Token)
		if err!=nil{
			shrack.Code=1
			shrack.Msg=err.Error()
			json.NewEncoder(w).Encode(shrack)
			return
		}
		if luinfo.Id!=shrreq.Data.OwnerId{
			shrack.Code=2
			shrack.Msg="Invalid user"
			json.NewEncoder(w).Encode(shrack)
			return
		}

		// user info checked ok
		// reference crypt.go:dbop.SaveMeta
		if err=dbop.WriteShareInfo(shrreq.Data);err!=nil{
			shrack.Code=1
			shrack.Msg=err.Error()
		}else{
			shrack.Code=0
			shrack.Msg="OK"
		}
		json.NewEncoder(w).Encode(shrack)
	}else{
		http.NotFound(w,r)
	}
}

func UpdateDataFunc(w http.ResponseWriter, r *http.Request){
	if r.Method=="POST"{
		udack:=api.NewUpdateDataAck()
		w.Header().Set("Content-Type","application/json")
		var udreq api.UpdateDataInfoReq
		err:=json.NewDecoder(r.Body).Decode(&udreq)
		if err!=nil{
			log.Println("Decode json error:",err)
			json.NewEncoder(w).Encode(udack)
			return
		}
		luinfo,err:=GetLoginUserInfo(udreq.Token)
		if err!=nil{
			udack.Code=1
			udack.Msg=err.Error()
			json.NewEncoder(w).Encode(udack)
			return
		}
		dinfo,err:=dbop.GetEncDataInfo(udreq.Uuid)
		if err!=nil{
			udack.Code=2
			udack.Msg=err.Error()
			json.NewEncoder(w).Encode(udack)
			return
		}
		if luinfo.Id!=dinfo.OwnerId{
			udack.Code=2
			udack.Msg="Invalid user"
			json.NewEncoder(w).Encode(udack)
			return
		}

		// user info checked ok
		// reference crypt.go:dbop.SaveMeta
		if err=dbop.UpdateMeta(&udreq);err!=nil{
			udack.Code=1
			udack.Msg=err.Error()
		}else{
			udack.Code=0
			udack.Msg="OK"
		}
		json.NewEncoder(w).Encode(udack)
	}else{
		http.NotFound(w,r)
	}
}

