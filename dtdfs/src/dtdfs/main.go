package main
import (
	"unsafe"
	"os"
	"coredata"
	"fmt"
)

/*
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <sys/types.h>
*/
import "C"

func checkpath()error{
	return nil
}

func main(){
	if err:=checkpath();err!=nil{ // check in,out and import access here
		fmt.Println(err)
		return
	}
	C.setuid(0);
	C.setgid(0);
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

