package main

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

long decodefd(int sfd,int dfd, const char* passwd){
	long i,blocks;
	off_t flen,total=0;
	int padlen,orglen;
	char buf[FILEBLOCK],plain[FILEBLOCK],unpad[FILEBLOCK];
	struct stat st;
	fstat(sfd,&st);
	flen=st.st_size;
	if(flen%AESBLOCK){
		printf("Warning: error file size,decoding may be wrong,cancelled.\n");
		return -1;
	}
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

void do_encode(const char* from, const char* dfile, const char *passwd)
{
	int sfd,dfd;
	sfd=open(from,O_RDONLY);
    struct stat st;
    fstat(sfd,&st);
	printf("%s->%s,paswd %s\n",from,dfile,passwd);
    dfd=creat(dfile,st.st_mode);
    if(dfd){
        printf("%ld bytes encoded\n",encodefd(sfd,dfd,passwd));
        close(dfd);
    }
    close(sfd);
}

*/
//#cgo LDFLAGS: -lssl -lcrypto
import "C"


import(
	"fmt"
	"flag"
	"os"
	"unsafe"
	"errors"
	"os/exec"
	"strings"
	"crypto/rand"
)

const (
	INVALID=iota
	ENCODE
	DISTRIBUTE
	MOUNT
)

const (
	RAWDATA=iota
	TAG
)

type EncryptedData struct{
	Uuid string
	Descr string
	FromType int
	FromObj string
	OwnerId int
	EncryptedKey []byte
}

const AES_KEY_LEN=128

var definpath , inpath string
var defoutpath,outpath string
var defuser, user string

func (linfo* LoginInfo)GetRawKey(src []byte)([]byte, error){
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
}


func LoadConfig(){
	definpath=os.Getenv("DATA_IN_PATH")
	defoutpath=os.Getenv("HOME")+"/.cmitdata"
	defuser=os.Getenv("DATA_DEF_USER")
}

func GetFunction() int {
	var bEnc,bShare,bMnt bool
	flag.BoolVar(&bEnc,"e",false,"encrypt raw data")
	flag.BoolVar(&bShare,"s",false,"share data to other users")
	flag.BoolVar(&bMnt,"m",false,"mount encrypted data")
	flag.StringVar(&inpath,"in",definpath,"original data path (may be a file or a directory)")
	flag.StringVar(&outpath,"out",definpath,"original data path (may be a file or a directory)")
	flag.StringVar(&user,"user",defuser, "login user name")
	flag.Parse()
	if bEnc{
		if (bShare || bMnt ==false){
			return ENCODE
		}else{
			return INVALID
		}
	}else if bShare{
		if bMnt==false{
			return DISTRIBUTE
		}else{
			return INVALID
		}
	}else if bMnt{
		return MOUNT
	}
	return INVALID
}

func EncodeDir(ipath string, opath string, user string) error{
	return nil
}

func RecordMetaFromRaw(pdata *EncryptedData ,linfo *LoginInfo, passwd []byte,ipath string, opath string)error{
	// passwd: raw passwd, need to be encrypted with linfo.Keylocalkey
	// RecordLocal && Record Remote
	return nil
}

func GetUuid()(string,error){
	if output,err:=exec.Command("uuidgen").Output();err!=nil{
		return "",err
	}else{
		return strings.TrimSpace(string(output)),nil
	}
}

func EncodeFile(ipath string, opath string, user string) error{
//	fmt.Println(ipath,opath,user)

	if user==""{
		fmt.Println("use parameter -user=NAME to set login user")
		return errors.New("empty user")
	}
	linfo,err:=Login(user)
	if err!=nil{
		fmt.Println("login error:",err)
		return err
	}

	passwd,err:=RandPasswd()
	if err!=nil{
		return nil
	}else{
	cpasswd:=(*C.char)(unsafe.Pointer(&passwd[0]))
	cipath:=C.CString(ipath)
	copath:=C.CString(opath)
//	tmpkey:=C.CString(linfo.Keylocalkey)
	defer C.free(unsafe.Pointer(cipath))
	defer C.free(unsafe.Pointer(copath))
//	defer C.free(unsafe.Pointer(tmpkey))
//	C.do_encode(cipath,copath,tmpkey) // just for test, next line is the real final way
	C.do_encode(cipath,copath,cpasswd)
	pdata:=new(EncryptedData)
	pdata.Uuid,_=GetUuid()
	pdata.Descr=""
	pdata.FromType=RAWDATA
	pdata.FromObj=ipath
	pdata.OwnerId=linfo.Id
	pdata.EncryptedKey=passwd
	RecordMetaFromRaw(pdata,linfo,passwd,ipath,opath)
	linfo.Logout()
	return nil
	}
}

func RandPasswd()([]byte,error){
	buf:=make([]byte,AES_KEY_LEN/8)
	if rdlen,err:=rand.Read(buf);rdlen==len(buf) && err==nil{
		fmt.Println(buf)
		return buf,nil
	}else {
		return nil,err
	}
}

func doEncode(){
	if inpath==""{
		fmt.Println("You should set inpath explicitly")
		return
	}
	if outpath==""{
		outpath="./"
	}
	if info,err:=os.Stat(inpath);err!=nil{
		fmt.Println("Can't find ",inpath)
		return
	}else{
		if info.IsDir(){
			EncodeDir(inpath,outpath,user)
		}else{
			EncodeFile(inpath,outpath,user)
		}
	}
}

func main(){
	LoadConfig()
	fun:=GetFunction()
	switch fun{
	case ENCODE:
		doEncode()
	case DISTRIBUTE:
	case MOUNT:
	default:
		fmt.Println("Error parameters,use -h or --help for usage")
	}
}
