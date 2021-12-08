package coredata

import (
	"net"
	"bytes"
	"encoding/binary"
	"os"
)

type EncryptedData struct{
    Uuid string
    Descr string
    FromType int
    FromObj string
    OwnerId int32
    EncryptedKey []byte
	Path	string
}
const (
    INVALID=iota
    ENCODE
	DECODE
    DISTRIBUTE
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
	Md5Sum	[32] byte //32
	FromType byte
	FromObj [255] byte //255
	Time	int64
	EKey	[16] byte //16
	Descr	[100] byte // 100
	Padding	[96] byte // 512-4-32-256-24-100=96
}

func (tag *TagInFile) SaveToDisk(fname string)error{
	fd,err:=os.Create(fname)
	if err!=nil{
		return err
	}
	defer fd.Close()
	buf:=new(bytes.Buffer)
	binary.Write(buf,binary.LittleEndian,&tag.OwnerId)
	binary.Write(buf,binary.LittleEndian,tag.Md5Sum)
	binary.Write(buf,binary.LittleEndian,&tag.FromType)
	binary.Write(buf,binary.LittleEndian,tag.FromObj)
	binary.Write(buf,binary.LittleEndian,&tag.Time)
	binary.Write(buf,binary.LittleEndian,tag.EKey)
	binary.Write(buf,binary.LittleEndian,tag.Descr)
	binary.Write(buf,binary.LittleEndian,tag.Padding)
	fd.Write(buf.Bytes())
	return nil
}

func LoadFromDisk(fname string)(*TagInFile,error){
	fd,err:=os.Open(fname)
	if err!=nil{
		return nil,err
	}
	defer fd.Close()
	tag:=new(TagInFile)
	rawbuf:=make([]byte,512)
	fd.Read(rawbuf)
	buf:=bytes.NewBuffer(rawbuf)
    binary.Read(buf,binary.LittleEndian,&tag.OwnerId)
    binary.Read(buf,binary.LittleEndian,tag.Md5Sum)
    binary.Read(buf,binary.LittleEndian,&tag.FromType)
    binary.Read(buf,binary.LittleEndian,tag.FromObj)
    binary.Read(buf,binary.LittleEndian,&tag.Time)
    binary.Read(buf,binary.LittleEndian,tag.EKey)
    binary.Read(buf,binary.LittleEndian,tag.Descr)
    binary.Read(buf,binary.LittleEndian,tag.Padding)
	return tag,nil
}

func (info *LoginInfo) Logout() error{
    return  nil
}

