package main

/*
#include <stdio.h>
#include <unistd.h>
#include <stdlib.h>
#include <fcntl.h>

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
	"dbop"
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
	err=MountDirInC(ipath,tmpdir,passwd,"rw")
	if err!=nil{
		fmt.Println("mount dir in c error:",err)
		return err
	}
	dmap:=make(map[string]MountOpt)
	dmap[tmpdir]=MountOpt{"/mnt","rw"}
	err=CreatePod("centos",dmap)
	err=exec.Command("umount",tmpdir).Run()
	if err!=nil{
		fmt.Println("umount err",err)
	}
	err=os.Remove(tmpdir)
	if err!=nil{
		fmt.Println("remove error:",err)
	}
	return nil
}

func CreatePod(imgname string,dirmap map[string]MountOpt) error{
	ctcmd:="docker run -it --rm "
	for k,v:=range dirmap{
		ctcmd=ctcmd+" -v "+k+":"+v.dstpt+":"+v.access
	}
	ctcmd=ctcmd+" "+imgname+" /bin/bash"

	ccmd:=C.CString(ctcmd)
	defer C.free(unsafe.Pointer(ccmd))
	C.system(ccmd)
	return nil
}

func PrepareInDir(sinfo *core.ShareInfo)(string,string,error){
	fmt.Println("Preparing container environment...")
    uuid,_:=core.GetUuid()
    tmpdir:=os.TempDir()+"/"+uuid
	err:=os.MkdirAll(tmpdir,0755)
	if err!=nil{
		fmt.Println("mkdir error in PrepareDir:",err)
		return tmpdir,"",err
	}
	if sinfo.IsDir!=0{
	    st,_:=os.Stat(sinfo.FileUri) // we have read fileheader from it before, so Stat should return no error
		size:=st.Size()-60  // the format of fileheader has been validated before, so the result should not be negtive
		zfile,_:=os.Open(sinfo.FileUri)
		csdrd:=NewCSDReader(zfile)
		err:=core.UnzipFromFile(csdrd,size,tmpdir)
		if err!=nil{
			fmt.Println("Unzip from",sinfo.FileUri,"to",tmpdir,"error:",err)
			return tmpdir,"",err
		}
		zfile.Close()
	}else{
		err=exec.Command("dd","if="+sinfo.FileUri,"of="+tmpdir+"/"+sinfo.OrgName,"bs=1","skip=60").Run()
		if err!=nil{
			fmt.Println("exec error:",err)
		}
	}
	orgkey:=make([]byte,16)
	DoDecodeInC(sinfo.EncryptedKey,sinfo.RandKey,orgkey,16)
	uuid,_=core.GetUuid()
	plaindir:=os.TempDir()+"/"+uuid
	err=os.MkdirAll(plaindir,0755)
	if err!=nil{
		fmt.Println("Mkdir ",plaindir,"error:",err)
		return tmpdir,plaindir,err
	}
	err=MountDirInC(tmpdir,plaindir,orgkey,"ro")
	if err!=nil{
		fmt.Println("Mount cmfs in prepare indata error:",err)
		return tmpdir,plaindir,err
	}
	return tmpdir,plaindir,nil
}

func PrepareOutDir(odir string,key []byte)(string,string,error){
	uuidsrc,_:=core.GetUuid()
	uuiddst,_:=core.GetUuid()
	srcdir:=odir+"/"+uuidsrc
	dstdir:=os.TempDir()+"/"+uuiddst
	os.MkdirAll(srcdir,0755)
	os.MkdirAll(dstdir,0755)
	MountDirInC(srcdir,dstdir,key,"rw")
	return srcdir,dstdir,nil
}

func RecordNewDataInfo(outsrc string,linfo *core.LoginInfo)error{
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
		head,err:=core.LoadShareInfoHead(ipath)
		if err!=nil{
			fmt.Println("Load share info head in MountFile error:",err)
			return err
		}
		sinfo,err:=dbop.LoadShareInfo(head)
		if err!=nil{
			fmt.Println("Load share info from head error:",err)
			return err
		}
		sinfo.FileUri=ipath
		inlist:=false
		for _,user:=range sinfo.Receivers{
			if linfo.Name==user{
				inlist=true
				break
			}
		}
		if !inlist{
			fmt.Println(linfo.Name,"is not in shared user list")
			return errors.New("Not shared user")
		}

		insrc,indst,err:=PrepareInDir(sinfo)
		if insrc!=""{
			defer func(){
				if err:=os.RemoveAll(insrc);err!=nil{
					fmt.Println("remove src dir error:",err)
				}
			}()
		}
		if indst!=""{
			defer func(){
				if err:=os.RemoveAll(indst);err!=nil{
					fmt.Println("Remove dst dir error:",err)
				}
			}()
		}
		if err!=nil{
			fmt.Println("Error in prepare indata:",err)
			return err
		}else{
			defer func(){
				exec.Command("umount",indst).Run()
			}()
		}
		mntmap:=make(map[string] MountOpt)
		mntmap[indst]=MountOpt{"/indata","ro"}
		var outsrc,outdst string
		if sinfo.Perm&1 !=0{
			if outpath==""{
				fmt.Println("use parameter -out to set output path")
				return errors.New("missing output dir")
			}
		//	randpass:=[]byte("123456")
		//	randpass=append(randpass,0,0,0,0,0,0,0,0,0,0)
		//	fmt.Println("pass:",randpass)
			randpass,_:=core.RandPasswd()
			outsrc,outdst,err=PrepareOutDir(outpath,randpass)
			// outsrc should exist and be recorded later
			if outdst!=""{
				defer os.Remove(outdst)
			}
			if err!=nil{
				fmt.Println("Prepare outdir error:",err)
				return err
			}else{
				defer func(){
					exec.Command("umount",outdst).Run()
				}()
			}
			mntmap[outdst]=MountOpt{"/outdata","rw"}
			CreatePod("cmrw",mntmap)
			RecordNewDataInfo(outsrc,linfo)
		}else{
			CreatePod("cmro",mntmap)
		}

	}else if ftype==core.RAWDATA{
		/* mount a encrypted local datus
		0. not a dir, so mkdir tmpdir and hard link data after ShareInfoHeader into it
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

