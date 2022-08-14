package coredata

import (
//	"net"
	"time"
	"bytes"
	"errors"
//	"io"
	"encoding/binary"
	"crypto/rand"
	"os/exec"
	"regexp"
	"strings"
	"fmt"
	"os"
	api "apiv1"
)

const (
    INVALID=iota
    ENCODE
	DECODE
    DISTRIBUTE
	TRACE
	LIST
    MOUNT
	LOGIN
	VERSION
)

const (
    RAWDATA=iota-1
	ENCDATA
	CSDFILE
	UNKNOWN=0xff
)

const (
	SINGLEFILE=iota
	ZIPDIR
)

const (
	BINCONTENT=iota
	REMOTEURL
)

type LoginInfo struct{
	Token string
    Name string
    Id int32
    Keylocalkey []byte
}
/*
type TagInFile struct{ // .tag
	OwnerId int32
	Uuid	[36] byte //36
	Md5Sum	[32] byte //32
	FromType byte
	IsDir	byte
	FromObj [254] byte //254
	Time	int64
	EKey	[16] byte //16
	Descr	[100] byte // 100
	OrgName [254] byte
}*/

type EncDataTag struct{
	Uuid	[36] byte
	EKey	[16] byte
}
/*
type SourceObj struct{
	DataType int
	DataUuid string
}

type ImportFile struct{
	RelName string
	FileDesc string
	Sha256	string
	Size	int64
}

type RunContext struct{
	Id int64
	UserId int32
	InputData []SourceObj
	ImportPlain []ImportFile
	OS string
	BaseImg string
	OutputUuid string
	StartTime string
	EndTime string
}*/

type EncryptedData struct{
    Uuid string
    Descr string
	IsDir	byte
	OwnerName string

    FromRCId int64 // 0 means from local plain, >0 indicates a run-context id in database, from which FromContext can be created, 

	// if FromRCId==0, OrgName is plain data filename/dirname; otherwize, it comes from a mount operation, and a new name(with -dataname cmdline parameter should be set as OrgName)
	OrgName string
	FromContext *api.RCInfo

    OwnerId int32
    EncryptingKey []byte
	Path	string
	CrTime	string
}
/*
type ShareInfoHeader struct{ // .csd , cmit shared data
	MagicStr [6] byte // CMITFS
	Uuid	[36] byte // uuid used to search in db
	EncryptedKey	[16] byte // raw key encrypted with temp key(saved on server)
	ContentType byte //  0 for direct binary data, for remote file url
	IsDir byte //  data type(both for local and remote): 0 for single file, 1 for compressed dir packge
}	// 60 bytes total, should be placed in the end of file
*/

type ShareInfoHeader_V2 struct{ // .csd , cmit shared data
	MagicStr [8] byte // CSDFMTV2
	Uuid	[36] byte // uuid used to search in db
	EncryptedKey	[16] byte // raw key encrypted with temp key(saved on server)
//	Sha256	[64] byte // file content sha256 sum
//	Sign	[512] byte // 2048-bit / 8 (8 bits per byte)  / 2 (each byte need 2 ascii-char)
}	// 60/640 bytes total, .csd file header

const CSDV2HDSize=60



type ShareInfo struct{
	Uuid string
	OwnerId int32
	OwnerName string
	Descr string
	Perm	int32
	Sha256 string
	Receivers []string
	RcvrIds	[]int32
	Expire	string // convert to time.Time later
	MaxUse	int32
	LeftUse	int32
	EncryptedKey	[]byte
	RandKey	[]byte
	FromType	int
	FromUuid	string
//	ContentType int
	IsDir	byte
	CrTime	string
	FileUri	string // source local filename or remote url
	OrgName string
}

func IsUbt()bool{
	f,err:=os.Open("/etc/issue")
	if err!=nil{
		return true
	}
	defer f.Close()
	var os string
	fmt.Fscanf(f,"%s",&os)
	return strings.Contains(strings.ToLower(os),"ubuntu")
}

func GetCurTime()string{
	tm:=time.Now()
	return fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",tm.Year(),tm.Month(),tm.Day(),tm.Hour(),tm.Minute(),tm.Second())
}

func GetUuid()(string,error){
    if output,err:=exec.Command("uuidgen").Output();err!=nil{
        return "",err
    }else{
        return strings.ToLower(strings.TrimSpace(string(output))),nil
    }
}

func GetSelfPath()string{
    pid:=os.Getpid()
    exe:=fmt.Sprintf("/proc/%d/exe",pid)
    rname,err:=os.Readlink(exe)
    if err!=nil{
        return ""
    }
    finfo,err:=os.Stat(rname)
    if err!=nil{
        return ""
    }
    return strings.TrimSuffix(rname,finfo.Name())
}

func IsValidUuid(uuid string)bool{
    pat,_:=regexp.Compile("^[a-zA-Z0-9]{8}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{4}-[a-zA-Z0-9]{12}$")
    return pat.MatchString(uuid)
}

func RandPasswd()([]byte,error){
    buf:=make([]byte,16)
    if rdlen,err:=rand.Read(buf);rdlen==len(buf) && err==nil{
        return buf,nil
    }else {
        return nil,err
    }
}

func GetFileType(fname string)(int,error){
    finfo,err:=os.Stat(fname)
    if err!=nil{
        return -1,err
    }
    if IsValidUuid(finfo.Name()){
        _,err=os.Stat(fname+".tag")
        if err==nil{
            return ENCDATA,nil
        }
    }
    if !finfo.IsDir() && (strings.HasSuffix(fname,".csd") || strings.HasSuffix(fname,".CSD")){
        return CSDFILE,nil
    }
    return UNKNOWN,nil
}

func NewShareInfo(luser* LoginInfo,fromtype int, fromobj string /* need a local file, uuid named raw data or .csd format sharedfile */)(*ShareInfo,error){
	sinfo:=new (ShareInfo)
	// later register in db outside
	var err error
	sinfo.Uuid,err=GetUuid()
	if err!=nil{
		return nil,err
	}

	sinfo.OwnerId=luser.Id
	sinfo.Descr=""
	sinfo.Perm=-1
	sinfo.Receivers=nil
	sinfo.Expire="2999-12-31 00:00:00"
	sinfo.MaxUse=-1
	sinfo.LeftUse=-1
	sinfo.FromType=fromtype
	sinfo.RandKey,_=RandPasswd()
	if fromtype==ENCDATA{
	// if fromtype!=UNKNOWN {
		st,_:=os.Stat(fromobj)
		sinfo.FromUuid=st.Name()
		if !IsValidUuid(sinfo.FromUuid){
			return nil,errors.New("Local encrypted file does not have a valid uuid filename")
		}
		if st.IsDir(){
			sinfo.IsDir=1
		}else{
			sinfo.IsDir=0
		}
	}else if fromtype==CSDFILE{
		// unknown file type: Uuid, OrgName  and IsDir will be filled according to source csd file outside
		// seems nothing to do here
	}else{
		fmt.Println("Invalid fromtype:",fromtype)
		return nil,errors.New("Invalid fromtype in NewShareInfo")
	}
	sinfo.FileUri=fromobj
	sinfo.EncryptedKey=make([]byte,16) // calc outside later
	// TODO be sure: orgname and isdir is filled outsize, because new edtion removed these info in CSD File Header
	return sinfo,nil
}

func UIntToBytes(n int64) []byte {
    data := int64(n)
    bytebuf := bytes.NewBuffer([]byte{})
    binary.Write(bytebuf, binary.BigEndian, data)
    return bytebuf.Bytes()
}

func BytesToUInt(bys []byte) uint64 {
    bytebuff := bytes.NewBuffer(bys)
    var data uint64
    binary.Read(bytebuff, binary.BigEndian, &data)
    return uint64(data)
}

func BinkeyToString(binkey []byte)string{
	ret:=""
	for _,onebyte:=range binkey{
		ret+=fmt.Sprintf("%02x",onebyte)
	}
	return ret
}

func StringToBinkey(strkey string)[]byte{
	keylen:=len(strkey)/2
	ret:=make([]byte,keylen)
	for i:=0;i<keylen;i++{
		onebit:=fmt.Sprintf("%c%c",strkey[i*2],strkey[i*2+1])
		fmt.Sscanf(onebit,"%x",&ret[i])
	}
	return ret
}

func IsLocalFile(uri string)bool{
	return true // may be http,ftp,or nfs...later, then will return false
}

/*
func (sinfo* ShareInfo)WriteFileHead(fw *os.File)byte{
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
	return head.ContentType
}


func (sinfo *ShareInfo)CreateCSDFile(dstfile string)error{
	fw,err:=os.Create(dstfile) // fixme: file mode should be assigned later
	if err!=nil{
		fmt.Println("CreateCSDFile error:",err)
			return err
	}
	defer fw.Close()
	if sinfo.FromType==ENCDATA{
		if sinfo.WriteFileHead(fw)==BINCONTENT{
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
        if sinfo.WriteFileHead(fw)==BINCONTENT{
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
	}
	return nil
}


func DataFromTag(tag *EncDataTag) *EncryptedData{
// TODO Load Info from remote server
	data:=new(EncryptedData)
	data.Uuid=string(tag.Uuid[:])

	data.Descr=string(tag.Descr[:])
	data.FromType=int(tag.FromType)
//	data.FromObj=string(tag.FromObj[:])
	end:=254
	for i,v:=range tag.FromObj[:]{
		if v==0{
			end=i
			break
		}
	}
	data.FromObj=string(tag.FromObj[0:end])

	end=254
	for i,v:=range tag.OrgName[:]{
		if v==0{
			end=i
			break
		}
	}
	data.OrgName=string(tag.OrgName[0:end])

	data.OwnerId=tag.OwnerId
	data.HashMd5=string(tag.Md5Sum[:])
	data.EncryptingKey=make([]byte,16)
	data.IsDir=tag.IsDir
	// EncryptedKey and Path will be filled later outside
	return data
}
*/
func (tag *EncDataTag) SaveTagToDisk(fname string)error{
	fd,err:=os.Create(fname)
	if err!=nil{
		return err
	}
	defer fd.Close()
	buf:=new(bytes.Buffer)
	binary.Write(buf,binary.LittleEndian,tag)
	fd.Write(buf.Bytes())
	return nil
}

func LoadTagFromDisk(fname string /* uuid file name*/)(*EncDataTag,error){
	fname=strings.TrimSuffix(fname,"/")
	if !strings.HasSuffix(fname,".tag") && !strings.HasSuffix(fname,".TAG"){
		fname+=".tag"
	}
	f,err:=os.Open(fname)
	if err!=nil{
		return nil,err
	}
	defer f.Close()
	tag:=new(EncDataTag)
	if err=binary.Read(f,binary.LittleEndian,tag);err==nil{
		return tag,nil
	}else{
		fmt.Println("decode error:",err)
		return nil,err
	}
}

func LoadShareInfoHead(fname string)(*ShareInfoHeader_V2,error){
	fr,err:=os.Open(fname)
	if err!=nil{
		fmt.Println("Open file error",fname)
		return nil,err
	}
	defer fr.Close()

	head:=new (ShareInfoHeader_V2)
	if err=binary.Read(fr,binary.LittleEndian,head);err!=nil{
		fmt.Println("Load share info head error",err)
		return nil,err
	}
	if string(head.MagicStr[:])=="CSDFMTV2" && IsValidUuid(string(head.Uuid[:])){
		return head,nil
	}else{
		// TODO : check sign and SHA256
		return nil,errors.New("Invalid csd file format")
	}

}
/*
func (info *LoginInfo) Logout() error{ //  should be implemented later
	// TODO invoke logout API
    return  nil
}*/

/*
func FillShareInfo(apidata *api.ShareInfoData,uuid string,isdir byte, ctype int, encryptedkey []byte)*ShareInfo{
    sinfo:=new (ShareInfo)
    sinfo.Uuid=uuid
    sinfo.OwnerId=apidata.OwnerId
    sinfo.OwnerName=GetNameFromID(apidata.OwnerId)
    sinfo.Descr=apidata.Descr
    sinfo.Perm=apidata.Perm
    sinfo.Receivers=apidata.Receivers
    sinfo.RcvrIds=apidata.RcvrIds
    sinfo.Expire=apidata.Expire
    sinfo.MaxUse=apidata.MaxUse
    sinfo.LeftUse=apidata.LeftUse
    sinfo.RandKey=StringToBinkey(apidata.EncKey)
    sinfo.EncryptedKey=encryptedkey
    sinfo.FromType=apidata.FromType
    sinfo.FromUuid=apidata.FromUuid
    sinfo.ContentType=ctype
    sinfo.IsDir=isdir
    sinfo.CrTime=apidata.CrTime
 //   sinfo.FileUri=apidata.FileUri
    sinfo.OrgName=apidata.OrgName
    return sinfo
}

func FillEncDataInfo(adata *api.EncDataInfo)*EncryptedData{
    sinfo:=new (EncryptedData)
    sinfo.Uuid=adata.Uuid
    sinfo.OwnerId=adata.OwnerId
    sinfo.Descr=adata.Descr
    sinfo.FromType=adata.FromType
    sinfo.FromObj=adata.FromObj
    sinfo.HashMd5=adata.Hash256
    sinfo.IsDir=adata.IsDir
    sinfo.CrTime=adata.CrTime
    sinfo.OrgName=adata.OrgName
    return sinfo
}
*/
