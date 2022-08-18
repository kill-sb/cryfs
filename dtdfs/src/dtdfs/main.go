package main
import (
	"unsafe"
	"os"
	"coredata"
	"syscall"
	"path/filepath"
	"errors"
	"strings"
	"fmt"
)

/*
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <sys/types.h>
*/
import "C"

func GetParentDir(path string)string{
	if path==""{
		return ""
	}
	for{
		if !strings.HasSuffix(path,"/"){
			break
		}
		path=strings.TrimSuffix(path,"/")
	}
	entries:=strings.Split(path,"/")
	return strings.TrimSuffix(path,entries[len(entries)-1])
}

func TestRead(path string)error{
	if err:=syscall.Access(path,4);err!=nil{
		return errors.New(fmt.Sprintf("%s can't be read, check the pathname or permission.",path))
	}
	return nil
}

func TestWrite(path string)error{
	if err:=syscall.Access(path,2);err!=nil{
		return errors.New(fmt.Sprintf("%s is not writable, check the pathname or permission.",path))
	}
	return nil
}

func CheckIn(path string) error{
    paths:=strings.Split(path,",")
    for _,str:=range paths{
        apath,err:=filepath.Abs(str)
        if err!=nil{
			return err
        }else{
			err=TestRead(apath)
			if err!=nil{
				return err
			}
        }
    }
	return nil
}

func CheckOut(path string) error{
	/*
	permission check:
		if exist  
			-> if isdir  ->check self 
			-> else if isfile -> check parentdir
		else // not exist
			check parentdir
	*/
	apath,err:=filepath.Abs(path)
	if err!=nil{
		return err
	}
	fi,err:=os.Stat(apath)
	if err!=nil{ // not exist or no permission
		pdir:=GetParentDir(apath)
		if pdir==""{
			return errors.New("Invalid -out parameter:"+path)
		}
		return TestWrite(pdir)
	}
	if fi.IsDir(){
		return TestWrite(apath)
	}
	pdir:=GetParentDir(apath)
	if pdir==""{
		return errors.New("Invalid -out parameter:"+path)
	}
	return TestWrite(pdir)
}

func CheckTool(path string) error{
	apath,err:=filepath.Abs(path)
	if err!=nil{
		return err
	}
/*	err=TestRead(apath)
	if err!=nil{
		return err
	}*/
	// don't check dir here, mnt will do it later
    err=filepath.Walk(apath, func (curpath string, info os.FileInfo, err error)error{
        return TestRead(curpath)
    })
	return err
}

func checkpath()error{
	var in,out,tool string
	StringVar(&in,"in","","")
	StringVar(&out,"out","","")
	StringVar(&tool,"import","","")
	Parse()
//	fmt.Println(in,",",out,",",tool)
	var err error
	if in!=""{
		err=CheckIn(in)
	}
	if err!=nil{
		return err
	}
	if out!=""{
		err=CheckOut(out)
	}
	if err!=nil{
		return err
	}
	if tool!=""{
		err=CheckTool(tool)
	}
	if err!=nil{
		return err
	}
	return nil
}

func main(){
	if syscall.Getuid()!=0{ // non root
		if err:=checkpath();err!=nil{ // check in,out and import access here
			fmt.Println(err)
			return
		}
		C.setuid(0);
		C.setgid(0);
	}
	strcmd:=fmt.Sprintf("unshare -m %s/datamgr ",coredata.GetSelfPath())
	nlen:=len(os.Args)
	for i:=1;i<nlen;i++{
		strcmd=strcmd+" \""+os.Args[i]+"\""
	}
	//strcmd+=" 2>/dev/null "
    ccmd:=C.CString(strcmd)
    defer C.free(unsafe.Pointer(ccmd))
    C.system(ccmd)
}

