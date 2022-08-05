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
			epack.Data=nil
			epack.Code=api.ERR_BADPARAM
			json.NewEncoder(w).Encode(epack)
			return
		}
        if g_config.Debug{
            DebugJson("Request:",&epreq)
            defer DebugJson("Response:",epack)
        }

		luinfo,err:=GetLoginUserInfo(epreq.Token)
		if err!=nil{
			epack.Data=nil
			epack.Code=api.ERR_INVDATA
			epack.Msg=err.Error()
			json.NewEncoder(w).Encode(epack)
			return
		}
		switch epreq.Data.Type{
		case api.ENCDATA:
			ownerid,err:=dbop.GetDataOwner(epreq.Data)
			if err!=nil{
				epack.Data=nil
	            epack.Code=api.ERR_INVDATA
	            epack.Msg=err.Error()
	            json.NewEncoder(w).Encode(epack)
	            return
	        }
			if luinfo.Id!=ownerid {
				epack.Data=nil
				epack.Code=api.ERR_ACCESS
				epack.Msg="Not owner of the data"
				json.NewEncoder(w).Encode(epack)
				return
			}
		case api.CSDFILE:
			sinfo,err:=dbop.GetShareInfoData(epreq.Data.Obj)
			if err!=nil{
				epack.Data=nil
	            epack.Code=api.ERR_INVDATA
	            epack.Msg=err.Error()
	            json.NewEncoder(w).Encode(epack)
	            return
			}
			if luinfo.Id!=sinfo.OwnerId{
				inlist:=false
				for _,rid:=range sinfo.RcvrIds{
					if rid==luinfo.Id{
						inlist=true
						break
					}
				}
				if !inlist{
					epack.Data=nil
					epack.Code=api.ERR_ACCESS
					epack.Msg="Neither owner nor receiver of the data"
					json.NewEncoder(w).Encode(epack)
					return
				}
			}
		default:
			epack.Data=nil
			epack.Code=api.ERR_INVDATA
			epack.Msg=err.Error()
			json.NewEncoder(w).Encode(epack)
			return
		}

		// user info checked ok now
		if epack.Data,err=dbop.NewExport(epreq.Data,luinfo.Id,&epreq.Comment);err!=nil{
			epack.Code=api.ERR_INTERNAL
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
			epack.Data=nil
			epack.Code=api.ERR_BADPARAM
			json.NewEncoder(w).Encode(epack)
			return
		}
        if g_config.Debug{
            DebugJson("Request:",&epreq)
            defer DebugJson("Response:",epack)
        }

		luinfo,err:=GetLoginUserInfo(epreq.Token)
        if err!=nil{
			epack.Data=nil
            epack.Code=api.ERR_INVDATA
            epack.Msg="You should login first"
            json.NewEncoder(w).Encode(epack)
            return
        }

		epinfo,err:=dbop.GetExportInfo(epreq.ExpId)
		if err!=nil{
			epack.Code=api.ERR_INVDATA
			epack.Data=nil
			epack.Msg=err.Error()
		}else{
			err=dbop.LoadProcQueue(epinfo)
			if err!=nil{
				epack.Code=api.ERR_INTERNAL
				epack.Data=nil
				epack.Msg=err.Error()
			}else{
				if luinfo.Id!=epinfo.DstData.UserId{ // check receiver
					if epinfo.ProcQueue==nil || len(epinfo.ProcQueue)==0{
						epack.Code=api.ERR_ACCESS
						epack.Msg="Invalid query user"
						epack.Data=nil
					}else{ // now check receive list
						bfind:=false
						for _,node:=range epinfo.ProcQueue{
							if node.ProcUid==luinfo.Id{
								bfind=true
								break
							}
						}
						if bfind{
							epack.Code=0
							epack.Msg="OK"
							epack.Data=epinfo
						}else{
							epack.Code=api.ERR_ACCESS
							epack.Msg="Invalid query user"
							epack.Data=nil
						}
					}
				}else{
					// sender , if agreed , get enckey
					if epinfo.Status==api.AGREE && epinfo.DstData.Type==api.CSDFILE{
						sinfo,err:=dbop.GetUserShareInfoData(epinfo.DstData.Uuid,-1)
						if err==nil && sinfo!=nil{
							epinfo.EncKey=sinfo.EncKey
						}
					}

					epack.Code=0
					epack.Msg="OK"
					epack.Data=epinfo
				}
			}
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
			seack.Data=nil
			seack.Code=api.ERR_BADPARAM
			json.NewEncoder(w).Encode(seack)
			return
		}

        if sereq.StartItem<0 || sereq.MaxCount<0{
            seack.Data=nil
            seack.Code=api.ERR_INVDATA
            seack.Msg="Invalid search parameter"
            json.NewEncoder(w).Encode(seack)
            return
        }

		uinfo,err:=GetLoginUserInfo(sereq.Token)
        if err!=nil{
			seack.Data=nil
            seack.Code=api.ERR_INVDATA
            seack.Msg="You should login first"
            json.NewEncoder(w).Encode(seack)
            return
        }
		if uinfo.Id!=sereq.ToUid && uinfo.Id!=sereq.FromUid{
			seack.Data=nil
			seack.Code=api.ERR_ACCESS
			seack.Msg="You should be either SENDER or RECEIVER of the exporting"
			json.NewEncoder(w).Encode(seack)
			return
		}

		objs,err:=dbop.SearchExpProc(&sereq)
		if err!=nil{
			seack.Data=nil
			seack.Code=api.ERR_INTERNAL
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
			reack.Code=api.ERR_BADPARAM
			json.NewEncoder(w).Encode(reack)
			return
		}
        if g_config.Debug{
            DebugJson("Request:",&rereq)
            defer DebugJson("Response:",reack)
        }
		luinfo,err:=GetLoginUserInfo(rereq.Token)
		if err!=nil{
			reack.Code=api.ERR_INVDATA
			reack.Msg=err.Error()
			json.NewEncoder(w).Encode(reack)
			return
		}

		// user info checked ok
		// reference crypt.go:dbop.SaveMeta
		if err=dbop.RespExportReq(luinfo.Id,&rereq);err!=nil{
			reack.Code=api.ERR_INVDATA
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

