package main

import(
    "archive/zip"
    "os"
    "io"
    "fmt"
    "path/filepath"
    "strings"
	core "coredata"
)
func StripAllSlash(path string)string{
		noprefix:=strings.TrimSuffix(path,"/")
		for noprefix!=path{
			path=noprefix
			noprefix=strings.TrimSuffix(path,"/")
		}
		return noprefix
}

func Zip(ipath string, target* os.File ){
	acv:=zip.NewWriter(target)
	defer acv.Close()

	filepath.Walk(ipath, func (path string, info os.FileInfo, err error)error{
		if path==ipath {
			if info.IsDir(){
				return nil // Walk may enumerate the root-path itself
			}else{
				header,_:=zip.FileInfoHeader(info)
				header.Name=info.Name()
				header.Method=zip.Deflate
				writer,_:=acv.CreateHeader(header)
				file,_:=os.Open(path)
				defer file.Close()
				io.Copy(writer,file)
				return nil
			}
		}
		header,_:=zip.FileInfoHeader(info)
		noprefix:=StripAllSlash(ipath)
		header.Name=strings.TrimSuffix(path,noprefix)
		if info.IsDir(){
			header.Name+="/"
		}else{
			header.Method=zip.Deflate
		}

		writer,_:=acv.CreateHeader(header)

		if !info.IsDir(){
			file,_:=os.Open(path)
			defer file.Close()
			io.Copy(writer,file)
		}
		return nil
	})
}

func Unzip(src *os.File,size int64,opath string){
	err:=os.MkdirAll(opath,0755)
	if err!=nil{
		fmt.Println("Create path error:",err)
		return
	}
	zrd,err:=zip.NewReader(src,size)
	if err!=nil{
		fmt.Println("zip.NewReader error:",err)
		return
	}
	for _,dst:=range zrd.File{
		info:=dst.FileInfo()
		dstname:=opath+"/"+dst.Name
		if info.IsDir(){
			os.MkdirAll(dstname,info.Mode())
		}else{
			lfile,err:=os.OpenFile(dstname,os.O_CREATE,info.Mode())
			//lfile,err:=os.Create(dstname)
			if err!=nil{
				fmt.Println("Create file ",dst.Name,"error:",err)
				return
			}
			defer lfile.Close()
			rd,err:=dst.Open()
			if err!=nil{
				fmt.Println("Open file in zip:",dst.Name,"error:",err)
				return
			}
			defer rd.Close()
			io.Copy(lfile,rd)
		}
	}
}

func EncodeDir(ipath string, opath string, linfo *core.LoginInfo) error{
    /* 
	1. prepare for EncryptData
    2. mkdir a dst dir in opath ,walk src dir , make same directory structure, and encrypt every file 
    */
    passwd,err:=core.RandPasswd()
    if err!=nil{
        return err
    }
    fname,err:=GetFileName(ipath)
    if err!=nil{
        return err
    }
    pdata:=new(core.EncryptedData)
    pdata.Uuid,_=core.GetUuid()
    pdata.Descr=""
    pdata.FromType=core.RAWDATA
    pdata.FromObj=fname
    pdata.OwnerId=linfo.Id
    pdata.EncryptingKey=passwd
    pdata.Path=opath
    pdata.IsDir=1

    ofile:=opath+"/"+pdata.Uuid
	finfo,_:=os.Stat(ipath)
	os.MkdirAll(ofile,finfo.Mode())
//    pdata.HashMd5,_=GetFileMd5(ofile)
    pdata.HashMd5=""
    RecordMetaFromRaw(pdata,linfo.Keylocalkey,passwd,ipath,opath)

	filepath.Walk(ipath, func (pathname string,info os.FileInfo, err error) error{
		if ipath==pathname{
			return nil
		}
		noprefix:=StripAllSlash(ipath)
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
    return nil
}

func DecodeDir(ipath,opath string , passwd []byte) error{
	fmt.Println("Decode dir ",ipath,opath)
	filepath.Walk(ipath, func (pathname string,info os.FileInfo, err error) error{
		if pathname==ipath{
			return nil
		}
		noprefix:=StripAllSlash(ipath)
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

