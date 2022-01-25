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
		encinfo:=api.NewDataAck()
		w.Header().Set("Content-Type","application/json")
		err:=json.NewDecoder(r.Body).Decode(encinfo)
		if err!=nil{
			log.Println("Decode json error:",err)
			json.NewEncoder(w).Encode(encinfo)
			return
		}
	}else{
		http.NotFound(w,r)
	}
}

func GetDataInfoFunc(w http.ResponseWriter, r *http.Request){
}


func ShareDataFunc(w http.ResponseWriter, r *http.Request){
}

