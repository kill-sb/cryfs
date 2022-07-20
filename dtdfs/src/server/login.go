package main

import (
	"net/http"
	"net"
	"encoding/json"
	"fmt"
	"crypto/sha256"
	"crypto/md5"
	"errors"
	"sync"
	"time"
	"dbop"
	"strings"
	"log"
	core "coredata"
	api "apiv1"
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

type LoginUserInfo struct{
    Name string
    Id int32
    Keylocalkey []byte
    PriMask int32
	Timeout int32
    LogExpire time.Time
	Lock sync.RWMutex
/*  Email string
    Descr string
    RegTime time.Time
    Mobile string
    */
}

const(
	EXPIRE_TIME=15*60
	WAKEUP_TIME=3*60
)

func (info* LoginUserInfo)RemoveToken(token string){
	tokenlock.Lock()
	if _,ok:=tokenmap[token];ok{
		delete(tokenmap,token)
	}
    tokenlock.Unlock()
}

func TokenCacheMgr(){
	tm:=time.NewTimer(time.Second*WAKEUP_TIME)
	for{
		<-tm.C
		cur:=time.Now()
		tokenlock.Lock()
		delist:=make([]string,0,len(tokenmap))
		for k,v:=range tokenmap{
			if cur.After(v.LogExpire){
				delist=append(delist,k)
			}
		}
		tokenlock.Unlock()
		tokenlock.Lock()
		for _,key:=range delist{
			if cur.After(tokenmap[key].LogExpire){
				delete(tokenmap,key)
			}
		}
		tokenlock.Unlock()
		tm.Reset(time.Second*WAKEUP_TIME)
	}
}

func AddToken(token string, luinfo *LoginUserInfo)error{
	// add to map, check map full,
	tokenlock.Lock()
	tokenmap[token]=luinfo
	tokenlock.Unlock()
	return nil
}

func (info* LoginUserInfo)UpdateToken(){
	info.Lock.Lock()
    info.LogExpire=time.Now().Add(time.Second*EXPIRE_TIME) // expire time 15 minite
	info.Lock.Unlock()
}
/*
func NewToken()*api.TokenInfo{
	token:=&api.TokenInfo{Id:-1,Token:"",Key:"",Status:-1,ErrInfo:"Error Parameter"}
	return token
}*/

func GetLoginUserInfo(token string)(*LoginUserInfo,error){
	tokenlock.RLock()
	ti,ok:=tokenmap[token]
	tokenlock.RUnlock()
	if ok{
		// cache clear routine may didn't clear it on time
		ti.Lock.RLock()
		after:=time.Now().After(ti.LogExpire)
		ti.Lock.RUnlock()
		if after{
			return nil,errors.New("Token expired")
		}else{
			ti.UpdateToken() // should update here?
			return ti,nil
		}
	}else{
		return nil,errors.New("Token not found")
	}
}

func makeToken(id int32)string{
	result:=id^XOR_BASE
	data:=core.UIntToBytes(int64(result)^time.Now().UnixNano())
	sum:=md5.Sum(data)
	return fmt.Sprintf("%x",sum)
}

func GenUserInfo(ainfo* api.AuthInfo, tinfo *api.TokenInfo)*LoginUserInfo{
	luinfo:=new (LoginUserInfo)
	luinfo.Name=ainfo.Name
	luinfo.Id=tinfo.Id
	luinfo.Keylocalkey=core.StringToBinkey(tinfo.Key)
	luinfo.PriMask=ainfo.PriMask
	luinfo.Timeout=tinfo.Timeout
	luinfo.UpdateToken()
	return luinfo
}

func LoginFunc(w http.ResponseWriter, r *http.Request){
	if r.Method=="POST"{
		token:=api.NewToken()
		w.Header().Set("Content-Type","application/json")
		var ainfo api.AuthInfo
		err:= json.NewDecoder(r.Body).Decode(&ainfo)
		if err!=nil{
			Debug("Decode json error:",r.Body,"-",err)
			json.NewEncoder(w).Encode(token)
			return
		}
		if g_config.Debug{
			DebugJson("Request:",&ainfo)
			defer DebugJson("Response:",token)
		}
//		log.Println("login:",ainfo)
//		token.GetUserInfo(&ainfo)
		// check user/passwd
		id,shasum,key,err:=dbop.LookupPasswdSHA(ainfo.Name)
		if err!=nil{
			json.NewEncoder(w).Encode(token)
			return
		}
		sharet:=sha256.Sum256([]byte(ainfo.Passwd))
		shastr:=""
		for _,ch:=range sharet{
			shastr=fmt.Sprintf("%s%02x",shastr,ch)
		}
		if shastr==shasum{// password check ok
			token.Data.Id=id
			token.Data.Key=key
			token.Data.Token=makeToken(token.Data.Id)
			token.Data.Timeout=EXPIRE_TIME
			luinfo:=GenUserInfo(&ainfo,token.Data)
			AddToken(token.Data.Token,luinfo)
		//	log.Println(*luinfo)
			token.Code=0
			token.Msg="OK"
		}else{
			token.Code=1
			token.Msg="Invalid user/password"
		}
		json.NewEncoder(w).Encode(token)
	}else{
//		http.Redirect(w,r,"/",http.StatusFound)
		http.NotFound(w,r)
	}
}


/*
func GetUserFunc(w http.ResponseWriter, r *http.Request){
	if r.Method=="POST"{
		usrack:=api.NewUserInfoAck()
		w.Header().Set("Content-Type","application/json")
		var usrreq api.GetUserReq
		err:=json.NewDecoder(r.Body).Decode(&usrreq)
		if err!=nil{
			Debug("Decode json error:",err)
			json.NewEncoder(w).Encode(usrack)
			return
		}
		if g_config.Debug{
			DebugJson("Request:",&usrreq)
			defer DebugJson("Response:",usrack)
		}
		usrack.Code=0
		usrack.Msg="OK"
		for _,v:=range usrreq.Id{
			usr,err:=dbop.GetUserInfo(v)
			if err!=nil{
				usrack.Code=3
				usrack.Msg=fmt.Sprintf("search userid=%d error: %s",v,err.Error())
				usrack.Data=[]api.UserInfoData{}
				break
			}else{
				usrack.Data=append(usrack.Data,*usr)
			}
		}
		json.NewEncoder(w).Encode(usrack)
	}else{
		http.NotFound(w,r)
	}
}

func FindUserNameFunc(w http.ResponseWriter, r *http.Request){
	if r.Method=="POST"{
		usrack:=api.NewUserInfoAck()
		w.Header().Set("Content-Type","application/json")
		var usrreq api.FindUserNameReq
		err:=json.NewDecoder(r.Body).Decode(&usrreq)
		if err!=nil{
			Debug("Decode json error:",err)
			json.NewEncoder(w).Encode(usrack)
			return
		}
        if g_config.Debug{
            DebugJson("Request:",&usrreq)
            defer DebugJson("Response:",usrack)
        }

		usrack.Code=0
		usrack.Msg="OK"
		for _,v:=range usrreq.Name{
			usr,err:=dbop.GetUserInfoByName(v)
			if err!=nil{
				usrack.Code=3
				usrack.Msg=fmt.Sprintf("search user %s error: %s",v,err.Error())
				usrack.Data=[]api.UserInfoData{}
				break
			}else{
				Debug(usr.Name,usr.Id)
				usrack.Data=append(usrack.Data,*usr)
			}
		}
		json.NewEncoder(w).Encode(usrack)
	}else{
		http.NotFound(w,r)
	}

}*/

func RefreshTokenFunc(w http.ResponseWriter, r *http.Request){
	if r.Method=="POST"{
		rfack:=api.NewLoginStatAck()
		w.Header().Set("Content-Type","application/json")
		var rfreq api.LoginStatReq
		err:=json.NewDecoder(r.Body).Decode(&rfreq)
		if err!=nil{
			log.Println("Decode json error:",err)
			json.NewEncoder(w).Encode(rfack)
			return
		}
        if g_config.Debug{
            DebugJson("Request:",&rfreq)
            defer DebugJson("Response:",rfack)
        }

		info,err:=GetLoginUserInfo(rfreq.Token)
        if err!=nil{
            rfack.Code=1
            rfack.Msg="Invalid token"
            json.NewEncoder(w).Encode(rfack)
            return
        }
		rfack.Code=0
		rfack.Msg="OK"
		rfack.Data.Timeout=info.Timeout
		json.NewEncoder(w).Encode(rfack)
	}else{
		http.NotFound(w,r)
	}
}

func HasLocalIPAddr(ipstr string) bool {
	ip:=net.ParseIP(ipstr)
/*	if ip.IsLoopback() {
		return true
	}
*/
	ip4 := ip.To4()
	if ip4 == nil {
		return false
	}

	return ip4[0] == 10 || // 10.0.0.0/8
		(ip4[0] == 172 && ip4[1] >= 16 && ip4[1] <= 31) || // 172.16.0.0/12
		(ip4[0] == 169 && ip4[1] == 254) || // 169.254.0.0/16
		(ip4[0] == 192 && ip4[1] == 168) // 192.168.0.0/16
}

func GetClientPublicIP(r *http.Request) string {
	var ip string
	for _, ip = range strings.Split(r.Header.Get("X-Forwarded-For"), ",") {
		ip = strings.TrimSpace(ip)
		if ip != "" && !HasLocalIPAddr(ip) {
			return ip
		}
	}

	ip = strings.TrimSpace(r.Header.Get("X-Real-Ip"))

	if ip != "" && !HasLocalIPAddr(ip) {
		return ip
	}

	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		/*
		if !HasLocalIPAddr(ip) {
			return ip
		}*/
		return ip
	}

	return ""
}

func LogoutFunc(w http.ResponseWriter, r *http.Request){
	if r.Method=="POST"{
		loack:=api.NewLoginStatAck()
		w.Header().Set("Content-Type","application/json")
		var loreq api.LoginStatReq
		err:=json.NewDecoder(r.Body).Decode(&loreq)
		if err!=nil{
			log.Println("Decode json error:",err)
			json.NewEncoder(w).Encode(loack)
			return
		}
        if g_config.Debug{
            DebugJson("Request:",&loreq)
            defer DebugJson("Response:",loack)
        }
		info,err:=GetLoginUserInfo(loreq.Token)
        if err!=nil{
            loack.Code=1
            loack.Msg="Invalid token"
            json.NewEncoder(w).Encode(loack)
            return
        }
		info.RemoveToken(loreq.Token)
		loack.Code=0
		loack.Msg="OK"
		json.NewEncoder(w).Encode(loack)
	}else{
		http.NotFound(w,r)
	}
}

