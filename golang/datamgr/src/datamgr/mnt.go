package main

import(
	"fmt"
	"os"
//	"strings"
//	"errors"
//	"dbop"
	core "coredata"
)

func doMount(){
	if inpath==""{
        fmt.Println("You should set inpath explicitly")
        return
    }

    info,err:=os.Stat(inpath)
	if err!=nil{
        fmt.Println("Can't find ",inpath)
        return
    }

    if outpath=="" {
        fmt.Println("You should set outpath explicitly")
		return
    }

	if loginuser==""{
		fmt.Println("use parameter -user=NAME to set login user")
		return
	}
	linfo,err:=Login(loginuser)
	if err!=nil{
		fmt.Println("login error:",err)
		return
	}
	defer linfo.Logout()

	if info.IsDir(){
		MountDir(inpath,outpath,linfo) // must be encrypted data, not shared file
	}else{
		MountFile(inpath,outpath,linfo) // may be a encrypted data or a csdfile(may also be a file or a dir)
	}
}

func MountDir(ipath, opath string, linfo *core.LoginInfo)error{
	return nil
}

func MountFile(ipath ,opath string, linfo *core.LoginInfo)error {
	return nil
}

