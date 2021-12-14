package coredata

import (
	"net"
	"time"
	"bytes"
	"errors"
	"io"
	"encoding/binary"
	"crypto/rand"
	"os/exec"
	"strings"
	"fmt"
	"os"
)

const (
    INVALID=iota
    ENCODE
	DECODE
    DISTRIBUTE
	TRACE
	LIST
    MOUNT
)

const (
    RAWDATA=iota
	CSDFILE
	UNKNOWN
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
    Conn net.Conn
    Name string
    Id int32
    Keylocalkey []byte
}

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
	Padding	[60] byte // 512-4--36-32-256-24-100=60
}

type EncryptedData struct{
    Uuid string
    Descr string
	IsDir	byte
    FromType int
    FromObj string
    OwnerId int32
	HashMd5 string
    EncryptingKey []byte
	Path	string
}

type ShareInfoHeader struct{ // .csd , cmit shared data
	MagicStr [6] byte // CMITFS
	Uuid	[36] byte // uuid used to search in db
	EncryptedKey	[16] byte // raw key encrypted with temp key(saved on server)
	ContentType byte //  0 for direct binary data, for remote file url
	IsDir byte //  data type(both for local and remote): 0 for single file, 1 for compressed dir packge
}	// 60 bytes total, should be placed in the end of file

type ShareInfo struct{
	Uuid string
	OwnerId int32
	Descr string
	Perm	int32
	Receivers []string
	RcvrIds	[]int32
	Expire	string // convert to time.Time later
	MaxUse	int32
	LeftUse	int32
	EncryptedKey	[]byte
	RandKey	[]byte
	FromType	int
	FromUuid	string
	ContentType int
	IsDir	byte // get info from database by uuid
	CrTime	string
	FileUri	string // source local filename or remote url
}

func GetCurTime()string{
	tm:=time.Now()
	return fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",tm.Year(),tm.Month(),tm.Day(),tm.Hour(),tm.Minute(),tm.Second())
}

func GetUuid()(string,error){
    if output,err:=exec.Command("uuidgen").Output();err!=nil{
        return "",err
    }else{
        return strings.TrimSpace(string(output)),nil
    }
}

func IsValidUuid(uuid string)bool{
	return true
}

func RandPasswd()([]byte,error){
    buf:=make([]byte,16)
    if rdlen,err:=rand.Read(buf);rdlen==len(buf) && err==nil{
        fmt.Println("randpasswd:",BinkeyToString(buf))
        return buf,nil
    }else {
        return nil,err
    }
}

func GetIsDirFromUuid(uuid string)byte{
	return 0 // todo: get from server/db by uuid later
}

func GetFileType(fname string)(int,error){
    finfo,err:=os.Stat(fname)
    if err!=nil{
        return -1,err
    }
    if IsValidUuid(finfo.Name()){
        _,err=os.Stat(fname+".tag")
        if err==nil{
            return RAWDATA,nil
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
	sinfo.Expire="2999:12:31 0:00:00"
	sinfo.MaxUse=-1
	sinfo.LeftUse=-1
	sinfo.FromType=fromtype
	sinfo.RandKey,_=RandPasswd()
	if fromtype==RAWDATA{
	//if fromtype!=UNKNOWN{
		st,_:=os.Stat(fromobj)
		sinfo.FromUuid=st.Name()
		if !IsValidUuid(sinfo.FromUuid){
			return nil,errors.New("Local encrypted file does not have a valid uuid filename")
		}
		sinfo.IsDir=GetIsDirFromUuid(sinfo.FromUuid)
	}else if fromtype==CSDFILE{
		// unknown file type: Uuid and IsDir will be filled according to source csd file outside
	}
	sinfo.FileUri=fromobj
	sinfo.EncryptedKey=make([]byte,16) // calc outside later
	return sinfo,nil
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

func (sinfo* ShareInfo)WriteFileHead(fw *os.File)byte /* ContentType*/{
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
	if sinfo.FromType==RAWDATA{
		if sinfo.WriteFileHead(fw)==BINCONTENT{
			fr,err:=os.Open(sinfo.FileUri)
			if err!=nil{
				fmt.Println("Open FileUri error:",err)
				return err
			}
			defer fr.Close()
			io.Copy(fw,fr)
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

func DataFromTag(tag *TagInFile) *EncryptedData{

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
	data.OwnerId=tag.OwnerId
	data.HashMd5=string(tag.Md5Sum[:])
	data.EncryptingKey=make([]byte,16)
	data.IsDir=tag.IsDir
	// EncryptedKey and Path will be filled later outside
	return data
}

func (tag *TagInFile) SaveTagToDisk(fname string)error{
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

func (tag *TagInFile) GetDataInfo()(*EncryptedData,error){
	if(tag.FromType==RAWDATA){
		return DataFromTag(tag),nil
	}else{
		fmt.Println("data from tag will be finished soon")
		return nil,nil
	}
}

func LoadTagFromDisk(fname string /* uuid file name*/)(*TagInFile,error){
	if !strings.HasSuffix(fname,".tag") && !strings.HasSuffix(fname,".TAG"){
		fname+=".tag"
	}
	f,err:=os.Open(fname)
	if err!=nil{
		return nil,err
	}
	defer f.Close()
	tag:=new(TagInFile)
	if err=binary.Read(f,binary.LittleEndian,tag);err==nil{
		return tag,nil
	}else{
		fmt.Println("decode error:",err)
		return nil,err
	}
}

func LoadShareInfoHead(fname string)(*ShareInfoHeader,error){
	fr,err:=os.Open(fname)
	if err!=nil{
		fmt.Println("Open file error",fname)
		return nil,err
	}
	defer fr.Close()

	head:=new (ShareInfoHeader)
	if err=binary.Read(fr,binary.LittleEndian,head);err!=nil{
		fmt.Println("Load share info head error",err)
		return nil,err
	}
	if string(head.MagicStr[:])=="CMITFS" && IsValidUuid(string(head.Uuid[:])){
		return head,nil
	}else{
		return nil,errors.New("Invalid csd file format")
	}

}

func (info *LoginInfo) Logout() error{
    return  nil
}

