package main

import (
	"fmt"
	"bytes"
//	"io/ioutil"
	"unsafe"
	"time"
	"errors"
//	"crypto/sha256"
	"crypto/tls"
	"net/http"
	"encoding/json"
	api "apiv1"
	core "coredata"
)

/*
#include <string.h>
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <errno.h>

int get_passwd(char *buf,int len)
{
	char *pass;
	int i;
	memset(buf,0,len);
	pass=getpass("Input passwd:");
	for (i=0;i<len && pass[i]!='\0'; i++)
		buf[i]=pass[i];
	return i;
}*/
import "C"
var APIServer string="https://127.0.0.1:8080/api/v1/"

func CallHttpAPI(param interface{},ret interface{},entry string)error{
	obj,_:=json.Marshal(param)
	fmt.Println("Sending:",string(obj))
	req,err:=http.NewRequest("POST",APIServer+entry,bytes.NewBuffer(obj))
	if err!=nil{
		fmt.Println("New request error:",err)
		return err
	}
    req.Header.Set("Content-Type","application/json")
//	tr:=&http.Transport{TLSClientConfig:&tls.Config{InsecureSkipVerify:false}}
	tr:=&http.Transport{TLSClientConfig:&tls.Config{InsecureSkipVerify:true}}
	client:=&http.Client{Transport:tr, Timeout:time.Second*10}
	resp,err:=client.Do(req)
	if err!=nil{
		fmt.Println("client do req error:",err)
		return err
	}
	defer resp.Body.Close()
	err= json.NewDecoder(resp.Body).Decode(ret)
	return err
}

func doAuth(user string)(*api.ITokenInfo,error){
    passwd:=make([]byte,16) // max 16 bytes password
    cpasswd:=(*C.char)(unsafe.Pointer(&passwd[0]))
    length:=C.get_passwd(cpasswd,16)
    passwd=passwd[:length]
	var ainfo api.AuthInfo
	ainfo.Name=user
	ainfo.Passwd=string(passwd)
	ainfo.PriMask=0
	token:=new (api.ITokenInfo)
	err:=CallHttpAPI(&ainfo,token,"login")
	if err==nil{
		fmt.Println("call api ok:",token,",data:",token.Data)
		return token,nil
	}else{
		fmt.Println("call api error:",err)
		return nil,err
	}
}
/*
func do_login(user string, passwd []byte)(*core.LoginInfo,error){
	if id,shasum,key,err:=dbop.LookupPasswdSHA(user);err!=nil{
		return nil,err
	}else{
		sharet:=sha256.Sum256(passwd)
		shastr:=""
		for _,ch:=range sharet{
			shastr=fmt.Sprintf("%s%02x",shastr,ch)
		}
//		fmt.Printf("login info: sharet %s, sha in db: %s\n",shastr,shasum)
		if	shastr==shasum{
			linfo:=&core.LoginInfo{Name:user,Id:id}
			linfo.Keylocalkey=core.StringToBinkey(key)
			return linfo,nil
		}
	}
	return nil,errors.New("Auth error")
}
func (linfo* LoginInfo)GetRawKey(src []byte)([]byte, error){
	if(linfo.Keylocalkey!=nil && len(linfo.Keylocalkey)!=0){
		srclen:=len(src)
		dst:=make([]byte,srclen)
		csrc:=(*C.char)(unsafe.Pointer(&src[0]))
		cdst:=(*C.char)(unsafe.Pointer(&dst[0]))
		cpasswd:=(*C.char)(unsafe.Pointer(&linfo.Keylocalkey[0]))
		C.decode(csrc,cpasswd,cdst,srclen)
		return dst,nil
	}
	return nil,errors.New("Load key for decrypt localkey error")
}
*/

func Login(user string)(*core.LoginInfo, error){
	token,err:=doAuth(user)
	if err!=nil{
		return nil,err
	}
	if token.Code!=0{
		return nil,errors.New(token.Msg)
	}
	linfo:=new(core.LoginInfo)
	linfo.Name=user
	linfo.Id=token.Data.Id
	linfo.Token=token.Data.Token
	linfo.Keylocalkey=core.StringToBinkey(token.Data.Key)
	return linfo,nil
//	linfo:=new (LoginInfo)
//	linfo.Name=user
/*	passwd:=make([]byte,16) // max 16 bytes password
	cpasswd:=(*C.char)(unsafe.Pointer(&passwd[0]))
	length:=C.get_passwd(cpasswd,16)
	passwd=passwd[:length]
	//if linfo,err:=do_login(user,passwd);err!=nil{
	if linfo,err:=do_api_login(user,passwd);err!=nil{
		fmt.Println("Login error:",err)
		return nil,err
	}else{
	//	linfo.Keylocalkey=string(passwd) // just for test, Keylocalkey is used to encrypt random key
		return linfo,nil
	}*/
}
