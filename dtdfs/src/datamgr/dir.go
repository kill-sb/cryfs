package main

import(
    "os"
	"unsafe"
    "fmt"
    "path/filepath"
    "strings"
	core "coredata"
)
/*
#include <stdlib.h>
*/
import "C"

func EncodeDir(ipath string, opath string, linfo *core.LoginInfo) (string , error){
    /* 
	1. prepare for EncryptData
    2. mkdir a dst dir in opath ,walk src dir , make same directory structure, and encrypt every file 
    */
    passwd,err:=core.RandPasswd()
    if err!=nil{
        return "",err
    }
    fname,err:=GetFileName(ipath)
    if err!=nil{
        return "",err
    }
    pdata:=new(core.EncryptedData)
    pdata.Uuid,_=core.GetUuid()
	pdata.FromRCId=0
	pdata.FromContext=nil
    pdata.Descr="cmit encrypted dir"
    pdata.OrgName=fname
    pdata.OwnerId=linfo.Id
    pdata.EncryptingKey=passwd
    pdata.Path=opath
    pdata.IsDir=1

    ofile:=opath+"/"+pdata.Uuid
	finfo,_:=os.Stat(ipath)
	os.MkdirAll(ofile,finfo.Mode())
    RecordMetaFromRaw(pdata,linfo.Keylocalkey,passwd,linfo.Token)

	filepath.Walk(ipath, func (pathname string,info os.FileInfo, err error) error{
		if ipath==pathname{
			return nil
		}
		noprefix:=core.StripAllSlash(ipath)
		relative:=strings.TrimPrefix(pathname,noprefix)
		if info.IsDir(){
			fmt.Println(ofile+relative,info.Mode())
			err:=os.MkdirAll(ofile+relative,info.Mode())
			if err!=nil{
				fmt.Println("Mkdir error",err)
			}
		}else{
//			fmt.Println(pathname,"->",ofile+relative)
			DoEncodeFileInC(pathname,ofile+relative,passwd)
		}
		return nil
	})
/*	if ouid!=0{
		ChOwner(ofile,true)
	}*/
    return pdata.Uuid,nil
}

func DecodeDir(ipath,opath string , passwd []byte) error{
	fmt.Println("Decode dir ",ipath,opath)
	filepath.Walk(ipath, func (pathname string,info os.FileInfo, err error) error{
		if pathname==ipath{
			return nil
		}
		noprefix:=core.StripAllSlash(ipath)
		relative:=strings.TrimPrefix(pathname,noprefix)
		if info.IsDir(){
//			fmt.Println("mkdir ",opath+relative)
			err:=os.MkdirAll(opath+relative,info.Mode())
			if err!=nil{
				fmt.Println("mkdir ",opath+relative,len(opath+relative),"error:",err)
			}
		}else{
//			fmt.Println(pathname,"->",opath+"/"+relative)
			DoDecodeFileInC(pathname,opath+relative,passwd,0)
		}
		return nil
	})

	return nil
}

type CSDReader struct{
	orgfile *os.File
}

func NewCSDReader(ifile* os.File)*CSDReader{
	rdr:=new (CSDReader)
	rdr.orgfile=ifile
	return rdr
}

func (rdr *CSDReader)ReadAt(p[]byte, off int64)(int,error){
	return rdr.orgfile.ReadAt(p, off+core.CSDV2HDSize)
}

func DecodeCSDToDir(ifile,opath string, passwd []byte)error{
/*
	1. unzip left part of csd file(from offset 60) to dst dir
	2. walk dstdir and decode every file
*/
	st,_:=os.Stat(ifile) // we have read fileheader from it before, so Stat should return no error
	size:=st.Size()-core.CSDV2HDSize	// the format of fileheader has been validated before, so the result should not be negtive
	zfile,_:=os.Open(ifile)
	csdrd:=NewCSDReader(zfile)
	err:=core.UnzipFromFile(csdrd,size,opath)
	if err!=nil{
		fmt.Println("Unzip from",ifile,"to",opath,"error:",err)
		return err
	}
	zfile.Close()

	filepath.Walk(opath, func (pathname string,info os.FileInfo, err error) error{
		if pathname==opath{
			return nil
		}
		if !info.IsDir(){
			tmpfile:=opath+"/.___"+info.Name()+"___.tmp"
			DoDecodeFileInC(pathname,tmpfile,passwd,0)
			err=os.Rename(tmpfile,pathname)
			if err!=nil{
				fmt.Println("rename error:",err)
				return err
			}
		}
		return nil
	})
	return nil
}

func ChOwner(dst string,addtag bool)error{
	var cmd string
	if addtag{
		cmd=fmt.Sprintf("chown -R %d:%d %s %s >/dev/null 2>/dev/null", ouid,ogid,dst,dst+".tag")
	}else{
		cmd=fmt.Sprintf("chown -R %d:%d %s >/dev/null 2>/dev/null", ouid,ogid,dst)
	}
	ccmd:=C.CString(cmd)
	defer C.free(unsafe.Pointer(ccmd))
	C.system(ccmd)
	return nil
}

