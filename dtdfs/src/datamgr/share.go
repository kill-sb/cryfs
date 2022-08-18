package main

import(
	"fmt"
	"os"
	"bufio"
	"time"
	"io"
	"strings"
	"strconv"
	"unsafe"
	"errors"
	"bytes"
	"encoding/binary"
	api "apiv1"
	core "coredata"
)

func doShare(){
	if inpath==""{
        fmt.Println("You should set inpath explicitly")
        return
    }
    if outpath==""{
        fmt.Println("You should set outpath explicitly")
        return
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

	st,err:=os.Stat(ipath)
	if err==nil{
		fname:=st.Name()
		if strings.HasSuffix(fname,".csd") || strings.HasSuffix(fname,".CSD"){
			return core.CSDFILE
		}else if core.IsValidUuid(strings.TrimSuffix(fname,".tag"))|| core.IsValidUuid(strings.TrimSuffix(fname,".TAG")) {
			return core.ENCDATA
		}
	}
		return core.UNKNOWN
}

func ListContacts(token string){
	users,err:=GetContacts_API(token)
	if err==nil && len(users)>0{
		fmt.Println("Available contacts of current user:")
		for _,user:=range users{
			fmt.Print(user.Name,"  ")
		}
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
		fmt.Printf("The data does not belong to %s(userid:%d)\n",linfo.Name,linfo.Id)
		return
	}
	DoEncodeInC(dinfo.EncryptingKey,sinfo.RandKey,sinfo.EncryptedKey,16)

	if config==""{
		ListContacts(linfo.Token)
		err=InputShareInfo(sinfo) // input share info from terminal
		if(err!=nil){
			fmt.Println(err)
			return
		}
	}else{
		LoadShareInfoConfig(sinfo)
	}
   // sinfo.IsDir=1 has already been determined in NewShareInfo
    sinfo.OrgName=dinfo.OrgName

	st,err:=os.Stat(opath)
	dst:=opath
	if err==nil && st.IsDir(){
		dst=opath+"/"+sinfo.Uuid+".csd"
	}
	err=CreateCSDFile(sinfo,dst)
	if err!=nil{
		return
	}
	sinfo.Sha256,err=GetFileSha256(dst)
	if err!=nil{
		fmt.Println(err)
		os.RemoveAll(dst)
		return
	}
//	sinfo.CrTime=core.GetCurTime()
	err=WriteShareInfo(linfo.Token,sinfo)
	if err!=nil{
		fmt.Println(err)
		return
	}

	fmt.Println(dst," created ok, you can share it to ", sinfo.Receivers)
	return

}

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
		dinfo,_,err:=GetEncDataFromDisk(linfo,ipath)
		if(err!=nil){
			fmt.Println("GetEncData error:",err)
			return err
		}

		if dinfo.OwnerId!=linfo.Id{
			fmt.Printf("The data does not belong to %s(userid:%d)\n",linfo.Name,linfo.Id)
			return errors.New("incorrect user")
		}
		sinfo.OrgName=dinfo.OrgName
		DoEncodeInC(dinfo.EncryptingKey,sinfo.RandKey,sinfo.EncryptedKey,16)
//		fmt.Println("encrypted key in csd:",core.BinkeyToString(sinfo.EncryptedKey))
	}else{ // CSDFILE
		head,err:=core.LoadShareInfoHead(ipath)
		if err!=nil{
			fmt.Println("Load share info during reshare error:",err)
			return err
		}
		ssinfo,err:=GetShareInfoFromHead(head,linfo,1)
		if err!=nil{
			fmt.Println("Load share info from head error:",err)
			return err
		}
		sha256,_:=GetFileSha256(ipath)
		if ssinfo.Sha256!="" && strings.ToLower(sha256)!=strings.ToLower(ssinfo.Sha256){
			fmt.Println("Invalid sha256sum of ",ipath)
			return errors.New("Invalid sha256sum of csdfile")
		}
		if ssinfo.Perm==0 || ssinfo.MaxUse!=-1 || !strings.HasPrefix(ssinfo.Expire,"2999-12-31"){
			fmt.Println("The file is not permitted to share.")
			return errors.New("File forbit to reshare")
		}
		inlist:=false
		for _,userid:=range ssinfo.RcvrIds{
			if linfo.Id==userid{
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
		ListContacts(linfo.Token)
		err=InputShareInfo(sinfo) // input share info from terminal
		if(err!=nil){
			fmt.Println(err)
			return err
		}
	}else{
		LoadShareInfoConfig(sinfo) // TODO implement share config file
	}
	st,err:=os.Stat(opath)
	dst:=opath
	if err==nil && st.IsDir(){
		dst=opath+"/"+sinfo.Uuid+".csd"
	}
	err=CreateCSDFile(sinfo,dst) // local or remote uri will be processed in diffrent way in CreateCSDFile
	if err!=nil{
		return err
	}
	sinfo.Sha256,err=GetFileSha256(dst)
	if err!=nil{
		fmt.Println(err)
		os.RemoveAll(dst)
		return err
	}

//	sinfo.CrTime=core.GetCurTime()
	err=WriteShareInfo(linfo.Token,sinfo)
	if err!=nil{
		fmt.Println(err)
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
		intret=append(intret,v.Id)
	}
    return strret,intret,nil
}

func GetValidInput(vfun func (string)(string,error))(string,error){
	bio:=bufio.NewReader(os.Stdin)
	line,_,err:=bio.ReadLine()
	if err!=nil{
		return "",err
	}
	str:=strings.TrimSpace(string(line))
	if vfun!=nil{
		return vfun(str)
	}
	return str,nil
}

func InputShareInfo(sinfo *core.ShareInfo) error{
	// fill: descr,perm,expire,maxuse/leftuse
	fmt.Println("\nInput receivers(seperate with ','):")
//	fmt.Scanf("%s",&recvlist)
	recvlist,err:=GetValidInput(func (s string)(string,error){
	    if len(strings.Split(s," "))>1{
			return "",errors.New("Invalid whitespace")
		}
		return s,nil
	})
	if err!=nil{
		return err
	}
	sinfo.Receivers,sinfo.RcvrIds,err=GetVisitors(recvlist)
	if err!=nil{
		fmt.Println("Get receivers error:",err)
		return err
	}
//	fmt.Println("input a brief description for the file to be shared:")
//	fmt.Scanf("%s",&sinfo.Descr)
	fmt.Println("input permission: (Press 'Enter' use default choice 1 for reshare, input 0 for readonly)")
	input,err:=GetValidInput(func(s string)(string,error){
		if s=="" || s=="1" || s=="0"{
			return s,nil
		}
		return "",errors.New("Invalid perm code")
	})
	if err!=nil{
		return err
	}
	if input==""{
		sinfo.Perm=1
	}else{
		fmt.Sscanf(input,"%d",&sinfo.Perm)
	}
	fmt.Println("expire date: YYYY-MM-DD (Press 'Enter' for no expire time limit)")
	//fmt.Scanf("%s",&sinfo.Expire)
	sinfo.Expire,err=GetValidInput(func (s string)(string,error){
		if s==""{
			return s,nil
		}
        _,e:=time.Parse(time.RFC3339,s+"T00:00:00+08:00")
		if e!=nil{
			return "",errors.New("Invalid date format")
		}
		return s+" 23:59:59",nil
	})
	if err!=nil{
		return err
	}
	if sinfo.Expire==""{
		sinfo.Expire="2999-12-31 00:00:00"
	}
//	input=""
	fmt.Println("limit open times: (Press 'Enter' use default value -1 , means no limit)")
//	fmt.Scanf("%s",&input)
	input,err=GetValidInput(func(s string)(string,error){
		if s=="" {
			return s,nil
		}
		_,e:=strconv.ParseInt(s,10,32)
		if e!=nil{
			return "",errors.New("Invalid number format")
		}
		return s,nil
	})
	if err!=nil{
		return err
	}
	if input==""{
		sinfo.MaxUse=-1
	}else{
		fmt.Sscanf(input,"%d",&sinfo.MaxUse)
		if sinfo.MaxUse==0 || sinfo.MaxUse< -1 {
			return errors.New("Invalid open times")
		}
	}
	sinfo.LeftUse=sinfo.MaxUse
	return nil
}

func CreateCSDFile(sinfo *core.ShareInfo,dstfile string)error{
	fw,err:=os.Create(dstfile) // fixme: file mode should be assigned later
	if err!=nil{
		fmt.Println("CreateCSDFile error:",err)
			return err
	}
	defer fw.Close()
	if sinfo.FromType==core.ENCDATA{
		if _,err=WriteCSDHead(sinfo,fw);err==nil{
			if sinfo.IsDir==0{
				fr,err:=os.Open(sinfo.FileUri)
				if err!=nil{
					fmt.Println("Open FileUri error:",err)
					return err
				}
				defer fr.Close()
				io.Copy(fw,fr)
			}else{
				core.ZipToFile(sinfo.FileUri,fw)
			}
		}else{
			fw.Write([]byte(sinfo.FileUri))
		}
	}else{
		if hd,er:=WriteCSDHead(sinfo,fw); er==nil{
            fr,err:=os.Open(sinfo.FileUri)
            if err!=nil{
                fmt.Println("Open FileUri error:",err)
                return err
            }
            defer fr.Close()
			fr.Seek(int64(unsafe.Sizeof(*hd)),0)
			//fr.Seek(60,0)
            io.Copy(fw,fr)
        }else{
            fw.Write([]byte(sinfo.FileUri))
        }
	}
	return nil
}


func WriteCSDHead(sinfo *core.ShareInfo, fw *os.File)(*core.ShareInfoHeader_V2, error) {
	head:=new (core.ShareInfoHeader_V2)
	copy(head.MagicStr[:],[]byte("CSDFMTV2"))
	copy(head.Uuid[:],[]byte(sinfo.Uuid))
	copy(head.EncryptedKey[:],sinfo.EncryptedKey)
//	copy(head.Sha256[:],[]byte(sha256))
//	copy(head.Sign[:],sign)
	buf:=new(bytes.Buffer)
	binary.Write(buf,binary.LittleEndian,head)
	fw.Write(buf.Bytes())
	return head,nil
}

func GetShareInfoFromHead(head* core.ShareInfoHeader_V2,linfo* core.LoginInfo,needkey byte)(*core.ShareInfo,error){
	uuid:=string(head.Uuid[:])
//	enckey:=head.EncryptedKey[:]
	var err error
//	var apiack *api.IShareInfoAck
	var asinfo *api.ShareInfoData
	if linfo==nil{
		asinfo,err=GetShareInfo_Public_API(uuid)
	}else{
		asinfo,err=GetShareInfo_User_API(linfo.Token,uuid,needkey)
	}
	if err!=nil{
//        fmt.Println(err)
        return nil,err
	}
	// TODO: Fill IsDir from remove server, remove ContentType
	sinfo:=FillShareInfo(asinfo,head.EncryptedKey[:])
	//sinfo:=FillShareInfo(apiack.Data,uuid,head.IsDir,int(head.ContentType),enckey)
	return sinfo,nil
}

func WriteShareInfo(token string, sinfo* core.ShareInfo)error{
	if datadesc!=""{
		sinfo.Descr=datadesc
	}
	err:=ShareData_API(token,sinfo)
	if err!=nil{
		return err
	}
	return nil
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
	sinfo.Sha256=apidata.Sha256
//    sinfo.ContentType=ctype

	sinfo.IsDir=apidata.IsDir
	sinfo.CrTime=apidata.CrTime
    sinfo.OrgName=apidata.OrgName
    return sinfo
}


func FillShareReqData(sinfo *core.ShareInfo)*api.ShareInfoData{
	asi:=new (api.ShareInfoData)
	asi.Uuid=sinfo.Uuid
    asi.OwnerId=sinfo.OwnerId
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
	asi.Sha256=sinfo.Sha256
    asi.OrgName=sinfo.OrgName
	asi.IsDir=sinfo.IsDir
	return asi
}
