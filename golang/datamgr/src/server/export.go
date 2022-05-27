package main

import (
	"net/http"
	"encoding/json"
	"log"
//	core "coredata"
//	"os"
	"dbop"
	api "apiv1"
)

func ExportDataFunc(w http.ResponseWriter, r *http.Request){
	if r.Method=="POST"{
		epack:=api.NewExProcAck()
		w.Header().Set("Content-Type","application/json")
		var epreq api.NewExportReq
		err:=json.NewDecoder(r.Body).Decode(&epreq)
		if err!=nil{
			Debug("Decode json error:",err)
			json.NewEncoder(w).Encode(epack)
			return
		}
        if g_config.Debug{
            DebugJson("Request:",&epreq)
            defer DebugJson("Response:",epack)
        }

		luinfo,err:=GetLoginUserInfo(epreq.Token)
		if err!=nil{
			epack.Code=1
			epack.Msg=err.Error()
			json.NewEncoder(w).Encode(epack)
			return
		}
/*		ownerid,err:=dbop.GetDataOwner(epreq.Data)
		if err!=nil{
            epack.Code=1
            epack.Msg=err.Error()
            json.NewEncoder(w).Encode(epack)
            return
        }

		if luinfo.Id!=ownerid{
			epack.Code=2
			epack.Msg="Invalid user"
			json.NewEncoder(w).Encode(epack)
			return
		}*/
		// user info checked ok
		if epack.Data,err=dbop.NewExport(epreq.Data,luinfo.Id,&epreq.Comment);err!=nil{
			epack.Code=1
			epack.Msg=err.Error()
			epack.Data=nil
		}else{
			epack.Code=0
			epack.Msg="OK"
		}
		json.NewEncoder(w).Encode(epack)
	}else{
		http.NotFound(w,r)
	}
}

func GetExportStatFunc(w http.ResponseWriter, r *http.Request){
	if r.Method=="POST"{
		epack:=api.NewExProcAck()
		w.Header().Set("Content-Type","application/json")
		var epreq api.ExportProcReq
		err:=json.NewDecoder(r.Body).Decode(&epreq)
		if err!=nil{
			Debug("Decode json error:",err)
			json.NewEncoder(w).Encode(epack)
			return
		}
        if g_config.Debug{
            DebugJson("Request:",&epreq)
            defer DebugJson("Response:",epack)
        }

		_,err=GetLoginUserInfo(epreq.Token)
        if err!=nil{
            epack.Code=1
            epack.Msg="You should login first"
            json.NewEncoder(w).Encode(epack)
            return
        }

		epinfo,err:=dbop.GetExportInfo(epreq.ExpId)
		if err!=nil{
			epack.Code=2
			epack.Data=nil
			epack.Msg=err.Error()
		}else{
	/*		if luinfo.Id!=epinfo.DstData.UserId{// fixme: check watiqueue owner later
				Debug("luid:",luinfo.Id,"epid:",epinfo.DstData.UserId)
				epack.Code=3
				epack.Msg="Data not belong to login user"
				epack.Data=nil
			}else{*/
				err=dbop.LoadProcQueue(epinfo)
				if err!=nil{
					epack.Code=2
					epack.Msg=err.Error()
				}else{
					epack.Code=0
					epack.Msg="OK"
					epack.Data=epinfo
				}
	//		}
		}
        json.NewEncoder(w).Encode(epack)
	}else{
		http.NotFound(w,r)
	}
}

// from/to uid, start/end time
func SearchExportsFunc(w http.ResponseWriter, r *http.Request){
	if r.Method=="POST"{
		w.Header().Set("Content-Type","application/json")
		var sereq api.SearchExpReq
		seack:=api.NewSearchExpAck()
		err:=json.NewDecoder(r.Body).Decode(&sereq)
		if err!=nil{
			log.Println("Decode json error:",err)
			json.NewEncoder(w).Encode(seack)
			return
		}

		uinfo,err:=GetLoginUserInfo(sereq.Token)
        if err!=nil{
            seack.Code=1
            seack.Msg="You should login first"
            json.NewEncoder(w).Encode(seack)
            return
        }
		if uinfo.Id!=sereq.ToUid && uinfo.Id!=sereq.FromUid{
			seack.Code=3
			seack.Msg="You should be either SENDER or RECEIVER of the exporting"
			json.NewEncoder(w).Encode(seack)
			return
		}

		objs,err:=dbop.SearchExpProc(&sereq)
		if err!=nil{
			seack.Code=2
			seack.Msg=err.Error()
		}else{
			seack.Data=objs
			seack.Code=0
			seack.Msg="OK"
		}
		json.NewEncoder(w).Encode(seack)
        if g_config.Debug{
            DebugJson("Request:",&sereq)
            DebugJson("Response:",seack)
        }
	}
}

func RespExportFunc(w http.ResponseWriter, r *http.Request){
	if r.Method=="POST"{
		reack:=api.NewSimpleAck()
		var rereq api.RespExpReq
		w.Header().Set("Content-Type","application/json")
		err:=json.NewDecoder(r.Body).Decode(&rereq)
		if err!=nil{
			log.Println("Decode json error:",err)
			json.NewEncoder(w).Encode(reack)
			return
		}
        if g_config.Debug{
            DebugJson("Request:",&rereq)
            defer DebugJson("Response:",reack)
        }
		luinfo,err:=GetLoginUserInfo(rereq.Token)
		if err!=nil{
			reack.Code=1
			reack.Msg=err.Error()
			json.NewEncoder(w).Encode(reack)
			return
		}

		// user info checked ok
		// reference crypt.go:dbop.SaveMeta
		if err=dbop.RespExportReq(luinfo.Id,&rereq);err!=nil{
			reack.Code=1
			reack.Msg=err.Error()
		}else{
			reack.Code=0
			reack.Msg="OK"
		}
		json.NewEncoder(w).Encode(reack)
	}else{
		http.NotFound(w,r)
	}
}

