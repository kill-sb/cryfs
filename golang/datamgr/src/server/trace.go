package main

import (
    "net/http"
	"encoding/json"
	"log"
	api "apiv1"
    "dbop"
    "fmt"
   // "os"
)

func TraceBackFunc(w http.ResponseWriter, r *http.Request){
	if r.Method=="POST"{
		w.Header().Set("Content-Type","application/json")
		var tbreq api.TraceBackReq
		err:=json.NewDecoder(r.Body).Decode(&tbreq)
		tback:=api.NewTraceBackAck(len(tbreq.Data))
		if err!=nil{
			log.Println("Decode json error:",err)
			json.NewEncoder(w).Encode(tback)
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
		tback.Code=0
		tback.Msg="OK"
		for _,v:=range tbreq.Data{
			objs,err:=dbop.GetDataParent(&v)
			if err!=nil{
				tback.Code=3
				tback.Msg=fmt.Sprintf("search userid=%d error: %s",v,err.Error())
				break
			}else{
				tback.Data=append(tback.Data,objs)
			}
		}
		json.NewEncoder(w).Encode(tback)
	}else{
		http.NotFound(w,r)
	}

}

func TraceForwardFunc(w http.ResponseWriter, r *http.Request){
}


