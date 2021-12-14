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

func EncodeDir(ipath string, opath string, luser *core.LoginInfo) error{
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
    pdata.IsDir=0

    ofile:=opath+"/"+pdata.Uuid
    cpasswd:=(*C.char)(unsafe.Pointer(&passwd[0]))
    cipath:=C.CString(ipath)
    cofile:=C.CString(ofile)
    defer C.free(unsafe.Pointer(cipath))
    defer C.free(unsafe.Pointer(cofile))
    C.do_encodefile(cipath,cofile,cpasswd)
    pdata.HashMd5,_=GetFileMd5(ofile)
    RecordMetaFromRaw(pdata,linfo.Keylocalkey,passwd,ipath,opath)
    return nil

	filepath.Walk(ipath, func (path string,info os.FileInfo, err error) error{
		noprefix:=StripAllSlash(ipath)
		relative:=strings.TrimSuffix(path,noprefix)
		if info.IsDir(){
			os.MkdirAll(opath+"/"+relative)
		}else{

		}
	}
    return nil
}

func DecodeDir(inpath,outpath string , linfo *core.LoginInfo) error{
}

