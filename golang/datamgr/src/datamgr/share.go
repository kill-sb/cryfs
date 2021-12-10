package main

import(
	"fmt"
	"os"
	"errors"
	"dbop"
	core "coredata"
)

func doShare(){
	if inpath==""{
        fmt.Println("You should set inpath explicitly")
        return
    }
    if outpath==""{
        outpath=inpath+".csd"
    }
    info,err:=os.Stat(inpath)
	if err!=nil{
        fmt.Println("Can't find ",inpath)
        return
    }
	if info.IsDir(){
		shareDir(inpath,outpath,loginuser)
	}else{
		shareFile(inpath,outpath,loginuser)
	}
}

func shareDir(ipath,opath,user string){
}

func GetDataType(ipath string /* .tag or .csd stand for local encrypted data or data shared from other user */) int{
	return core.RAWDATA
}

func shareFile(ipath,opath,user string)error {
	if user==""{
		fmt.Println("use parameter -user=NAME to set login user")
		return errors.New("empty user")
	}
	linfo,err:=Login(user)
	if err!=nil{
		fmt.Println("login error:",err)
		return err
	}
	defer linfo.Logout()
	fromtype:=GetDataType(ipath)
	sinfo,err:=core.NewShareInfo(linfo,fromtype,ipath)
	if err!=nil{
		fmt.Println("new share info error:",err)
		return err
	}
	if fromtype==core.RAWDATA{
		dinfo,err:=GetEncDataFromDisk(linfo,ipath)
		if(err!=nil){
			fmt.Println("GetEncData error:",err)
			return err
		}
		DoEncodeInC(dinfo.EncryptingKey,sinfo.RandKey,sinfo.EncryptedKey,16)
		fmt.Println("encrypted key in csd:",core.BinkeyToString(sinfo.EncryptedKey))
	}else{
		// todo: from a csdfile, decrypt key and encode with another random key
	}
	err=InputShareInfo(sinfo) // input share info from terminal
	if(err!=nil){
		fmt.Println(err)
		return err
	}
	err=dbop.WriteShareInfo(sinfo)
	if err!=nil{
		fmt.Println(err)
		return err
	}
	st,err:=os.Stat(opath)
	dst:=opath
	if err==nil && st.IsDir(){
		dst=opath+"/"+sinfo.Uuid+".csd"
	}
	err=sinfo.CreateCSDFile(dst) // local or remote uri will be processed in diffrent way in CreateCSDFile
	if err!=nil{
		return err
	}
	fmt.Println(dst," created ok, you can share it to ", sinfo.Receivers)
	return nil
}

func InputShareInfo(sinfo *core.ShareInfo) error{
	// fill: descr,perm,expire,maxuse/leftuse
	var recvlist string
	fmt.Println("\nInput receivers(seperate with ','):")
	fmt.Scanf("%s",&recvlist)
	var err error
	sinfo.Receivers,sinfo.RcvrIds,err=dbop.ParseVisitors(recvlist)
	if err!=nil{
		fmt.Println("Get receivers error:",err)
		return err
	}
	fmt.Println("input a brief description for the file to be shared:")
	fmt.Scanf("%s",&sinfo.Descr)
	fmt.Println("input visit count limit:")
	fmt.Scanf("%d",&sinfo.MaxUse)
	sinfo.LeftUse=sinfo.MaxUse;
	sinfo.Perm=-1 // all perms
	//sinfo.Expire=  set it later
	return nil
}
