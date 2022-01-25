package main

import (
	"time"
	"fmt"
	"errors"
	"crypto/md5"
	//"dbop"
	//"log"
	core "coredata"
/*	"os"
	_"dbop"
	*/
)

const(
	EXPIRE_TIME=15*60
	WAKEUP_TIME=3*60
)

func TokenCacheMgr(){
	tm:=time.NewTimer(time.Second*WAKEUP_TIME)
	for{
		<-tm.C
		fmt.Println("Cur:",len(tokenmap))
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
		fmt.Println("Left:",len(tokenmap))
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

