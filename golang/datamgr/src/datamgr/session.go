package main

import (
	"fmt"
	"net"
	"unsafe"
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

const char *get_passwd(char *buf,int len)
{
	char *pass;
	int i;
	memset(buf,0,len);
	pass=getpass("Input passwd:");
	for (i=0;i<len && pass[i]!='\0'; i++)
		buf[i]=pass[i];
	return buf;
}*/
import "C"

type LoginInfo struct{
	Conn net.Conn
	Name string
	Id int
	Keylocalkey string
}

func (*LoginInfo) Logout() error{
	return  nil
}

func do_login(user string,passwd []byte)error{
	return nil
}

func Login(user string)(*LoginInfo, error){
	linfo:=new (LoginInfo)
	linfo.Name=user
	passwd:=make([]byte,16) // max 16 bytes password
	cpasswd:=(*C.char)(unsafe.Pointer(&passwd[0]))
	C.get_passwd(cpasswd,16)
	fmt.Println("Passwd: ",string(passwd))
	if err:=do_login(user,passwd);err!=nil{
		fmt.Println("Login error:",err)
		return nil,err
	}
	linfo.Keylocalkey=string(passwd) // just for test, Keylocalkey is used to encrypt random key
	return linfo,nil
}
