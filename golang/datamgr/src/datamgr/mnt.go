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
        execlp("/usr/local/bin/cmfs","/usr/local/bin/cmfs",src,dst,"-o",opt,NULL); // TODO verify checksum later
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
	api  "apiv1"
)

type MountOpt struct{
	dstpt string
	access string
}

func LocalTempDir(ipath string)string{
	// return a dir that must be in the same block device with given path, invoker will add a unique name to the dirname later
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
	tools:=make([]string,0,10)
	if mntimport!=""{
		paths=strings.Spit(mntimport,",")
		for _,str:=range paths{
			if str=strings.TrimSpace(str);str!=""{
				tools=append(tools,str)
			}
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
	defer Logout(linfo)

	MountObjs(linfo,multisrc,tools)
}

func PrepareOutputDir(opath string)(string,string,[]byte,error){
	// check output path valid, create uuid-named dir, mount in cmitfs
	if opath==""{
		return "","",nil,nil
	}
	st,err:=os.Stat(opath)
	if err!=nil || !st.IsDir(){
		return "","",nil,errors.New("Invalid outpath")
	}
	uuidsrc,_:=core.GetUuid()
	uuiddst,_:=core.GetUuid()
	srcdir:=opath+"/"+uuidsrc
	dstdir:=opath+"/."+uuiddst
	key,_:=core.RandPasswd()
	os.MkdirAll(srcdir,0755)
	os.MkdirAll(dstdir,0755)
	MountDirInC(srcdir,dstdir,key,"rw")
	return uuidsrc,dstdir,key,nil
}

func ValidateImports(toolpath string)([]*api.ImportFile,error){
    if toolpath==""{
        return nil,nil
    }
	toolpath=strings.TrimSuffix(toolpath,"/")+"/"
    st,err:=os.Stat(toolpath)
    if err!=nil || !st.IsDir(){
        return nil,errors.New("Invalid tool path")
    }
	imlist:=make([]*api.ImportFile,0,20)
    filepath.Walk(toolpath, func (path string, info os.FileInfo, err error)error{
		if !info.IsDir(){
			relpath=strings.TrimPrefix(path,toolpath)
			st,err=os.Stat(path)
			if err!=nil{
				continue
			}
			imnode=new(api.ImportFile)
			if imnode.Sha256,err=GetFileSha256(path);err!=nil{
				continue
			}
			imnode.RelName=relpath
			if imnode.FileDesc,err=exec.Command("file",path).Output();err!=nil{
				imnode.FileDesc=""
			}
			imnode.Size=st.Size()
			imlist=append(imlist,imnode)
        }
	}
	return imlist,nil
}

func ValidateInputs(linfo *core.LoginInfo,inputs []string)(bool, error){
	// check readonly, limit times
	rdonly:=false
	for _,idata:=range inputs{
		st,err:=os.Stat(idata)
		if err!=nil{
			return false,errors.New("Invalid inputdata:",idata)
		}
		dtype:=GetDataType(idata)
		switch dtype{
		case core.CSDFILE:
			head,err:=core.LoadShareInfoHead(idata)
			if err!=nil{
				return rdonly,err
			}
			sinfo,err:=GetShareInfoFromHead(head,linfo,0)
			if sinfo.Perm & 1==0 || sinfo.Perm!=-1{
				rdonly=true
			}
			// check readonly
			if sinfo.LeftUse==0{
				return rdonly,errors.New(idata+"Invalid user or open times exhaused")
			}
			strexp:=strings.Replace(sinfo.Expire," ","T",1)+"+08:00"
			tmexp,err:=time.Parse(time.RFC3339,strexp)
			if err!=nil{
				fmt.Println("Parse expire time error:",err)
				return rdonly,err
			}
			tmnow:=time.Now()
			if tmnow.After(tmexp){
				fmt.Println("The shared data has expired at :",sinfo.Expire)
				return rdonly,errors.New("Data expired")
			}

		case core.ENCDATA:
			dinfo,err:=GetEncDataInfo(st.Name())
			if dinfo.OwnerId!=linfo.Id{
				return rdonly, errors.New(idata+" does not belong to current user")
			}
		default:
			return rdonly, errors.New("Invalid inputdata:"+idata)
		}
	}
	return rdonly,nil
}

func MountCSDFile(filepath string)(string,string,error){
	head,err:=core.LoadShareInfoHead(filepath)
	if err!=nil{
		fmt.Println("Load share info head in MountFile error:",err)
		return err
	}
	sinfo,err:=GetShareInfoFromHead(head,linfo,1)
	if err!=nil{
		fmt.Println("Load share info from head error:",err)
		return err
	}
	sinfo.FileUri=filepath
	return PrepareCSDFileDir(filepath,sinfo)
}

func MountEncDir(dirname string)(string,string,error){
	tag,err:=core.LoadTagFromDisk(dirname)
	if err!=nil{
		fmt.Println("LoadTagFromDisk error in MountDir:",err)
		return "","",err
	}
/*	
	process:
	1. get enc passwd
	2. mkdir tmpdir 
	3. mount encdir -> tmpdir with passwd
*/
	passwd:=make([]byte,16)
	DoDecodeInC(tag.EKey[:],linfo.Keylocalkey,passwd,16)
	dstuuid,_:=core.GetUuid()
	tmpdir:=LocalTempDir(dirname)+"."+dstuuid
	err=os.MkdirAll(tmpdir,0755)
	err=MountDirInC(dirname,tmpdir,passwd,"ro")
	if err!=nil{
		fmt.Println("mount dir in c error:",err)
		return "","",err
	}
	return "",tmpdir,nil // trick here: srcdir should be "" , avoid  being deleted later
}

func MountEncFile(filepath string)(string,string,error){
	tag,err:=core.LoadTagFromDisk(filepath)
	if err!=nil{
		fmt.Println("LoadTagFromDisk error in MountEncFile:",err)
		return err
	}
	dinfo,err:=GetDataInfo(tag)
	if err!=nil{
		fmt.Println("GetDataInfo in MountEncFile error:",err)
		return err
	}
/*	
	process:
	get enc passwd
	mkdir tmpdir ->srcdir,dstdir
	mount srcdir -> dstdir with passwd
*/
	passwd:=make([]byte,16)
	DoDecodeInC(tag.EKey[:],linfo.Keylocalkey,passwd,16)

	srcdir,dstdir,err:=PrepareEncFileDir(filepath,dinfo.OrgName)

	if err!=nil{
		fmt.Println("Prepare dir error:",err)
		return err
	}
	err=MountDirInC(srcdir,dstdir,passwd,"ro")
	if err!=nil{
		fmt.Println("mount dir in c error:",err)
		return srcdir,dstdir,err
	}
	return srcdir,dstdir,nil
}

func MountEncData(linfo *core.LoginInfo,inputs []string)([]string,[]string,error){
	var err error
	srcs:=make([]string,0,20)
	dsts:=make([]string,0,20)
	for _,idata:=range inputs{
		dtype:=GetDataType(idata)
		var src,dst string
		switch dtype{
		case core.CSDFILE:
			src,dst,err=MountCSDFile(idata)
			if err!=nil{
				return nil,nil,err
			}
		case core.ENCDATA:
			st,_:=os.Stat(idata)
			if st.IsDir(){
				src,dst,err=MountEncDir(idata)
			}else{
				src,dst,err=MountEncFile(idata)
			}
			if err!=nil{
				return nil,nil,err
			}
		default:
			return rdonly, errors.New("Invalid inputdata:"+idata)
		}
		srcs=append(srcs,src)
		dsts=append(dsts,dst)
	}
	return srcs,dsts,nil
}

func MountObjs(linfo *core.LoginInfo, inputs []string, tool string){
/*
	1. check local path valid
	2. check data access permission, any readonly data will cause not output path(mount return different value),then mount them with different key
	3. prepare local input/output dir(mount)
	4. put all cmfs-mounted dir and import dir in a slice, then mount in-paths to /indata/dataN, mount import-path to /indata/import to pod
	5. run finish,output data uuid has already been determined in step 3,  if there is output data in the end, invoke createrc here(since endtime and data uuid has been gotten, no need to invoke updaterc later)
	6. record data uuid to server(rcid was available in step 5)
*/

	fmt.Println("Checking import tools...")
	tlinfo,err:=ValidateImports(tool)// TODO need a COW filesystem
	if err!=nil{
		fmt.Println("Import files error:",err)
		return
	}

	// check if there is any readonly(and time-limit?) csd files
	rdonly,err:=ValidateInputs(linfo,inputs)
	if err!=nil{
		fmt.Println("Input source error:",err)
		return
	}
	if rdonly && outpath!=""{
		fmt.Println("There is readonly or open-time-limited data , exporting is forbidden")
		return
	}

	// establish output dir, and mount to a temp dir with random passwd
	var outsrc,outdst string
	var outkey []byte
	if outpath!=""{
		outsrc,outdst,outkey,err=PrepareOutputDir(outpath)
		if err!=nil{
			fmt.Println("Output dir validation error:",err)
			return
		}
		if outdst!=""{
			defer os.Remove(outdst)// outsrc may be remained or removed according to if output data is a single file
		}
	}

	fmt.Println("Processing input data...")
	insrcs,indsts,err:=MountEncData(linfo,inputs)

    defer func(){
		for i,indst:=range indsts{
			if indst!=""{
				exec.Command("umount",indst).Run()
				os.RemoveAll(indst)
			}
			if insrcs[i]!=""{
				os.RemoveAll(insrc[i])
			}
		}
    }()

	if err!=nil{
		fmt.Println("Mount data error:",err)
		return
	}

	// fill input,output and tool dir mount info to MountOpt map
	opts:=PrepareMntOpts(indsts,tool,outdst)

	// need not send to server now
	rc:=CreateOrgRC(linfo,indsts,tlinfo,podimg)// TODO 
	fmt.Println("Creating container...")
	err=CreatePod(podimg,opts)
	if err!=nil{
		fmt.Println("Create container error:",err)
		return
	}
	err=CleanInDirs(indsts)
	if err!=nil{
		fmt.Println("Clear temp input dirs error:",err)
	}

	if outpath!="" && outsrc!=""{
		outuuid,err:=ProcOutputData(outsrc) // will umount and check single file or multi file(dir), return new data uuid (outsrc may be renamed or removed, according to if output data is a dir or empty)
		if err==nil && outuuid!=""{
			fmt.Println("Processing output data...")
			err=RegisterRC(linfo,rc,outuuid)
			if err==nil{
				err=RecordDataInfo(linfo,outuuid,outkey,oname,rc) // invoke createrc and newdata
				if err!=nil{
					fmt.Println("UpdateMetaInfo error:",err)
				}
			}else{
					fmt.Println("Register rc error:",err)
			}
		}
	}

	// tools dir need not be cleaned now, but when it is implemented by a COW filesystem, clean work need to be done here
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
	// TODO : OwnerId need be gotten from server
	/*
	if tag.OwnerId!=linfo.Id{
		fmt.Println("The data does not belong to",linfo.Name)
		return errors.New("Invalid user")
	}*/
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
	ctcmd:="docker run -it --rm --privileged=true "
	for k,v:=range dirmap{
		ctcmd=ctcmd+" -v "+k+":"+v.dstpt+":"+v.access
	}
	ctcmd=ctcmd+" "+imgname+" /bin/bash"

	ccmd:=C.CString(ctcmd)
	defer C.free(unsafe.Pointer(ccmd))
	C.system(ccmd)
	return nil
}

func PrepareCSDFileDir(filepath string, sinfo *core.ShareInfo)(string,string,error){
//	fmt.Println("Preparing container environment...")
    uuid,_:=core.GetUuid()
    tmpdir:=LocalTempDir(filepath)+"."+uuid
	err:=os.MkdirAll(tmpdir,0755)
	if err!=nil{
		fmt.Println("mkdir error in PrepareCSDFileDir:",err)
		return tmpdir,"",err
	}
	if sinfo.IsDir!=0{
	    st,_:=os.Stat(filepath) // we have read fileheader from it before, so Stat should return no error
		size:=st.Size()-core.CSDV2HDSize  // the format of fileheader has been validated before, so the result should not be negtive
		zfile,_:=os.Open(filepath)
		csdrd:=NewCSDReader(zfile)
		err:=core.UnzipFromFile(csdrd,size,tmpdir)
		if err!=nil{
			fmt.Println("Unzip from",filepath,"to",tmpdir,"error:",err)
			return tmpdir,"",err
		}
		zfile.Close()
	}else{
		err=exec.Command("dd","if="+filepath,"of="+tmpdir+"/"+sinfo.OrgName,"bs=4M",fmt.Sprintf("skip=%d",core.CSDV2HDSize),"iflag=skip_bytes").Run()
		if err!=nil{
			fmt.Println("exec error:",err)
		}
	}
	orgkey:=make([]byte,16)
	DoDecodeInC(sinfo.EncryptedKey,sinfo.RandKey,orgkey,16)
	uuid,_=core.GetUuid()
	plaindir:=LocalTempDir(filepath)+"."+uuid
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

/*
func PrepareOutDir(odir string,key []byte)(string,string,error){
	uuidsrc,_:=core.GetUuid()
	uuiddst,_:=core.GetUuid()
	srcdir:=odir+"/"+uuidsrc
	dstdir:=LocalTempDir(outpath)+"."+uuiddst
	//dstdir:=os.TempDir()+"/"+uuiddst
	os.MkdirAll(srcdir,0755)
	os.MkdirAll(dstdir,0755)
	MountDirInC(srcdir,dstdir,key,"rw")
	return uuidsrc,dstdir,nil
}*/

func SingleFileInDir(dirname string)(string, bool){
	cont,err:=ioutil.ReadDir(dirname)
	if err==nil && len(cont)==1{
		sname=cont[0].Name()
		if oname==""{
			oname=sname
		}
		if !cont[0].IsDir(){
			return strings.TrimSuffix(dirname,"/")+"/"+sname,true
		}
	}
	return "",false
}

func RecordDataInfo(linfo *core.LoginInfo,outuuid string ,outkey []byte, oname string ,rc *api.RCInfo)error{
	if oname==""{
		oname=outuuid
	}
	return nil
}

func RecordNewDataInfo(opath,datauuid string ,passwd []byte,linfo *core.LoginInfo,sinfo *core.ShareInfo)error{
    pdata:=new(core.EncryptedData)

    pdata.Uuid=datauuid
    pdata.Descr="cmit encrypted dir"
	// FIXME: should be replaced later because of multi-source processing
    //pdata.FromType=core.CSDFILE
//    pdata.FromPlain=0
//    pdata.FromObj=sinfo.Uuid
	finfo,_:=os.Stat(sinfo.FileUri)
	pdata.OrgName=finfo.Name()+".outdata" //FIXME: modified to parameter: newname
    pdata.OwnerId=linfo.Id
    pdata.EncryptingKey=passwd
    pdata.Path=opath
    pdata.IsDir=1


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
		sinfo,err:=GetShareInfoFromHead(head,linfo,1)
		if err!=nil{
			fmt.Println("Load share info from head error:",err)
			return err
		}
		sinfo.FileUri=ipath
/*		inlist:=false
		for _,user:=range sinfo.Receivers{
			if linfo.Name==user{
				inlist=true
				break
			}
		}
		if !inlist{
			fmt.Println(linfo.Name,"is not in shared user list")
			return errors.New("Not shared user")
		}*/

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

func PrepareEncFileDir(ipath string,fname string)(string,string,error){
	srcuuid,_:=core.GetUuid()
	dstuuid,_:=core.GetUuid()
	tmpdir:=LocalTempDir(ipath)
	srcdir:=tmpdir+"."+srcuuid
	dstdir:=tmpdir+"."+dstuuid
	os.MkdirAll(srcdir,0755)
	os.MkdirAll(dstdir,0755)
	return srcdir,dstdir,os.Link(ipath,srcdir+"/"+fname)
}
/*
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
*/
/*	
	process:
	0. get enc passwd
	1. mkdir tmpdir , mount encdir -> tmpdir with passwd
	2. create pod with tmpdir->/mnt
	3. waiting for pod exit and umount tmpdir,rm tmpdir
*/
//	passwd:=make([]byte,16)
//	DoDecodeInC(tag.EKey[:],linfo.Keylocalkey,passwd,16)
		/* mount a encrypted local datus
		0. not a dir, so mkdir tmpdir and hard link data after ShareInfoHeader into it
		1. checkperm
			0 mount tmpdir to indata
			1. mount tmpdir to indata and write new crypted tag and update db
			2. startcontainer
			3. delete tmpdir
		*/
/*	srcdir,dstdir,err:=PrepareSingleFileDir(ipath,dinfo.OrgName)
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
	return err
}*/


// TODO  the function should be removed, since modify an encrypted file will create a new encrypted file, instead of update the original one
/*func UpdateMetaInfo(ipath string,tag *core.TagInFile,dinfo *core.EncryptedData, linfo *core.LoginInfo)error{
	dinfo.HashMd5,_=GetFileMd5(ipath)
    for i,j:=range []byte(dinfo.HashMd5){
        tag.Md5Sum[i]=j
    }
    tag.Time=time.Now().Unix()
	tag.SaveTagToDisk(strings.TrimSuffix(ipath,".tag")+".tag")
	return UpdateDataInfo_API(dinfo,linfo)
	//return dbop.UpdateMeta(dinfo)
}*/

func CreateRunContext(token string,baseimg string,srcobj []*api.SourceObj, tools []*api.ImportFile)(*api.RCInfo,error){
//	rc:=new (core.RunContext)
	rc:=new (api.RCInfo)
	rc.BaseImg=baseimg
	rc.OS="Linux"
	rc.InputData=srcobj
	rc.ImportPlain=tools
	rc.StartTime=core.GetCurTime()
	err:=CreateRunContext_API(token,rc)
	if err!=nil{
		return nil,err
	}
	return rc,nil
}

func UpdateRunContext(token string, rc *api.RCInfo, datauuid string) error{
// add datauuid, EndTime
	rc.EndTime=core.GetCurTime()
	rc.OutputUuid=datauuid
	if err:=UpdateRunContext_API(token,rc);err!=nil{
		return err
	}
	return nil
}

