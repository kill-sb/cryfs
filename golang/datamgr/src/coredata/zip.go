package coredata

import(
    "archive/zip"
    "os"
    "io"
    "fmt"
    "path/filepath"
    "strings"
)

func StripAllSlash(path string)string{
		noprefix:=strings.TrimSuffix(path,"/")
		for noprefix!=path{
			path=noprefix
			noprefix=strings.TrimSuffix(path,"/")
		}
		return noprefix
}

func ZipToFile(ipath string, target* os.File )error{
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
		relative:=strings.TrimPrefix(strings.TrimPrefix(path,noprefix),"/")
		header.Name=relative
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
	return nil
}

func UnzipFromFile(src io.ReaderAt,size int64,opath string)error{
	err:=os.MkdirAll(opath,0755)
	if err!=nil{
		fmt.Println("Create path error:",err)
		return err
	}
	zrd,err:=zip.NewReader(src,size)
	if err!=nil{
		fmt.Println("zip.NewReader error:",err)
		return err
	}
	for _,dst:=range zrd.File{
		info:=dst.FileInfo()
		dstname:=opath+"/"+dst.Name
		if info.IsDir(){
			os.MkdirAll(dstname,info.Mode())
		}else{
			lfile,err:=os.OpenFile(dstname,os.O_CREATE|os.O_RDWR,info.Mode())
			if err!=nil{
				fmt.Println("Create file ",dst.Name,"error:",err)
				return err
			}
			defer lfile.Close()
			rd,err:=dst.Open()
			if err!=nil{
				fmt.Println("Open file in zip:",dst.Name,"error:",err)
				return err
			}
			defer rd.Close()
			io.Copy(lfile,rd)
			if err!=nil{
				fmt.Println("io.Copy error:",err)
				return err
			}
		}
	}
	return nil
}


