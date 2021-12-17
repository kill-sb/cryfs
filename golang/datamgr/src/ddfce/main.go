package main
import (
	"unsafe"
	"os"
)

/*
#include <stdio.h>
#include <stdlib.h>
*/
import "C"

func main(){
	strcmd:="unshare -m ./datamgr "
	nlen:=len(os.Args)
	for i:=1;i<nlen;i++{
		strcmd=strcmd+" "+os.Args[i]
	}
    ccmd:=C.CString(strcmd)
    defer C.free(unsafe.Pointer(ccmd))
    C.system(ccmd)
}

