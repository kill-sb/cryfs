package main

/*
#include <stdio.h>
#include <unistd.h>
#include <stdlib.h>
#include <fcntl.h>

int initmntopt()
{
//	unshare
}

int mount_cmfs(const char* src, const char* dst,const char* passwd, const char* opt){

    int fd[2];
    int p=pipe(fd);
    if(fork()==0){
        close(fd[1]);
        dup2(fd[0],0);
        execlp("/usr/local/bin/cmfs","/usr/local/bin/cmfs",src,dst,"-o",opt,NULL);
    }else{
        close(fd[0]);
        write(fd[1],passwd,16);
    }
}

*/
import "C"

import(
	"fmt"
	"os"
	"os/exec"
	"unsafe"
//	"strings"
	"errors"
//	"dbop"
	core "coredata"
)

type MountOpt struct{
	dstpt string
	access string
}

func doMount(){
	if inpath==""{
        fmt.Println("You should set inpath explicitly")
        return
    }

    info,err:=os.Stat(inpath)
	if err!=nil{
        fmt.Println("Can't find ",inpath)
        return
    }
/*
    if outpath=="" {
        fmt.Println("You should set outpath explicitly")
		return
    }
*/
	if loginuser==""{
		fmt.Println("use parameter -user=NAME to set login user")
		return
	}
	linfo,err:=Login(loginuser)
	if err!=nil{
		fmt.Println("login error:",err)
		return
	}
	defer linfo.Logout()

	if info.IsDir(){
		MountDir(inpath,linfo) // must be encrypted data, not shared file
	}else{
		MountFile(inpath,linfo) // may be a encrypted data or a csdfile(may also be a file or a dir)
	}
}

func MountDirInC(src,dst string,passwd []byte,mode string)error{
	cpasswd:=(*C.char)(unsafe.Pointer(&passwd[0]))
	csrc:=C.CString(src)
	cdst:=C.CString(dst)
	copt:=C.CString(mode)
	defer C.free(unsafe.Pointer(csrc))
	defer C.free(unsafe.Pointer(cdst))
	defer C.free(unsafe.Pointer(copt))
	C.mount_cmfs(csrc,cdst,cpasswd,copt)
	return nil
}

func MountDir(ipath string, linfo *core.LoginInfo)error{
	/*
		since the data belongs to user himself, we just mount the dir with r/w perm
		when data is written , it is encrypted with same key that decrypted original data. 
		we will not check the md5 of a encrypted dir
	*/
	tag,err:=core.LoadTagFromDisk(ipath)
	if err!=nil{
		fmt.Println("LoadTagFromDisk error in MountDir:",err)
		return err
	}
/*	
	process:
	0. get enc passwd
	1. mkdir tmpdir , mount encdir -> tmpdir with passwd
	2. create pod with tmpdir->/mnt
	3. waiting for pod exit and umount tmpdir,rm tmpdir
*/
	passwd:=make([]byte,16)
	DoDecodeInC(tag.EKey[:],linfo.Keylocalkey,passwd,16)
	uuid,_:=core.GetUuid()
	tmpdir:=os.TempDir()+"/"+uuid
	err=os.MkdirAll(tmpdir,0755)
	if err==nil{
	fmt.Println(tmpdir,"mkdir ok")
	}
	defer os.Remove(tmpdir)
	err=MountDirInC(ipath,tmpdir,passwd,"rw")
	if err!=nil{
		fmt.Println("mount dir in c error:",err)
		return err
	}else{
		fmt.Println("Mount ok")
	}
	dmap:=make(map[string]MountOpt)
	dmap[tmpdir]=MountOpt{"/mnt","rw"}
	err=CreatePod("centos",dmap)
	if err==nil{
	fmt.Println("mkdir ok")
	}
	exec.Command("umount","/mnt").Run()
	return nil
}


func CreatePod(imgname string,dirmap map[string]MountOpt) error{
	ctcmd:=[]string{"run","-it","--rm"}
	for k,v:=range dirmap{
		ctcmd=append(ctcmd,"-v",k+":"+v.dstpt+":"+v.access)
	}
	ctcmd=append(ctcmd,imgname,"/bin/bash")
	fmt.Println(ctcmd)
	err:=exec.Command("docker","run","-it","centos","/bin/bash").Run()
	//err:=exec.Command("docker",ctcmd).Run()
	if err!=nil{
		fmt.Println("Create container error:",err)
		return err
	}
	return nil
}

func MountFile(ipath string, linfo *core.LoginInfo)error {
	ftype:=GetDataType(ipath)
	if ftype==core.CSDFILE{
		/* common shared file:
		0. Get IsDir()
			false: mkdir a tmpdir and copy encrypted data into it
			true: unzip to a tmpdir
		1. check perm:
			0 unzip to a tmpdir , mount tmpdir to indata, 
			1 mount dir to indata and mount write dir to outdata, write new encrypted tag and update db
		2. start container
		3. delete tmpdir(may keep it for next use later)
		*/
	}else if ftype==core.RAWDATA{
		/* mount a encrypted local datus
		0. not a dir, so mkdir tmpdir and copy data after ShareInfoHeader into it
		1. checkperm
			0 mount tmpdir to indata
			1. mount tmpdir to indata and write new crypted tag and update db
			2. startcontainer
			3. delete tmpdir
		*/
	}else{
		fmt.Println("Unknow data format")
		return errors.New("Unknown data format")
	}

	return nil
}

