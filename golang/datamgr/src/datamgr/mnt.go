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

//var multisrc []string

import(
	"fmt"
	"os"
	"io/ioutil"
	"os/exec"
	"time"
	"unsafe"
	"strings"
	"errors"
	core "coredata"
)

type MountOpt struct{
	dstpt string
	access string
}

func LocalTempDir(ipath string)string{
	finfo,err:=os.Stat(ipath)
	if err!=nil{
		fmt.Println(ipath, "not exists")
		return ipath
	}
	if finfo.IsDir(){
		return strings.TrimSuffix(ipath,"/")+"/"
	}
	return strings.TrimSuffix(ipath,finfo.Name())
	// return value must with a '/' in the end

}

func doMount(){
	if inpath==""{
        fmt.Println("You should set inpath explicitly")
        return
    }
	paths:=strings.Split(inpath,",")
	multisrc:=make([]string,0,len(paths))
	for _,str:=range paths{
		str=strings.TrimSpace(str)
		if str!=""{
			multisrc=append(multisrc,str)
		}
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

	/*
    info,err:=os.Stat(inpath)
	if err!=nil{
        fmt.Println("Can't find ",inpath)
        return
    }*/

/*
	if info.IsDir(){
		MountDir(inpath,linfo) // must be encrypted data, not shared file
	}else{

		MountFile(inpath,linfo) // may be a encrypted data or a csdfile(may also be a file or a dir)
	}*/
	MountObjs(multisrc)
}

func MountObjs(inputs []string){
/*

	1. detect path valid
	2. check data access permission, any readonly data will cause not output path(mount return different value),then mount them with different key
	3. put all cmfs-mounted dir and import dir in a slice, then mount in-paths to /indata/dataN, mount import-path to /indata/import
	4. write data to /output as old
	5. record all in-paths and import-path to server(db)
*/

//	for _,obj:=range inputs{
//	}
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
	if tag.OwnerId!=linfo.Id{
		fmt.Println("The data does not belong to",linfo.Name)
		return errors.New("Invalid user")
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
	tmpdir:=LocalTempDir(inpath)+"."+uuid
	//tmpdir:=os.TempDir()+"/"+uuid
	err=os.MkdirAll(tmpdir,0755)
	err=MountDirInC(ipath,tmpdir,passwd,"rw")
	if err!=nil{
		fmt.Println("mount dir in c error:",err)
		return err
	}
	dmap:=make(map[string]MountOpt)
	dmap[tmpdir]=MountOpt{"/mnt","rw"}
	err=CreatePod("cmit",dmap)
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
    tmpdir:=LocalTempDir(inpath)+"."+uuid
    //tmpdir:=os.TempDir()+"/"+uuid
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
	plaindir:=LocalTempDir(inpath)+"."+uuid
	//plaindir:=os.TempDir()+"/"+uuid
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
	dstdir:=LocalTempDir(inpath)+"."+uuiddst
	//dstdir:=os.TempDir()+"/"+uuiddst
	os.MkdirAll(srcdir,0755)
	os.MkdirAll(dstdir,0755)
	MountDirInC(srcdir,dstdir,key,"rw")
	return uuidsrc,dstdir,nil
}

func SingleFileInDir(dirname string)(string, bool){
	cont,err:=ioutil.ReadDir(dirname)
	if err==nil && len(cont)==1{
		return strings.TrimSuffix(dirname,"/")+"/"+cont[0].Name(),true
	}
	return "",false
}

func RecordNewDataInfo(opath,datauuid string ,passwd []byte,linfo *core.LoginInfo,sinfo *core.ShareInfo)error{
    pdata:=new(core.EncryptedData)

    pdata.Uuid=datauuid
    pdata.Descr="cmit encrypted dir"
	// FIXME: should be replaced later because of multi-source processing
    //pdata.FromType=core.CSDFILE
    pdata.FromPlain=0
    pdata.FromObj=sinfo.Uuid
	finfo,_:=os.Stat(sinfo.FileUri)
	pdata.OrgName=finfo.Name()+".outdata" //FIXME: modified to parameter: newname
    pdata.OwnerId=linfo.Id
    pdata.EncryptingKey=passwd
    pdata.Path=opath
    pdata.IsDir=1

    pdata.HashMd5=""


	if dataname,single:=SingleFileInDir(opath+"/"+datauuid);single{
		newuuid,_:=core.GetUuid()
		finfo,_=os.Stat(dataname)
		pdata.Uuid=newuuid
		pdata.OrgName=finfo.Name()
		pdata.Descr="cmit encrypted data"
		os.Link(dataname,opath+"/"+newuuid)
		os.Remove(dataname)
		if err:=os.Remove(opath+"/"+datauuid);err!=nil{
			fmt.Println("Error in RecordNewDataInfo(dir not empty):",err)
			return err
		}
		pdata.IsDir=0
	}
	err:=RecordMetaFromRaw(pdata,linfo.Keylocalkey,passwd,sinfo.FileUri,opath,linfo.Token)
	return err
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
		sinfo,err:=GetShareInfoFromHead(head,linfo)
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

		// check expire first
		strexp:=strings.Replace(sinfo.Expire," ","T",1)+"+08:00"
		tmexp,err:=time.Parse(time.RFC3339,strexp)
		if err!=nil{
			fmt.Println("Parse expire time error:",err)
			return err
		}
		tmnow:=time.Now()
		if tmnow.After(tmexp){
			fmt.Println("The shared data has expired at :",sinfo.Expire)
			return errors.New("Data expired")
		}

		// check left time
/*		if sinfo.LeftUse==0{
			fmt.Printf("The max open times(%d) has been exhausted\n",sinfo.MaxUse)
			return errors.New("open times exhausted")
		}*/
		if sinfo.Perm&1 !=0 &&outpath==""{
				fmt.Println("use parameter -out to set output path")
				return errors.New("missing output dir")
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
		var outuuid,outdst string
		if sinfo.Perm&1 !=0{
			randpass,_:=core.RandPasswd()
			outuuid,outdst,err=PrepareOutDir(outpath,randpass)
			if outdst!=""{
				defer os.Remove(outdst)
			}
			if err!=nil{
				fmt.Println("Prepare outdir error:",err)
				return err
			}/*else{
				defer func(){
					exec.Command("umount",outdst).Run()
				}()
			}*/
			mntmap[outdst]=MountOpt{"/outdata","rw"}
			CreatePod("cmrw",mntmap)
			exec.Command("umount",outdst).Run()
			if err=os.Remove(outpath+"/"+outuuid);err!=nil{ // if the directory is not empty ,it will fail, otherwize, we treat it as empty
				err=RecordNewDataInfo(outpath,outuuid,randpass,linfo,sinfo)
				if err!=nil{
					fmt.Println("Record data metainfo error:",err)
					return err
				}
			}else{
				fmt.Println("outdata directory is empty, output data register ignored")
			}
		}else{
			if outpath!=""{
				fmt.Println("Warning: the data is not permitted to be reshared or output any process result, '-out",outpath+"'", "parameter ignored")
			}
			CreatePod("cmro",mntmap)
		}

	}else if ftype==core.ENCDATA{
		if err:=MountSingleEncFile(ipath,linfo);err!=nil{
			fmt.Println("Mount single encoded file error:",err)
			return err
		}
	}else{
		fmt.Println("Unknow data format")
		return errors.New("Unknown data format")
	}
	return nil
}

func PrepareSingleFileDir(ipath string,fname string)(string,string,error){
	srcuuid,_:=core.GetUuid()
	dstuuid,_:=core.GetUuid()
	tmpdir:=LocalTempDir(ipath)
	srcdir:=tmpdir+"."+srcuuid
	//srcdir:=os.TempDir()+"/"+srcuuid
	dstdir:=tmpdir+"."+dstuuid
//	dstdir:=os.TempDir()+"/"+dstuuid
	os.MkdirAll(srcdir,0755)
	os.MkdirAll(dstdir,0755)
	return srcdir,dstdir,os.Link(ipath,srcdir+"/"+fname)
}

func MountSingleEncFile(ipath string,linfo *core.LoginInfo)error{
	tag,err:=core.LoadTagFromDisk(ipath)
	if err!=nil{
		fmt.Println("LoadTagFromDisk error in MountDir:",err)
		return err
	}
	dinfo,err:=GetDataInfo(tag)
	if err!=nil{
		fmt.Println("GetDataInfo in MountSingleEncFile error:",err)
		return err
	}

	if dinfo.OwnerId!=linfo.Id{
		fmt.Println("The data does not belong to",linfo.Name)
		return errors.New("Invalid user")
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
		/* mount a encrypted local datus
		0. not a dir, so mkdir tmpdir and hard link data after ShareInfoHeader into it
		1. checkperm
			0 mount tmpdir to indata
			1. mount tmpdir to indata and write new crypted tag and update db
			2. startcontainer
			3. delete tmpdir
		*/
	srcdir,dstdir,err:=PrepareSingleFileDir(ipath,dinfo.OrgName)
	if srcdir!=""{
		defer os.RemoveAll(srcdir)
	}
	if dstdir!=""{
		defer os.Remove(dstdir)
	}
	if err!=nil{
		fmt.Println("Prepare dir error:",err)
		return err
	}
	err=MountDirInC(srcdir,dstdir,passwd,"rw")
	if err!=nil{
		fmt.Println("mount dir in c error:",err)
		return err
	}
	dmap:=make(map[string]MountOpt)
	dmap[dstdir]=MountOpt{"/mnt","rw"}
	err=CreatePod("cmit",dmap)
	exec.Command("umount",dstdir).Run()
	if err!=nil{
		fmt.Println("Create pod error:",err)
		return err
	}
	err=UpdateMetaInfo(ipath,tag,dinfo,linfo)
	if err!=nil{
		fmt.Println("UpdateMetaInfo error:",err)
//		return err
	}
	return err
}


// TODO  the function should be removed, since modify an encrypted file will create a new encrypted file, instead of update the original one
func UpdateMetaInfo(ipath string,tag *core.TagInFile,dinfo *core.EncryptedData, linfo *core.LoginInfo)error{
	dinfo.HashMd5,_=GetFileMd5(ipath)
    for i,j:=range []byte(dinfo.HashMd5){
        tag.Md5Sum[i]=j
    }
    tag.Time=time.Now().Unix()
	tag.SaveTagToDisk(strings.TrimSuffix(ipath,".tag")+".tag")
	return UpdateDataInfo_API(dinfo,linfo)
	//return dbop.UpdateMeta(dinfo)
}

