package main

import(
	"fmt"
	"os"
	"strings"
	"errors"
	api "apiv1"
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
	if loginuser==""{
		fmt.Println("use parameter -user=NAME to set login user")
		return
	}
	linfo,err:=Login(loginuser)
	if err!=nil{
		fmt.Println("login error:",err)
		return
	}
	defer Logout(linfo)

	if info.IsDir(){
		shareDir(inpath,outpath,linfo)
	}else{
		shareFile(inpath,outpath,linfo)
	}
}

func GetDataType(ipath string /* .tag or .csd stand for local encrypted data or data shared from other user */) int{

    // should be replaced later because of multi-source processing

	if strings.HasSuffix(ipath,".csd") || strings.HasSuffix(ipath,".CSD"){
		return core.CSDFILE
	}else if core.IsValidUuid(strings.TrimSuffix(ipath,".tag"))|| core.IsValidUuid(strings.TrimSuffix(ipath,".TAG")) {
		return core.ENCDATA
	}else{
		return core.UNKNOWN
	}
}

func shareDir(ipath,opath string, linfo *core.LoginInfo){
/*	fromtype shoud be ENCDATA
	0. write shareinfo header
	1. zip path to file after header
	2. rename the file to ofile
*/
	sinfo,err:=core.NewShareInfo(linfo,core.ENCDATA,ipath)
	dinfo,_,err:=GetEncDataFromDisk(linfo,ipath)
	if err!=nil{
		fmt.Println("GetEncDataFromDisk in shareDir error:",err)
		return
	}

	if dinfo.OwnerId!=linfo.Id{
		fmt.Println("The data does not belong to",linfo.Name,dinfo.OwnerId,linfo.Id)
		return
	}
	DoEncodeInC(dinfo.EncryptingKey,sinfo.RandKey,sinfo.EncryptedKey,16)

	if config==""{
		err=InputShareInfo(sinfo) // input share info from terminal
		if(err!=nil){
			fmt.Println(err)
			return
		}
	}else{
		LoadShareInfoConfig(sinfo)
	}
	sinfo.CrTime=core.GetCurTime()
	sign,err:=WriteShareInfo(linfo.Token,sinfo)
	if err!=nil{
		fmt.Println(err)
		return
	}
	st,err:=os.Stat(opath)
	dst:=opath
	if err==nil && st.IsDir(){
		dst=opath+"/"+sinfo.Uuid+".csd"
	}
	err=CreateCSDFile(sinfo,sign,dst)
	if err!=nil{
		return
	}
	fmt.Println(dst," created ok, you can share it to ", sinfo.Receivers)
	return

}
/*
func LoadShareInfoFromTag(ipath string)(*core.ShareInfo,error){
	head,err:=core.LoadShareInfoHead(ipath)
	if err!=nil{
		fmt.Println("Load share info during reshare error:",err)
		return nil, err
	}
	return GetShareInfoFromHead(head)
}
*/
func shareFile(ipath,opath string, linfo *core.LoginInfo)error {
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
	if fromtype==core.ENCDATA{
		// todo: judge Isdir
		dinfo,_,err:=GetEncDataFromDisk(linfo,ipath)
		if(err!=nil){
			fmt.Println("GetEncData error:",err)
			return err
		}
		if dinfo.OwnerId!=linfo.Id{
			fmt.Println("The data does belong to",linfo.Name,dinfo.OwnerId,linfo.Id)
			return errors.New("incorrect user")
		}
		sinfo.OrgName=dinfo.OrgName
		DoEncodeInC(dinfo.EncryptingKey,sinfo.RandKey,sinfo.EncryptedKey,16)
//		fmt.Println("encrypted key in csd:",core.BinkeyToString(sinfo.EncryptedKey))
	}else{
		// todo: from a csdfile, decrypt key and encode with another random key
		head,err:=core.LoadShareInfoHead(ipath)
		if err!=nil{
			fmt.Println("Load share info during reshare error:",err)
			return err
		}
		ssinfo,err:=GetShareInfoFromHead(head,linfo)
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
		sinfo.OrgName=ssinfo.OrgName
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
	sinfo.CrTime=core.GetCurTime()
	sign,err:=WriteShareInfo(linfo.Token,sinfo)
	if err!=nil{
		fmt.Println(err)
		return err
	}
	st,err:=os.Stat(opath)
	dst:=opath
	if err==nil && st.IsDir(){
		dst=opath+"/"+sinfo.Uuid+".csd"
	}
	err=CreateCSDFile(sinfo,sign,dst) // local or remote uri will be processed in diffrent way in CreateCSDFile
	if err!=nil{
		return err
	}
	fmt.Println(dst," created ok, you can share it to ", sinfo.Receivers)
	return nil
}

func LoadShareInfoConfig(sinfo* core.ShareInfo) error{
	// load from config file, to be done
	return nil
}


func GetVisitors(recvlist string) ([]string,[]int32,error){
    strret:=strings.Split(recvlist,",")
	uinfo,err:=FindUserName_API(strret)
	if err!=nil{
		return nil,nil,err
	}
	intret:=make([]int32,0,len(strret))
	for  _,v:=range uinfo{
		if v.Id!=-1{
			intret=append(intret,v.Id)
		}else{
			return nil,nil,errors.New(fmt.Sprintf("Name:%s not found",v.Name))
		}
	}
    return strret,intret,nil
}

func InputShareInfo(sinfo *core.ShareInfo) error{
	// fill: descr,perm,expire,maxuse/leftuse
	var recvlist string
	fmt.Println("\nInput receivers(seperate with ','):")
	fmt.Scanf("%s",&recvlist)
	var err error
	sinfo.Receivers,sinfo.RcvrIds,err=GetVisitors(recvlist)
	if err!=nil{
		fmt.Println("Get receivers error:",err)
		return err
	}
//	fmt.Println("input a brief description for the file to be shared:")
//	fmt.Scanf("%s",&sinfo.Descr)
	fmt.Println("input permission(0 for readonly, 1 for reshare):")
	fmt.Scanf("%d",&sinfo.Perm)
	fmt.Println("expire time(Press 'Enter' for no expire time limit):")
	fmt.Scanf("%s",&sinfo.Expire)
	if sinfo.Expire==""{
		sinfo.Expire="2999:12:31 0:00:00"
	}
	fmt.Println("limit open times(-1 for no limit):")
	fmt.Scanf("%d",&sinfo.MaxUse)
	sinfo.LeftUse=sinfo.MaxUse
	//sinfo.Expire=  set it later
	return nil
}

func CreateCSDFile(sinfo *core.ShareInfo,sign []byte, dstfile string)error{
/*	fw,err:=os.Create(dstfile) // fixme: file mode should be assigned later
	if err!=nil{
		fmt.Println("CreateCSDFile error:",err)
			return err
	}
	defer fw.Close()
	if sinfo.FromType==ENCDATA{
		if WriteCSDHead(sinfo,sign,fw)==BINCONTENT{
			if sinfo.IsDir==0{
				fr,err:=os.Open(sinfo.FileUri)
				if err!=nil{
					fmt.Println("Open FileUri error:",err)
					return err
				}
				defer fr.Close()
				io.Copy(fw,fr)
			}else{
				ZipToFile(sinfo.FileUri,fw)
			}
		}else{
			fw.Write([]byte(sinfo.FileUri))
		}
	}else{
        if WriteCSDHead(sinfo,sign,fw)==BINCONTENT{
            fr,err:=os.Open(sinfo.FileUri)
            if err!=nil{
                fmt.Println("Open FileUri error:",err)
                return err
            }
            defer fr.Close()
			fr.Seek(60,0)
            io.Copy(fw,fr)
        }else{
            fw.Write([]byte(sinfo.FileUri))
        }
	}*/
	return nil
}


func WriteCSDHead(sinfo *core.ShareInfo, sign []byte,fw *os.File) error {
	/*
	head:=new (ShareInfoHeader)
	copy(head.MagicStr[:],[]byte("CMITFS"))
	copy(head.Uuid[:],[]byte(sinfo.Uuid))
	copy(head.EncryptedKey[:],sinfo.EncryptedKey)
	if IsLocalFile(sinfo.FileUri){
		head.ContentType=BINCONTENT
	}else{
		head.ContentType=REMOTEURL
	}
	head.IsDir=sinfo.IsDir
	buf:=new(bytes.Buffer)
	binary.Write(buf,binary.LittleEndian,head)
	fw.Write(buf.Bytes())
*/
/*	err,sha256:=GetFileSha256(sinfo.FileUri)
	if err!=nil{
		return err
	}
*/
	head:=new (core.ShareInfoHeader_V2)
	copy(head.MagicStr[:],[]byte("CSDFMTV2"))
	copy(head.Uuid[:],[]byte(sinfo.Uuid))
	copy(head.EncryptedKey[:],sinfo.EncryptedKey)
//	copy(head.Sha256[:],[]byte(sha256))
//	copy(head.Sign[:],sign)
	return nil
}

func GetShareInfoFromHead(head* core.ShareInfoHeader_V2,linfo* core.LoginInfo)(*core.ShareInfo,error){
	uuid:=string(head.Uuid[:])
//	enckey:=head.EncryptedKey[:]
	var err error
//	var apiack *api.IShareInfoAck
	var asinfo *api.ShareInfoData
	if linfo==nil{
		asinfo,err=GetShareInfo_Public_API(uuid)
	}else{
		asinfo,err=GetShareInfo_User_API(linfo.Token,uuid)
	}
	if err!=nil{
        fmt.Println("GetShareInfo_API error:",err)
        return nil,err
	}
	// TODO: Fill IsDir from remove server, remove ContentType
	sinfo:=FillShareInfo(asinfo,head.EncryptedKey[:])
	//sinfo:=FillShareInfo(apiack.Data,uuid,head.IsDir,int(head.ContentType),enckey)
	return sinfo,nil
}

func WriteShareInfo(token string, sinfo* core.ShareInfo)([]byte,error){
//func WriteShareInfo(token string, sinfo* core.ShareInfo)(error){
	err:=ShareData_API(token,sinfo)
	if err!=nil{
		return nil,err
	}
	return nil,nil // TODO: sign
}

func FillShareInfo(apidata *api.ShareInfoData, encryptedkey []byte)*core.ShareInfo{
    sinfo:=new (core.ShareInfo)
    sinfo.Uuid=apidata.Uuid
    sinfo.OwnerId=apidata.OwnerId
    sinfo.OwnerName,_=GetUserName(apidata.OwnerId)
    sinfo.Descr=apidata.Descr
    sinfo.Perm=apidata.Perm
    sinfo.Receivers=apidata.Receivers
    sinfo.RcvrIds=apidata.RcvrIds
    sinfo.Expire=apidata.Expire
    sinfo.MaxUse=apidata.MaxUse
    sinfo.LeftUse=apidata.LeftUse
    sinfo.RandKey=core.StringToBinkey(apidata.EncKey)
    sinfo.EncryptedKey=encryptedkey
    sinfo.FromType=apidata.FromType
    sinfo.FromUuid=apidata.FromUuid
//    sinfo.ContentType=ctype

	apidata.IsDir=sinfo.IsDir
    sinfo.CrTime=apidata.CrTime
    sinfo.OrgName=apidata.OrgName
    return sinfo
}


func FillShareReqData(sinfo *core.ShareInfo)*api.ShareInfoData{
	asi:=new (api.ShareInfoData)
	asi.Uuid=sinfo.Uuid
    asi.OwnerId=sinfo.OwnerId
  //  asi.OwnerName=sinfo.OwnerName
    asi.Descr=sinfo.Descr
    asi.Perm=sinfo.Perm
    asi.Receivers=sinfo.Receivers
    asi.RcvrIds=sinfo.RcvrIds
    asi.Expire=sinfo.Expire
    asi.MaxUse=sinfo.MaxUse
    asi.LeftUse=sinfo.LeftUse
    asi.EncKey=core.BinkeyToString(sinfo.RandKey)
    asi.FromType=sinfo.FromType
    asi.FromUuid=sinfo.FromUuid
    asi.CrTime=sinfo.CrTime
   // asi.FileUri=sinfo.FileUri
    asi.OrgName=sinfo.OrgName
	asi.IsDir=sinfo.IsDir
	return asi
}
