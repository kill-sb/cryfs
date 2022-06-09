package main

import(
	"fmt"
	//"io"
	//"time"
	"os"
	"unsafe"
	"errors"
	"os/exec"
	"strings"
	core "coredata"
)
/*

#include <string.h>
#include <stdio.h>
#include <openssl/aes.h>
#include <stdlib.h>
#include <unistd.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <errno.h>

#define AES_KEYLEN 128
#define FILEBLOCK 1024
#define AESBLOCK 16 

#ifndef PATH_MAX
#define PATH_MAX 4096
#endif

int pad_buf(const char* src, char* dst,int orgbytes) // return length  after pad
{
	int i;
	int padbytes=AESBLOCK-orgbytes%AESBLOCK;
	if(orgbytes)
		memcpy(dst,src,orgbytes);
	for(i=0;i<padbytes;i++){
		dst[orgbytes+i]=padbytes;
	}
	return padbytes+orgbytes;
}

int unpad_buf(const unsigned char *src, char* dst,int slen) // return original length,-1 on error
{
	unsigned int padsize=(unsigned int) src[slen-1];
	if((slen-=padsize)<0)
	{
		printf("Error padd\n");
		return -1;
	}
	if(slen)
		memcpy(dst,src,slen);
	return slen;
}

void encode(const char* src, const char* passwd, char *dst,int len) // cbc only
{
  	AES_KEY aes;
	unsigned char iv[AESBLOCK] = {0};
	AES_set_encrypt_key(passwd,AES_KEYLEN,&aes);
	AES_cbc_encrypt(src,dst,len,&aes,iv,AES_ENCRYPT);
}

void decode(const char* src, const char* passwd, char* dst,int len)
{
	AES_KEY aes;
	unsigned char iv[AESBLOCK] = {0};
	AES_set_decrypt_key(passwd,AES_KEYLEN,&aes);
	AES_cbc_encrypt(src,dst,len,&aes,iv,AES_DECRYPT);
}

long encodefd(int sfd,int dfd, const char* passwd){
	off_t flen,total=0;
	struct stat st;
	long i,blocks;
	int leftf,lefta;
	char blockbuf[FILEBLOCK],padbuf[FILEBLOCK];
	char cibuf[FILEBLOCK];
	fstat(sfd,&st);
	flen=st.st_size;
	if(flen==0) return 0;
	blocks=flen/FILEBLOCK+1;
//	if(flen%FILEBLOCK)
//		blocks++;
	for(i=0;i<blocks-1;i++){ // last block will be padded later
		read(sfd,blockbuf,FILEBLOCK);
		encode(blockbuf,passwd,cibuf,FILEBLOCK);
		total+=write(dfd,cibuf,FILEBLOCK);
	}
	// process last few bytes(may be 0)
	leftf=read(sfd,blockbuf,FILEBLOCK);
	lefta=pad_buf(blockbuf,padbuf,leftf);
	encode(padbuf,passwd,cibuf,lefta);
	total+=write(dfd,cibuf,lefta);
	return total;
}

long decodefd(int sfd,int dfd, const char* passwd,off_t offset){
	long i,blocks;
	off_t flen,total=0;
	int padlen,orglen;
	char buf[FILEBLOCK],plain[FILEBLOCK],unpad[FILEBLOCK];
	struct stat st;
	fstat(sfd,&st);
	flen=st.st_size-offset;
	if(flen%AESBLOCK){
		printf("Warning: error file size,decoding may be wrong,cancelled.(flen:%d)\n",flen);
		return -1;
	}
	lseek(sfd,offset,SEEK_SET);
	blocks=flen/FILEBLOCK;
	if(flen%FILEBLOCK)
		blocks++;
	for(i=0;i<blocks-1;i++){
		total+=read(sfd,buf,FILEBLOCK);
		decode(buf,passwd,plain,FILEBLOCK);
		write(dfd,plain,FILEBLOCK);
	}
	padlen=read(sfd,buf,FILEBLOCK);
	decode(buf,passwd,plain,padlen);
	orglen=unpad_buf(plain,unpad,padlen);
	if(orglen>0){
		total+=write(dfd,unpad,orglen);
	}else if (orglen<0){
		printf("Error occured on unpadding,check your data\n");
		return -1;
	}
	return total;
}

void do_encodefile(const char* from, const char* dfile, const char *passwd)
{
	int sfd,dfd;
	sfd=open(from,O_RDONLY);
    struct stat st;
    fstat(sfd,&st);
//	printf("%s->%s\n",from,dfile);
    dfd=creat(dfile,st.st_mode);
    if(dfd){
//        printf("%ld bytes encoded\n",encodefd(sfd,dfd,passwd));
		encodefd(sfd,dfd,passwd);
        close(dfd);
    }
    close(sfd);
}

void do_decodefile(const char* from, const char* dfile, const char *passwd,off_t offset)
{
	int sfd,dfd;
	sfd=open(from,O_RDONLY);
    struct stat st;
    fstat(sfd,&st);
//	printf("%s->%s\n",from,dfile);
    dfd=creat(dfile,st.st_mode);
    if(dfd){
//		lseek(sfd,offset,SEEK_SET);
  //      printf("%ld bytes decoded\n",decodefd(sfd,dfd,passwd,offset));
  		decodefd(sfd,dfd,passwd,offset);
        close(dfd);
    }
    close(sfd);
}
*/
//#cgo LDFLAGS: -lssl -lcrypto
import "C"
/*
func GetRawKey(linfo *core.LoginInfo ,src []byte)([]byte, error){
    if(linfo.Keylocalkey!=nil && len(linfo.Keylocalkey)!=0){
        srclen:=len(src)
        dst:=make([]byte,srclen)
        csrc:=(*C.char)(unsafe.Pointer(&src[0]))
        cdst:=(*C.char)(unsafe.Pointer(&dst[0]))
        cpasswd:=(*C.char)(unsafe.Pointer(&linfo.Keylocalkey[0]))
        C.decode(csrc,cpasswd,cdst,C.int(srclen))
        return dst,nil
    }
    return nil,errors.New("Load key for decrypt localkey error")
}*/

func GetDataInfo(tag *core.EncDataTag)(*core.EncryptedData,error){
	return GetEncDataInfo(string(tag.Uuid[:]))
}

func GetEncDataFromDisk(linfo *core.LoginInfo,fname string)(*core.EncryptedData,*core.EncDataTag,error){
    tag,err:=core.LoadTagFromDisk(fname)
    if(err!=nil){
        return nil,nil,err
    }
    data,err:=GetDataInfo(tag)
	if(err!=nil){
		fmt.Println("GetDataInfo error",err)
		return nil,nil,err
	}
    data.Path=fname
    DoDecodeInC(tag.EKey[:],linfo.Keylocalkey,data.EncryptingKey,16)
	return data,tag,nil
}

func GetFileMd5(fname string)(string,error){
	if output,err:=exec.Command("md5sum",fname).Output();err!=nil{
		return "" ,err
	}else{
		return (strings.Split(string(output)," "))[0],nil
	}
}

func GetFileSha256(fname string)(string,error){
	if output,err:=exec.Command("sha256sum",fname).Output();err!=nil{
		return "" ,err
	}else{
		return (strings.Split(string(output)," "))[0],nil
	}
}

func SaveLocalFileTag(pdata* core.EncryptedData, savedkey []byte)(*core.EncDataTag,error){
	tag:=new (core.EncDataTag)
	copy(tag.Uuid[:],[]byte(pdata.Uuid))
	copy(tag.EKey[:],savedkey)
	tag.SaveTagToDisk(pdata.Path+"/"+pdata.Uuid+".tag")
	return tag,nil
}

/*
func SaveLocalFileTag(pdata* core.EncryptedData, savedkey []byte)(*core.TagInFile,error){
	tag:=new (core.TagInFile)
	tag.OwnerId=pdata.OwnerId
	for k,v:=range []byte(pdata.Uuid){
		tag.Uuid[k]=v
	}
	for i,j:=range []byte(pdata.HashMd5){
		tag.Md5Sum[i]=j
	}
	tag.FromType=byte(pdata.FromType)
	for k,v:=range []byte(pdata.FromObj){
		tag.FromObj[k]=v
	}

	for k,v:=range []byte(pdata.OrgName){
		tag.OrgName[k]=v
	}

	tag.IsDir=pdata.IsDir
	tag.Time=time.Now().Unix()
	for k,v:=range []byte(savedkey){
		tag.EKey[k]=v
	}

	copy(tag.Descr[:],[]byte(pdata.Descr))
	//copy(tag.Descr[:],"cmit encrypted raw data")
	tag.SaveTagToDisk(pdata.Path+"/"+pdata.Uuid+".tag")
	return tag,nil
}

func SendMetaToServer_API(pdata *core.EncryptedData, token string)error{
	encreq:=api.EncDataReq{Token:token,Uuid:pdata.Uuid,Descr:pdata.Descr,IsDir:pdata.IsDir,FromType:pdata.FromType,FromObj:pdata.FromObj,OwnerId:pdata.OwnerId,Hash256:pdata.HashMd5,OrgName:pdata.OrgName}
    ack:=new (api.IEncDataAck)
	err:=HttpAPIPost(&encreq,ack,"newdata")
    if err!=nil{
        fmt.Println("call api error:",err)
    }
	if ack.Code!=0{
		return errors.New(ack.Msg)
	}
	return err
}*/

func DoEncodeInC(src,passwd,dst []byte,length int){
	csrc:=(*C.char)(unsafe.Pointer(&src[0]))
	cpasswd:=(*C.char)(unsafe.Pointer(&passwd[0]))
	cdst:=(*C.char)(unsafe.Pointer(&dst[0]))
	C.encode(csrc,cpasswd,cdst,C.int(length))
}

func DoDecodeInC(src, passwd, dst []byte,length int){
	csrc:=(*C.char)(unsafe.Pointer(&src[0]))
	cpasswd:=(*C.char)(unsafe.Pointer(&passwd[0]))
	cdst:=(*C.char)(unsafe.Pointer(&dst[0]))
	C.decode(csrc,cpasswd,cdst,C.int(length))
}

func RecordMetaFromRaw(pdata *core.EncryptedData ,keylocalkey []byte, passwd []byte,ipath string, opath string,token string)error{
	// passwd: raw passwd, need to be encrypted with linfo.Keylocalkey
	// RecordLocal && Record Remote
	savedkey:=make([]byte,128/8)
	DoEncodeInC(passwd , keylocalkey ,savedkey,128/8)
	SaveLocalFileTag(pdata,savedkey)
	SendMetaToServer_API(pdata,token)
	return nil
}

func GetFileName(ipath string)(string,error){
	finfo,err:=os.Stat(ipath)
	if err!=nil{
		return "",err
	}
	return finfo.Name(),nil
}

func DoEncodeFileInC(infile,outfile string,passwd []byte )error{
	cpasswd:=(*C.char)(unsafe.Pointer(&passwd[0]))
	cipath:=C.CString(infile)
	cofile:=C.CString(outfile)
	defer C.free(unsafe.Pointer(cipath))
	defer C.free(unsafe.Pointer(cofile))
	C.do_encodefile(cipath,cofile,cpasswd)
	return nil
}

func EncodeFile(ipath string, opath string, linfo *core.LoginInfo) (string,error){
//	fmt.Println(ipath,opath,user)
	passwd,err:=core.RandPasswd()
	if err!=nil{
		return "",err
	}
	fname,err:=GetFileName(ipath)
	if err!=nil{
		return "",err
	}

	pdata:=new(core.EncryptedData)
	pdata.Uuid,_=core.GetUuid()
	pdata.Descr="cmit encrypted data"
	pdata.OrgName=fname
	pdata.OwnerId=linfo.Id
	pdata.EncryptingKey=passwd
	pdata.Path=opath
	pdata.IsDir=0
	pdata.FromRCId=0
	pdata.FromContext=nil
	ofile:=opath+"/"+pdata.Uuid
/*	cpasswd:=(*C.char)(unsafe.Pointer(&passwd[0]))
	cipath:=C.CString(ipath)
	cofile:=C.CString(ofile)
	defer C.free(unsafe.Pointer(cipath))
	defer C.free(unsafe.Pointer(cofile))
	C.do_encodefile(cipath,cofile,cpasswd)
	*/
	DoEncodeFileInC(ipath,ofile,passwd)
//	pdata.HashMd5,_=GetFileMd5(ofile)
	RecordMetaFromRaw(pdata,linfo.Keylocalkey,passwd,ipath,opath,linfo.Token)
	return pdata.Uuid,nil
}

func doDecode(){
	if inpath==""{
		fmt.Println("You should set inpath explicitly")
		return
	}
	if outpath==""{
		outpath="./"
	}else{
		os.MkdirAll(outpath,0755)
	}
	finfo,err:=os.Stat(inpath)
	if err!=nil{
		fmt.Println("Can't find ",inpath)
		return
	}else{
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
/*
		if info.IsDir(){
			fmt.Println("Decoding dir",inpath,outpath)
			DecodeDir(inpath,outpath,linfo)
		}else{
			fmt.Println("Decoding file",inpath,outpath)
			DecodeFile(inpath,outpath,linfo)
		}*/
		DecodeFile(inpath,outpath,linfo,finfo)
	}
}

func DecodeFile(ipath,opath string,linfo *core.LoginInfo,finfo os.FileInfo)error{
	// judge raw uuid file or csd file
	ftype,err:=core.GetFileType(ipath)
	if err!=nil{
		fmt.Println("Get file type error",err)
		return err
	}
	switch ftype{
	case core.ENCDATA:
		return DecodeRawData(finfo,linfo,ipath,opath)
	case core.CSDFILE:
		return DecodeCSDFile(linfo,ipath,opath)
	default:
		fmt.Println("Unknow filetype of ",ipath,"---",ftype)
		return errors.New("Unknown filetype")
	}
}

func DecodeCSDFile(linfo *core.LoginInfo,ipath,opath string) error{
	head,err:=core.LoadShareInfoHead(ipath)
	if err!=nil{
		fmt.Println("Load share info head error:",err)
		return err
	}
	// now we have got a valid csd header, then load info from server
	sinfo,err:=GetShareInfoFromHead(head,linfo)
	if err!=nil{
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
	ofile:=sinfo.OrgName
//	fmt.Println("Get ofile ",ofile)
//	fmt.Println("enc keys:",core.BinkeyToString(sinfo.EncryptedKey),"randkey:",core.BinkeyToString(sinfo.RandKey))
	ofile=opath+"/"+ofile
	orgkey:=make([]byte,16)

	DoDecodeInC(sinfo.EncryptedKey,sinfo.RandKey,orgkey,16)
	var ret error
	if sinfo.IsDir==0{
/*		cpasswd:=(*C.char)(unsafe.Pointer(&orgkey[0]))
		cipath:=C.CString(ipath)
		cofile:=C.CString(ofile)
		defer C.free(unsafe.Pointer(cipath))
		defer C.free(unsafe.Pointer(cofile))
		C.do_decodefile(cipath,cofile,cpasswd,60) // ShareInfoHead offset
		*/
		ret=DoDecodeFileInC(ipath,ofile,orgkey,60)
	}else{
			// todo: it's a zipped dir
		ret=DecodeCSDToDir(ipath,ofile,orgkey)
	}
	if ret==nil{
		fmt.Println(ofile,"decoded ok")
	}
	return ret
}

func DoDecodeFileInC(ifile,ofile string, passwd []byte,offset int64)error{
		cpasswd:=(*C.char)(unsafe.Pointer(&passwd[0]))
		cipath:=C.CString(ifile)
		cofile:=C.CString(ofile)
		defer C.free(unsafe.Pointer(cipath))
		defer C.free(unsafe.Pointer(cofile))
		C.do_decodefile(cipath,cofile,cpasswd,C.long(offset)/* ShareInfoHead offset*/)
		return nil
}

func DecodeRawData(finfo os.FileInfo,linfo *core.LoginInfo,ipath,opath string)error{
	// todo : load tag, decode file
	tag,err:=core.LoadTagFromDisk(ipath)
	if err!=nil{
		fmt.Println("load tag information error",err)
		return err
	}

	pdata,_:=GetDataInfo(tag)
	if pdata.OwnerId!=linfo.Id{
		fmt.Println("The data does not belong to",linfo.Name)
		return errors.New("Invalid user")
	}
	pdata.Path=ipath
	DoDecodeInC(tag.EKey[:],linfo.Keylocalkey,pdata.EncryptingKey,16)
	ofile:=opath+"/"+pdata.OrgName
	if pdata.IsDir==0{
		DoDecodeFileInC(ipath,ofile,pdata.EncryptingKey,0)
	}else{
		err=os.MkdirAll(ofile,finfo.Mode())
		if err!=nil{
			fmt.Println("mkdir error:",err)
		}
		DecodeDir(ipath,ofile,pdata.EncryptingKey)
	}

	fmt.Println(ofile,"restored ok")
	return nil
}

func doEncode(){
	if inpath==""{
		fmt.Println("You should set inpath explicitly")
		return
	}
	if outpath==""{
		outpath="./"
	}else{
		os.MkdirAll(outpath,0755)
	}
	if info,err:=os.Stat(inpath);err!=nil{
		fmt.Println("Can't find ",inpath)
		return
	}else{
		if loginuser==""{
			fmt.Println("use parameter -user to set login user")
			return
		}
		linfo,err:=Login(loginuser)
		if err!=nil{
			fmt.Println("login error:",err)
			return
		}
		defer Logout(linfo)

		if info.IsDir(){
			if uuid,err:=EncodeDir(inpath,outpath,linfo); err==nil{
				fmt.Println("Encode ok, uuid:",uuid)
			}else{
				fmt.Println("Encode error:",err)
			}
		}else{
			if uuid,err:=EncodeFile(inpath,outpath,linfo); err==nil{
				fmt.Println("Encode ok, uuid:",uuid)
			}else{
				fmt.Println("Encode error:",err)
			}
		}
	}
}

func ValidSepPath(ipath,opath string)(os.FileInfo,string,string,error){
	finfo,err:=os.Stat(opath)
	if err!=nil{
		return nil,"","",err
	}
	dinfo,err:=os.Stat(ipath)
	if err!=nil{
		return nil,"","",err
	}
	basedir:=strings.TrimSuffix(core.StripAllSlash(ipath),dinfo.Name())
	relstr:=strings.TrimPrefix(opath,ipath)
	if relstr==opath{
		return finfo,"","",errors.New(opath+" is not in "+ipath)
	}
	relstr=strings.TrimPrefix(relstr,"/")
	fmt.Println("validation ret:",relstr,basedir)
	return finfo,relstr,basedir,nil
}

/*
func doSep(){
	if inpath=="" || outpath==""{
		fmt.Println("You should set inpath end outpath explicitly")
		return
	}
	finfo,relname,basedir,err:=ValidSepPath(inpath,outpath)
	if err!=nil{
		fmt.Println("Invalid path or filename",err)
		return
	}
	// copy dst file out using a new uuid filename, reuse random encrypted key, modify fromtype and fromobj, copy orgname
	if loginuser==""{
		fmt.Println("use parameter -user to set login user")
		return
	}
	linfo,err:=Login(loginuser)
	if err!=nil{
		fmt.Println("login error:",err)
		return
	}
	defer Logout(linfo)

	dinfo,dtag,err:=GetEncDataFromDisk(linfo,inpath)
	if err!=nil{
		fmt.Println("Get EncDataFromDisk error in doSep error:",err)
		return
	}
	dst:=new (core.EncryptedData)
	dst.Uuid,_=core.GetUuid()
	dst.Descr=relname+" from "+inpath
	dst.FromType=dinfo.FromType
	dst.FromObj=dinfo.FromObj
	dst.OwnerId=dinfo.OwnerId
	dst.HashMd5,_=GetFileMd5(outpath)
	dst.EncryptingKey=dinfo.EncryptingKey
	dst.Path=basedir
	dst.IsDir=0
	dst.OrgName=finfo.Name()

	efile,err:=os.OpenFile(dst.Path+"/"+dst.Uuid,os.O_CREATE|os.O_RDWR,finfo.Mode())
	if err!=nil{
		fmt.Println("Create file error in doSep:",err)
		return
	}
	defer efile.Close()
	ifile,err:=os.Open(outpath)
	if err!=nil{
		fmt.Println("Open file error:",err)
		return
	}
	defer ifile.Close()
	io.Copy(efile,ifile)
	SaveLocalFileTag(dst,dtag.EKey[:])

    SendMetaToServer_API(dst,linfo.Token)

	fmt.Println(dst.Uuid,"seperated ok")
}
*/
