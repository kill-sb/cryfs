package main

import(
	"fmt"
	"time"
	"os"
	"unsafe"
	"errors"
	"os/exec"
	"strings"
	"dbop"
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
		printf("Warning: error file size,decoding may be wrong,cancelled.\n");
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
	printf("%s->%s\n",from,dfile);
    dfd=creat(dfile,st.st_mode);
    if(dfd){
        printf("%ld bytes encoded\n",encodefd(sfd,dfd,passwd));
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
	printf("%s->%s\n",from,dfile);
    dfd=creat(dfile,st.st_mode);
    if(dfd){
//		lseek(sfd,offset,SEEK_SET);
        printf("%ld bytes encoded\n",decodefd(sfd,dfd,passwd,offset));
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

func GetEncDataFromDisk(linfo *core.LoginInfo,fname string)(*core.EncryptedData,error){
    tag,err:=core.LoadTagFromDisk(fname)
    if(err!=nil){
        return nil,err
    }
    data,err:=tag.GetDataInfo()
	if(err!=nil){
		fmt.Println("GetDataInfo error",err)
		return nil,err
	}
    data.Path=fname
    DoDecodeInC(tag.EKey[:],linfo.Keylocalkey,data.EncryptingKey,16)
	return data,nil
}

func GetFileMd5(fname string)(string,error){
	if output,err:=exec.Command("md5sum",fname).Output();err!=nil{
		return "" ,err
	}else{
		return (strings.Split(string(output)," "))[0],nil
	}
}

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
	tag.IsDir=pdata.IsDir
	tag.Time=time.Now().Unix()
	for k,v:=range []byte(savedkey){
		tag.EKey[k]=v
	}

	copy(tag.Descr[:],"cmit encrypted raw data")
	tag.SaveTagToDisk(pdata.Path+"/"+pdata.Uuid+".tag")
	return tag,nil
}

func SendMetaToServer(pdata *core.EncryptedData)error{
	dbop.SaveMeta(pdata)
	return nil
}

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

func RecordMetaFromRaw(pdata *core.EncryptedData ,keylocalkey []byte, passwd []byte,ipath string, opath string)error{
	// passwd: raw passwd, need to be encrypted with linfo.Keylocalkey
	// RecordLocal && Record Remote
	savedkey:=make([]byte,128/8)
	DoEncodeInC(passwd , keylocalkey ,savedkey,128/8)
	SaveLocalFileTag(pdata,savedkey)
	SendMetaToServer(pdata)
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

func EncodeFile(ipath string, opath string, linfo *core.LoginInfo) error{
//	fmt.Println(ipath,opath,user)
	passwd,err:=core.RandPasswd()
	if err!=nil{
		return err
	}
	fname,err:=GetFileName(ipath)
	if err!=nil{
		return err
	}
	pdata:=new(core.EncryptedData)
	pdata.Uuid,_=core.GetUuid()
	pdata.Descr=""
	pdata.FromType=core.RAWDATA
	pdata.FromObj=fname
	pdata.OwnerId=linfo.Id
	pdata.EncryptingKey=passwd
	pdata.Path=opath
	pdata.IsDir=0

	ofile:=opath+"/"+pdata.Uuid
/*	cpasswd:=(*C.char)(unsafe.Pointer(&passwd[0]))
	cipath:=C.CString(ipath)
	cofile:=C.CString(ofile)
	defer C.free(unsafe.Pointer(cipath))
	defer C.free(unsafe.Pointer(cofile))
	C.do_encodefile(cipath,cofile,cpasswd)
	*/
	DoEncodeFileInC(ipath,ofile,passwd)
	pdata.HashMd5,_=GetFileMd5(ofile)
	RecordMetaFromRaw(pdata,linfo.Keylocalkey,passwd,ipath,opath)
	return nil
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
	if info,err:=os.Stat(inpath);err!=nil{
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
		defer linfo.Logout()

		if info.IsDir(){
			fmt.Println("Decoding dir",inpath,outpath)
			DecodeDir(inpath,outpath,linfo)
		}else{
			fmt.Println("Decoding file",inpath,outpath)
			DecodeFile(inpath,outpath,linfo)
		}
	}
}

func DecodeFile(ipath,opath string,linfo *core.LoginInfo)error{
	// judge raw uuid file or csd file
	ftype,err:=core.GetFileType(ipath)
	if err!=nil{
		fmt.Println("Get file type error",err)
		return err
	}
	switch ftype{
	case core.RAWDATA:
		return DecodeRawData(linfo,ipath,opath)
	case core.CSDFILE:
		return DecodeCSDFile(linfo,ipath,opath)
	default:
		fmt.Println("Unknow filetype of ",ipath)
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
	sinfo,err:=dbop.LoadShareInfo(head)
	if(err!=nil){
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
	ofile,err:=dbop.GetOrgFileName(sinfo)
	fmt.Println("Get ofile ",ofile)
	fmt.Println("enc keys:",core.BinkeyToString(sinfo.EncryptedKey),"randkey:",core.BinkeyToString(sinfo.RandKey))
	if err!=nil{
		fmt.Println("Get origin file name error",err)
		return err
	}
	ofile=opath+"/"+ofile
	orgkey:=make([]byte,16)

	if sinfo.IsDir==0{
		DoDecodeInC(sinfo.EncryptedKey,sinfo.RandKey,orgkey,16)
		fmt.Println("psd file",ipath,"key",core.BinkeyToString(orgkey))
/*		cpasswd:=(*C.char)(unsafe.Pointer(&orgkey[0]))
		cipath:=C.CString(ipath)
		cofile:=C.CString(ofile)
		defer C.free(unsafe.Pointer(cipath))
		defer C.free(unsafe.Pointer(cofile))
		C.do_decodefile(cipath,cofile,cpasswd,60) // ShareInfoHead offset
		*/
		DoDecodeFileInC(ipath,ofile,orgkey,60)
	}else{
			// todo: it's a zipped dir
	}
	return nil
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

func DecodeRawData(linfo *core.LoginInfo,ipath,opath string)error{
	// todo : load tag, decode file
	tag,err:=core.LoadTagFromDisk(ipath)
	if err!=nil{
		fmt.Println("load tag information error",err)
		return err
	}

	pdata,_:=tag.GetDataInfo()
	pdata.Path=ipath

	if(pdata.FromType==core.RAWDATA){
		if pdata.IsDir==0{
		DoDecodeInC(tag.EKey[:],linfo.Keylocalkey,pdata.EncryptingKey,16)
		fmt.Println("raw file",pdata.Uuid,"key",core.BinkeyToString(pdata.EncryptingKey))
		ofile:=opath+"/"+pdata.FromObj
/*		cpasswd:=(*C.char)(unsafe.Pointer(&pdata.EncryptingKey[0]))
		cipath:=C.CString(ipath)
		cofile:=C.CString(ofile)
		defer C.free(unsafe.Pointer(cipath))
		defer C.free(unsafe.Pointer(cofile))
		C.do_decodefile(cipath,cofile,cpasswd,0)
		*/
		DoDecodeFileInC(ipath,ofile,pdata.EncryptingKey,0)
		}else{
			// todo: it's a dir, not zipped
		}
	}else {
		//if pdata.FromType==core.CSDFILE,should be same with above, FromType is used only for trace source, the file is still a raw encrypted file
		// pdata.EncryptingKey need to be filled first
//		rootdata:=LoadRootData(pdata)
		fmt.Println(ipath,"is not valid raw encrypted data")
		return errors.New("Invalid encrypted file")
	}
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
		defer linfo.Logout()

		if info.IsDir(){
			EncodeDir(inpath,outpath,linfo)
		}else{
			EncodeFile(inpath,outpath,linfo)
		}
	}
}

