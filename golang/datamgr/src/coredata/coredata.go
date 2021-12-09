package coredata

import (
	"net"
	"bytes"
	"encoding/binary"
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
    TAG
)

type LoginInfo struct{
    Conn net.Conn
    Name string
    Id int32
    Keylocalkey []byte
}

type TagInFile struct{
	OwnerId int32
	Uuid	[36] byte //36
	Md5Sum	[32] byte //32
	FromType byte
	FromObj [255] byte //255
	Time	int64
	EKey	[16] byte //16
	Descr	[100] byte // 100
	Padding	[60] byte // 512-4--36-32-256-24-100=60
}

type EncryptedData struct{
    Uuid string
    Descr string
    FromType int
    FromObj string
    OwnerId int32
	HashMd5 string
    EncryptedKey []byte
	Path	string
}

type ShareInfo struct{
	Uuid string
	OwnerId int32
	Descr string
	Perm	int32
	Receivers []string
	Expire	string // convert to time.Time later
	MaxUse	int32
	LeftUse	int32
	EncryptedKey	[]byte
	FromType	int
	FromUuid	string
	FileUri	string
}

func GetUuid()(string,error){
    if output,err:=exec.Command("uuidgen").Output();err!=nil{
        return "",err
    }else{
        return strings.TrimSpace(string(output)),nil
    }
}

func NewShareTag(luser* LoginInfo,fromtype int, fromobj string /* need a local file, uuid named raw data or .csd format sharedfile */,recvrs []string, )(*ShareInfo,error){
	sinfo:=new (ShareInfo)
	// later register in db outside
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

func (sinfo *ShareInfo)CreateShareFile(fullname string)error{
	return nil
}

func DataFromTag(tag *TagInFile) *EncryptedData{
	data:=new(EncryptedData)
	data.Uuid=string(tag.Uuid[:])
	data.Descr=string(tag.Descr[:])
	data.FromType=int(tag.FromType)
	data.FromObj=string(tag.FromObj[:])
	data.OwnerId=tag.OwnerId
	data.HashMd5=string(tag.Md5Sum[:])
	data.EncryptedKey=make([]byte,16)
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
/*	binary.Write(buf,binary.LittleEndian,&tag.OwnerId)
	binary.Write(buf,binary.LittleEndian,tag.Uuid)
	binary.Write(buf,binary.LittleEndian,tag.Md5Sum)
	binary.Write(buf,binary.LittleEndian,&tag.FromType)
	binary.Write(buf,binary.LittleEndian,tag.FromObj)
	binary.Write(buf,binary.LittleEndian,&tag.Time)
	binary.Write(buf,binary.LittleEndian,tag.EKey)
	binary.Write(buf,binary.LittleEndian,tag.Descr)
	binary.Write(buf,binary.LittleEndian,tag.Padding)
	*/
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

func LoadFromDisk(fname string)(*TagInFile,error){
	f,err:=os.Open(fname)
	if err!=nil{
		return nil,err
	}
	defer f.Close()
	tag:=new(TagInFile)
	if err=binary.Read(f,binary.LittleEndian,tag);err==nil{
		fmt.Printf("uuid: %sTagtype: %d, md5 :%s obj %s, ekey: %x",string(tag.Uuid[:]),tag.FromType,string(tag.Md5Sum[:]),string(tag.FromObj[:]),tag.EKey)
		return tag,nil
	}else{
		fmt.Println("decode error:",err)
		return nil,err
	}
}

func (info *LoginInfo) Logout() error{
    return  nil
}

