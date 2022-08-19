package main

/*
#include <stdio.h>
#include <unistd.h>
#include <stdlib.h>
#include <fcntl.h>

int mount_cmfs(const char* cmfsbin,const char* src, const char* dst,const char* passwd, const char* opt){

    int fd[2];
    int p=pipe(fd);
    if(fork()==0){
        close(fd[1]);
        dup2(fd[0],0);
        execlp(cmfsbin,cmfsbin,src,dst,"-o","kernel_cache","-o","big_writes", "-o", "max_write=65535", "-o",opt,NULL); // TODO verify checksum later
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
	"path/filepath"
	core "coredata"
	api  "apiv1"
)

type MountOpt struct{
	dstpt string
	access string
}

type InputDataInfo struct{
	datatype int
	uuid string
	srcdir string
	dstdir string
	orgname string
}


func PrepareMntOpts(inputs []*InputDataInfo, tool string, output string) map[string]*MountOpt{
	rdroot:="/readonly"
	retmap:=make(map[string]*MountOpt)
	for i,idata:=range inputs{
		opt:=new(MountOpt)
		opt.dstpt=rdroot+fmt.Sprintf("/input/%d-%s",i+1,idata.orgname)
		opt.access="ro"
		retmap[idata.dstdir]=opt
	}
	if tool!=""{
		topt:=new(MountOpt)
		topt.dstpt=rdroot+"/tool"
		topt.access="ro"
		retmap[tool]=topt
	}
	if output!=""{
		oopt:=new(MountOpt)
		oopt.dstpt="/output"
		oopt.access="rw"
		retmap[output]=oopt
	}
	return retmap
}

func LocalTempDir(ipath string,usepdir bool)string{
	// return a dir that must be in the same block device with given path, invoker will add a unique name to the dirname later
	finfo,err:=os.Stat(ipath)
	if err!=nil{
		fmt.Println(ipath, "not exists")
		return ipath
	}
	if !usepdir{
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
		apath,err:=filepath.Abs(str)
		//apath,err:=filepath.Abs(strings.TrimSpace(str))
		if err==nil{
			if(apath!=""){
				multisrc=append(multisrc,apath)
			}
		}else{
			fmt.Println("Invalid input file:",str)
			return
		}
	}

    if outpath=="" && oname!="" {
        fmt.Println("You should set outpath explicitly")
		return
    }

	if outpath!=""{
		apath,err:=filepath.Abs(outpath)
		if err!=nil{
			fmt.Println("Invalid output path:",outpath)
			return
		}else{
			outpath=apath
		}
	}

	if mntimport!=""{
		apath,err:=filepath.Abs(mntimport)
		if err!=nil{
			fmt.Println("Invalid import-tool path:",mntimport)
			return
		}
		mntimport=apath
	}
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

	MountObjs(linfo,multisrc,strings.TrimSpace(mntimport))
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
	return srcdir,dstdir,key,nil
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
			relpath:=strings.TrimPrefix(path,toolpath)
			st,err=os.Stat(path)
			if err!=nil{
				fmt.Println("Can't stat:",path,",ignored")
				return nil
			}
			imnode:=new(api.ImportFile)
			if imnode.Sha256,err=GetFileSha256(path);err!=nil{
				fmt.Println("Can't get sha256sum of: ",path,",ignored")
				return nil
			}
			imnode.RelName=relpath
			if desc,err:=exec.Command("file",path).Output();err!=nil{
				imnode.FileDesc=""
			}else{
				imnode.FileDesc=strings.TrimSpace(string(desc))
			}
			imnode.Size=st.Size()
			imlist=append(imlist,imnode)
        }
		return nil
	})
	return imlist,nil
}

func ValidateInputs(linfo *core.LoginInfo,inputs []string)(bool, error){
	// check readonly, limit times
	rdonly:=false
	for _,idata:=range inputs{
		st,err:=os.Stat(idata)
		if err!=nil{
			return false,errors.New("Invalid inputdata:"+idata)
		}
		dtype:=GetDataType(idata)
		switch dtype{
		case core.CSDFILE:
			head,err:=core.LoadShareInfoHead(idata)
			if err!=nil{
				return rdonly,err
			}
			sinfo,err:=GetShareInfoFromHead(head,linfo,0)
			if err!=nil{
				return rdonly,errors.New("Bad csd file: "+idata)
			}
			if sinfo.Sha256!=""{
				sha256,_:=GetFileSha256(idata)
				if strings.ToLower(sha256)!=strings.ToLower(sinfo.Sha256){
					return rdonly,errors.New(idata+" invalid sha256 sum")
				}
			}

			if sinfo.Perm & 1==0 || sinfo.LeftUse!=-1 || !strings.HasPrefix(sinfo.Expire,"2999-12-31") {
				rdonly=true
			}
			// check readonly
			if sinfo.LeftUse==0{
				return rdonly,errors.New(idata+"  invalid user or open times exhaused")
			}

			strexp:=strings.Replace(sinfo.Expire," ","T",1)+"+08:00"
			tmexp,err:=time.Parse(time.RFC3339,strexp)
			if err!=nil{
				fmt.Println("Parse expire time error:",err)
				return rdonly,errors.New("Invalid expire time format: "+idata)
			}
			tmnow:=time.Now()
			if tmnow.After(tmexp){
				fmt.Println("The shared data has expired at :",sinfo.Expire)
				return rdonly,errors.New(idata+" data expired")
			}

		case core.ENCDATA:
			dinfo,err:=GetEncDataInfo(st.Name())
			if err!=nil{
				return rdonly,err
			}
			if dinfo.OwnerId!=linfo.Id{
				return rdonly, errors.New(idata+" does not belong to current user")
			}
		default:
			return rdonly, errors.New("Invalid inputdata:"+idata)
		}
	}
	return rdonly,nil
}

func MountCSDFile(linfo *core.LoginInfo, filepath string)(*InputDataInfo,error){
	head,err:=core.LoadShareInfoHead(filepath)
	if err!=nil{
		fmt.Println("Load share info head in MountFile error:",err)
		return nil,err
	}
	sinfo,err:=GetShareInfoFromHead(head,linfo,1)
	if err!=nil{
		fmt.Println("Load share info from head error:",err)
		return nil,err
	}
	ret:=new(InputDataInfo)
	ret.datatype=core.CSDFILE
	ret.uuid=sinfo.Uuid
	ret.orgname=sinfo.OrgName

	sinfo.FileUri=filepath
	ret.srcdir,ret.dstdir,err=PrepareCSDFileDir(filepath,sinfo)
	return ret,err
}

func MountEncDir(linfo *core.LoginInfo, dirname string)(*InputDataInfo,error){
	tag,err:=core.LoadTagFromDisk(dirname)
	if err!=nil{
		fmt.Println("LoadTagFromDisk error in MountDir:",err)
		return nil,err
	}

	dinfo,err:=GetDataInfo(tag)
	if err!=nil{
		fmt.Println("GetDataInfo in MountEncFile error:",err)
		return nil,err
	}

	ret:=new(InputDataInfo)
	ret.datatype=core.ENCDATA
	ret.uuid=dinfo.Uuid
	ret.orgname=dinfo.OrgName

/*	
	process:
	1. get enc passwd
	2. mkdir tmpdir 
	3. mount encdir -> tmpdir with passwd
*/
	passwd:=make([]byte,16)
	DoDecodeInC(tag.EKey[:],linfo.Keylocalkey,passwd,16)
	dstuuid,_:=core.GetUuid()
	tmpdir:=LocalTempDir(dirname,true)+"."+dstuuid
	err=os.MkdirAll(tmpdir,0755)
	err=MountDirInC(dirname,tmpdir,passwd,"ro")
	if err!=nil{
		fmt.Println("mount dir in c error:",err)
		return ret,err
	}
	ret.dstdir=tmpdir
	return ret,nil // trick here: srcdir should be "" , avoid  being deleted later
}

func MountEncFile(linfo *core.LoginInfo, filepath string)(*InputDataInfo,error){
	tag,err:=core.LoadTagFromDisk(filepath)
	if err!=nil{
		fmt.Println("LoadTagFromDisk error in MountEncFile:",err)
		return nil,err
	}
	dinfo,err:=GetDataInfo(tag)
	if err!=nil{
		fmt.Println("GetDataInfo in MountEncFile error:",err)
		return nil,err
	}
	ret:=new(InputDataInfo)
	ret.datatype=core.ENCDATA
	ret.uuid=dinfo.Uuid
	ret.orgname=dinfo.OrgName
/*	
	process:
	get enc passwd
	mkdir tmpdir ->srcdir,dstdir
	mount srcdir -> dstdir with passwd
*/
	passwd:=make([]byte,16)
	DoDecodeInC(tag.EKey[:],linfo.Keylocalkey,passwd,16)

	ret.srcdir,ret.dstdir,err=PrepareEncFileDir(filepath,dinfo.OrgName)

	if err!=nil{
		fmt.Println("Prepare dir error:",err)
		return ret,err
	}
	err=MountDirInC(ret.srcdir,ret.dstdir,passwd,"ro")
	if err!=nil{
		fmt.Println("mount dir in c error:",err)
		return ret,err
	}
	return ret,nil
}

func MountEncData(linfo *core.LoginInfo,inputs []string)([]*InputDataInfo,error){
	var err error
	inputdata:=make([]*InputDataInfo,0,20)
	for _,srcdata:=range inputs{
		dtype:=GetDataType(srcdata)
		var idata *InputDataInfo
		switch dtype{
		case core.CSDFILE:
			idata,err=MountCSDFile(linfo,srcdata)
			if err!=nil{
				return inputdata,err
			}
		case core.ENCDATA:
			st,_:=os.Stat(srcdata)
			if st.IsDir(){
				idata,err=MountEncDir(linfo,srcdata)
			}else{
				idata,err=MountEncFile(linfo,srcdata)
			}
			if err!=nil{
				return inputdata,err
			}
		default:
			return inputdata, errors.New("Invalid inputdata:"+srcdata)
		}
		inputdata=append(inputdata,idata)
	}
	return inputdata,nil
}

func CreateOrgRC(linfo *core.LoginInfo,inputdata []*InputDataInfo, tlinfo []* api.ImportFile,podimg string)*api.RCInfo{
	ninput:=len(inputdata)
	rcinfo:=new (api.RCInfo)
	rcinfo.UserId=linfo.Id
	rcinfo.InputData=make([]*api.SourceObj,ninput,ninput)
	for i:=0;i<ninput;i++{
		iinfo:=new (api.SourceObj)
		iinfo.DataType=inputdata[i].datatype
		iinfo.DataUuid=inputdata[i].uuid
		rcinfo.InputData[i]=iinfo
	}
	rcinfo.ImportPlain=tlinfo
	rcinfo.OS="linux"
	rcinfo.BaseImg=podimg
	rcinfo.StartTime=core.GetCurTime()
	return rcinfo
}

func CmfsValid()bool{
	cmfspath:=core.GetSelfPath()+"/cmfs"
	cmsum,err:=GetFileSha256(cmfspath)
	if err!=nil{
		return false
	}
	sum:=strings.Split(CmfsSum," ")[0]
    if cmsum!=sum{
		return false
    }
	return true
}

func RegisterRC(linfo *core.LoginInfo, rc *api.RCInfo , outuuid string)(*api.RCInfo,error){
	rc.OutputUuid=outuuid
	rc.EndTime=core.GetCurTime()
	return CreateRunContext_API(linfo.Token,rc)
}

func MountObjs(linfo *core.LoginInfo, inputs []string, tool string){
/*
	1. check local paths valid
	2. check data access permission, any readonly data will cause not output path(mount return different value),then mount them with different key
	3. prepare local input/output dir(do cmfs mount)
	4. put all cmfs-mounted dir and import dir in a slice, then bind in-paths to /mnt/input/N-names, bind import-path to /mnt/output, bind tool-dir to /mnt/tool to pod
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

	if !CmfsValid(){
		fmt.Println("Invalid cmfs file")
		return
	}
	// establish output dir, and mount to a temp dir with random passwd
	var outsrc,outdst string
	var outkey []byte
	outitems:=0
	if outpath!=""{
		outsrc,outdst,outkey,err=PrepareOutputDir(outpath)
		if err!=nil{
			fmt.Println("Output dir validation error:",err)
			return
		}
		if outdst!=""{
			defer func(){
				exec.Command("umount",outdst).Run()
				os.RemoveAll(outdst)
				if outitems==0{
					os.RemoveAll(outsrc)
				}
			}()
		}
	}
	fmt.Println("Processing input data...")
	inputinfo,err:=MountEncData(linfo,inputs)

    defer func(){
		for _,idata:=range inputinfo{
			if idata.dstdir!=""{
				exec.Command("umount",idata.dstdir).Run()
				os.RemoveAll(idata.dstdir)
			}
			if idata.srcdir!=""{
				os.RemoveAll(idata.srcdir)
			}
		}
    }()

	if err!=nil{
		fmt.Println("Mount data error:",err)
		return
	}

	// fill input,output and tool dir mount info to MountOpt map
	opts:=PrepareMntOpts(inputinfo,tool,outdst)

	// need not send to server now
	rc:=CreateOrgRC(linfo,inputinfo,tlinfo,podimg)
	fmt.Println("Creating container...")
	err=RunPod(podimg,opts)
	if err!=nil{
		fmt.Println("Create container error:",err)
		return
	}
	if outpath!="" && outdst!=""{
		var outuuid string
		var isdir bool
		outuuid,isdir,outitems,err=ProcOutputData(outsrc) // will umount and check single file or multi file(dir), return new data uuid (outsrc may be renamed or removed, according to if output data is a dir or empty)
		if err!=nil{
			fmt.Println("ProcOutputData error:",err)
			return
		}
		if outuuid!="" && outitems!=0{
			fmt.Println("Processing output data...")
			rc,err=RegisterRC(linfo,rc,outuuid)
			outorgname:=outuuid
			if err==nil{
				if oname!=""{
					outorgname=oname
				}
				err=RecordDataInfo(linfo,outuuid,outkey,outorgname,rc,isdir) // invoke createrc and newdata
				if err!=nil{
					fmt.Println("UpdateMetaInfo error:",err)
				}
			}else{
					fmt.Println("Register rc error:",err)
			}
			if ouid!=0{
				ChOwner(outpath+"/"+outuuid,true)
			}
			fmt.Println("New data generated:",outuuid)
		}else{
			fmt.Println("Empty output, no new data registered")
		}
	}

	// tools dir need not be cleaned now, but when it is implemented by a COW filesystem, clean work need to be done here
}

func ProcOutputData(outsrc string)(/*uuid*/ string,/*isdir */bool,/*items*/ int, error){
	if dataname,items,singlefile:=ItemsInDir(outsrc);singlefile{// ItemsInDir may change oname
		newuuid,_:=core.GetUuid()
		os.Link(dataname,outpath+"/"+newuuid)
		os.Remove(dataname)
		if err:=os.Remove(outsrc);err!=nil{
			fmt.Println("Error in ProcOutputData(dir not empty->single file delete failed?):",err)
			return "",false,items,err
		}
		return newuuid,false,items,nil
	}else{
		if items==0{
			return "",false,0,nil
		}
		finfo,err:=os.Stat(outsrc)
		if err!=nil{
			return "",true,items,err
		}
		fname:=finfo.Name()
		return fname/*uuid*/,true,items,nil
	}
}

func RunPod(podimg string,opts map[string]*MountOpt)error{
	ctcmd:="podman run -it --rm --network none --privileged=true "
	for k,v:=range opts{
		ctcmd=ctcmd+" -v "+k+":'"+v.dstpt+"':"+v.access
	}
	ctcmd=ctcmd+" "+podimg+" /bin/bash"
	ccmd:=C.CString(ctcmd)
	defer C.free(unsafe.Pointer(ccmd))
	C.system(ccmd)
	return nil
}

func MountDirInC(src,dst string,passwd []byte,mode string)error{
	cpasswd:=(*C.char)(unsafe.Pointer(&passwd[0]))
	csrc:=C.CString(src)
	cdst:=C.CString(dst)
	copt:=C.CString(mode)
	cmfspath:=core.GetSelfPath()+"/cmfs"
	cmfsbin:=(C.CString)(cmfspath)
	defer C.free(unsafe.Pointer(cmfsbin))
	defer C.free(unsafe.Pointer(csrc))
	defer C.free(unsafe.Pointer(cdst))
	defer C.free(unsafe.Pointer(copt))
	C.mount_cmfs(cmfsbin,csrc,cdst,cpasswd,copt)
	return nil
}

func PrepareCSDFileDir(filepath string, sinfo *core.ShareInfo)(string,string,error){
//	fmt.Println("Preparing container environment...")
    uuid,_:=core.GetUuid()
    tmpdir:=LocalTempDir(filepath,true)+"."+uuid
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
	plaindir:=LocalTempDir(filepath,true)+"."+uuid
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

func ItemsInDir(dirname string)(/*single filename*/ string,/*items*/ int,/*singlefile*/ bool){

	cont,err:=ioutil.ReadDir(dirname)
	if err!=nil{
		return "",0,false
	}
	items:=len(cont)
	if items==1{
		sname:=cont[0].Name()
		if oname==""{
			oname=sname
		}
		if !cont[0].IsDir(){
			return strings.TrimSuffix(dirname,"/")+"/"+sname,1,true
		}
	}
	return "",items,false
}

func RecordDataInfo(linfo *core.LoginInfo,outuuid string ,outkey []byte, orgname string ,rc *api.RCInfo,isdir bool)error{
    pdata:=new(core.EncryptedData)
    pdata.Uuid=outuuid
	if isdir{
		pdata.Descr="cmit encrypted dir"
	}else{
		pdata.Descr="cmit encrypted data"
	}
	if orgname==""{
		pdata.OrgName=outuuid
	}else{
		pdata.OrgName=orgname
	}
    pdata.OwnerId=linfo.Id
    pdata.EncryptingKey=outkey
    pdata.Path=outpath
	if isdir{
		pdata.IsDir=1
	}else{
		pdata.IsDir=0
	}
	pdata.FromRCId=rc.RCId
	pdata.FromContext=rc
	return RecordMetaFromRaw(pdata,linfo.Keylocalkey,outkey,linfo.Token)
}

func PrepareEncFileDir(ipath string,fname string)(string,string,error){
	srcuuid,_:=core.GetUuid()
	dstuuid,_:=core.GetUuid()
	tmpdir:=LocalTempDir(ipath,true)
	srcdir:=tmpdir+"."+srcuuid
	dstdir:=tmpdir+"."+dstuuid
	os.MkdirAll(srcdir,0755)
	os.MkdirAll(dstdir,0755)
	return srcdir,dstdir,os.Link(ipath,srcdir+"/"+fname)
}

func CreateRunContext(token string,baseimg string,srcobj []*api.SourceObj, tools []*api.ImportFile)(*api.RCInfo,error){
//	rc:=new (core.RunContext)
	rc:=new (api.RCInfo)
	rc.BaseImg=baseimg
	rc.OS="Linux"
	rc.InputData=srcobj
	rc.ImportPlain=tools
	rc.StartTime=core.GetCurTime()
	rc,err:=CreateRunContext_API(token,rc)
	if err!=nil{
		return nil,err
	}
	return rc,nil
}

func UpdateRunContext(token string, rc *api.RCInfo, datauuid string) error{
// add datauuid, EndTime
	rc.EndTime=core.GetCurTime()
	rc.OutputUuid=datauuid
	return UpdateRunContext_API(token,rc)
}

