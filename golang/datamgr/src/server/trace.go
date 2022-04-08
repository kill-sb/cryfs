package main

import (
    "net/http"
	"encoding/json"
	"errors"
	api "apiv1"
    "dbop"
    "fmt"
	core "coredata"
   // "os"
)

func TraceBackFunc(w http.ResponseWriter, r *http.Request){
	if r.Method=="POST"{
		w.Header().Set("Content-Type","application/json")
		var tbreq api.TraceReq
		err:=json.NewDecoder(r.Body).Decode(&tbreq)
		tback:=api.NewTraceAck()
		if err!=nil{
			Debug("Decode json error:",err)
			json.NewEncoder(w).Encode(tback)
			return
		}
        if g_config.Debug{
            DebugJson("Request:",&tbreq)
            defer DebugJson("Response:",tback)
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

		objs,err:=dbop.TraceBack(tbreq.Data)
			//objs,err:=dbop.GetDataParent(&v)
		if err!=nil{
			tback.Code=3
			tback.Msg=fmt.Sprintf("search uuid=%s error: %s",tbreq.Data.Obj,err.Error())
		}else{
			tback.Data=objs
		}
		json.NewEncoder(w).Encode(tback)
	}else{
		http.NotFound(w,r)
	}
}

func TraceForwardFunc(w http.ResponseWriter, r *http.Request){
	if r.Method=="POST"{
		w.Header().Set("Content-Type","application/json")
		var tbreq api.TraceReq
		err:=json.NewDecoder(r.Body).Decode(&tbreq)
		tback:=api.NewTraceAck()
		if err!=nil{
			Debug("Decode json error:",err)
			json.NewEncoder(w).Encode(tback)
			return
		}
        if g_config.Debug{
            DebugJson("Request:",&tbreq)
            defer DebugJson("Response:",tback)
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

		objs,err:=dbop.TraceForward(tbreq.Data)
			//objs,err:=dbop.GetDataParent(&v)
		if err!=nil{
			tback.Code=3
			tback.Msg=fmt.Sprintf("search uuid=%s error: %s",tbreq.Data.Obj,err.Error())
		}else{
			tback.Data=objs
		}
		json.NewEncoder(w).Encode(tback)
	}else{
		http.NotFound(w,r)
	}


}

func TraceParentsFunc(w http.ResponseWriter, r *http.Request){
	if r.Method=="POST"{
		w.Header().Set("Content-Type","application/json")
		var tbreq api.TraceReq
		err:=json.NewDecoder(r.Body).Decode(&tbreq)
		tback:=api.NewTraceAck()
		if err!=nil{
			Debug("Decode json error:",err)
			json.NewEncoder(w).Encode(tback)
			return
		}
        if g_config.Debug{
            DebugJson("Request:",&tbreq)
            defer DebugJson("Response:",tback)
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

		objs,err:=dbop.GetDataParents(tbreq.Data)
			//objs,err:=dbop.GetDataParent(&v)
		if err!=nil{
			tback.Code=3
			tback.Msg=fmt.Sprintf("search uuid=%s error: %s",tbreq.Data.Obj,err.Error())
		}else{
			tback.Data=objs
		}
		json.NewEncoder(w).Encode(tback)
	}else{
		http.NotFound(w,r)
	}
}

func TraceChildrenFunc(w http.ResponseWriter, r *http.Request){
	if r.Method=="POST"{
		w.Header().Set("Content-Type","application/json")
		var tbreq api.TraceReq
		err:=json.NewDecoder(r.Body).Decode(&tbreq)
		tback:=api.NewTraceAck()
		if err!=nil{
			Debug("Decode json error:",err)
			json.NewEncoder(w).Encode(tback)
			return
		}
        if g_config.Debug{
            DebugJson("Request:",&tbreq)
            defer DebugJson("Response:",tback)
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

		objs,err:=dbop.GetDataChildren(tbreq.Data)
			//objs,err:=dbop.GetDataParent(&v)
		if err!=nil{
			tback.Code=3
			tback.Msg=fmt.Sprintf("search uuid=%s error: %s",tbreq.Data.Obj,err.Error())
		}else{
			tback.Data=objs
		}
		json.NewEncoder(w).Encode(tback)
	}else{
		http.NotFound(w,r)
	}

}

func QueryObjsFunc(w http.ResponseWriter, r *http.Request){
	if r.Method=="POST"{
		w.Header().Set("Content-Type","application/json")
		var qoreq api.QueryObjsReq
		err:=json.NewDecoder(r.Body).Decode(&qoreq)
		qoack:=api.NewQueryObjsAck(qoreq.Data)
		if err!=nil{
			Debug("Decode json error:",err)
			json.NewEncoder(w).Encode(qoack)
			return
		}
        if g_config.Debug{
            DebugJson("Request:",&qoreq)
            defer DebugJson("Response:",qoack)
        }
		/*
		_,err=GetLoginUserInfo(sifreq.Token)
        if err!=nil{
            sifack.Code=1
            sifack.Msg="You should login first"
            json.NewEncoder(w).Encode(sifack)
            return
        }*/
		qoack.Code=0
		qoack.Msg="OK"
		for k,v:=range qoreq.Data{
			var obj api.IFDataDesc=nil
			var err error=nil
			switch v.Type{
			case core.ENCDATA:
				obj,err=dbop.GetEncDataInfo(v.Obj)
			case core.CSDFILE:
				obj,err=dbop.GetShareInfoData(v.Obj)
			default:
				err=errors.New("Unknown obj type")
			}

			if err!=nil{
				qoack.Code=3
				qoack.Msg="query obj '"+qoreq.Data[k].Obj+"' error:"+err.Error()
				qoack.Data=[]api.IFDataDesc{}
				break
			}else{
				qoack.Data=append(qoack.Data,obj)
			}
		}
		json.NewEncoder(w).Encode(qoack)
	}else{
		http.NotFound(w,r)
	}

}
