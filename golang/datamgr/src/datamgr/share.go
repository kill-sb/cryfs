package main

import(
	"fmt"
	"os"
	"strings"
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
	if strings.HasSuffix(ipath,".csd") || strings.HasSuffix(ipath,".CSD"){
		fmt.Println("type: CSDFILE")
		return core.CSDFILE
	}else if core.IsValidUuid(strings.TrimSuffix(ipath,".tag"))|| core.IsValidUuid(strings.TrimSuffix(ipath,".TAG")) {
		fmt.Println("type: RAWDATA")
		return core.RAWDATA
	}else{
		return core.UNKNOWN
	}
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
	if fromtype==core.UNKNOWN{
		fmt.Println("Unknown data file format")
		return errors.New("Unkonwn data type")
	}
	sinfo,err:=core.NewShareInfo(linfo,fromtype,ipath)
	if err!=nil{
		fmt.Println("new share info error:",err)
		return err
	}
	if fromtype==core.RAWDATA{
		// todo: judge Isdir
		dinfo,err:=GetEncDataFromDisk(linfo,ipath)
		if(err!=nil){
			fmt.Println("GetEncData error:",err)
			return err
		}
		DoEncodeInC(dinfo.EncryptingKey,sinfo.RandKey,sinfo.EncryptedKey,16)
//		fmt.Println("encrypted key in csd:",core.BinkeyToString(sinfo.EncryptedKey))
	}else{
		// todo: from a csdfile, decrypt key and encode with another random key
		head,err:=core.LoadShareInfoHead(ipath)
		if err!=nil{
			fmt.Println("Load share info during reshare error:",err)
			return err
		}
		ssinfo,err:=dbop.LoadShareInfo(head)
		if err!=nil{
			fmt.Println("Load share info from head error:",err)
			return err
		}
		if ssinfo.Perm==0{
			fmt.Println("The file is not permitted to share.")
			return errors.New("File forbit to reshare")
		}
		inlist:=false
		for _,user:=range ssinfo.Receivers{
			if linfo.Name==user{
				inlist=true
				break
			}
		}
		if !inlist{
			fmt.Println(linfo.Name,"is not in share user list")
			return errors.New("Not share user")
		}
		// access control check ok now
		sinfo.FromUuid=ssinfo.Uuid
		sinfo.IsDir=ssinfo.IsDir
		orgkey:=make([]byte,16)
		DoDecodeInC(ssinfo.EncryptedKey,ssinfo.RandKey,orgkey,16)
		DoEncodeInC(orgkey,sinfo.RandKey,sinfo.EncryptedKey,16)
	}
	if config==""{
		err=InputShareInfo(sinfo) // input share info from terminal
		if(err!=nil){
			fmt.Println(err)
			return err
		}
	}else{
		LoadShareInfoConfig(sinfo)
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

func LoadShareInfoConfig(sinfo* core.ShareInfo) error{
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
//	fmt.Println("input a brief description for the file to be shared:")
//	fmt.Scanf("%s",&sinfo.Descr)
	fmt.Println("input permission(0 for readonly, 1 for reshare:")
	fmt.Scanf("%d",&sinfo.Perm)
	sinfo.LeftUse=-1
	sinfo.MaxUse=-1;
	//sinfo.Expire=  set it later
	return nil
}
