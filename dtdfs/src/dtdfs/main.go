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
*/
import "C"

func main(){
	strcmd:=fmt.Sprintf("unshare -m %s/datamgr ",coredata.GetSelfPath())
//	strcmd:=fmt.Sprintf("%s/datamgr",coredata.GetSelfPath())
	nlen:=len(os.Args)
	for i:=1;i<nlen;i++{
		strcmd=strcmd+" \""+os.Args[i]+"\""
	}
	//strcmd+=" 2>/dev/null "
    ccmd:=C.CString(strcmd)
    defer C.free(unsafe.Pointer(ccmd))
    C.system(ccmd)
}

