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

const char *get_passwd(char *buf /* AESBLOCK bytes*/)
{
	char *pass;
	int i;
	memset(buf,0,AESBLOCK);
	pass=getpass("Input passwd:");
	for (i=0;i<AESBLOCK && pass[i]!='\0'; i++)
		buf[i]=pass[i];
	return buf;
}

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

long encodefile(int sfd,int dfd, const char* passwd){
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

long decodefile(int sfd,int dfd, const char* passwd){
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

int main(int c,char**v)
{
	char dfile[PATH_MAX];
	char passwd[AESBLOCK];
	int sfd,dfd;
	struct stat st;
	int enc=-1;
	if(c<3) {
		printf("Usage: %s -e(enc)/-d(dec) sourcefile\n",v[0]);
		return 1;
	}
	sfd=open(v[2],O_RDONLY);
	if(sfd<0){
		printf("Can't open file\n");
		return -errno;
	}
	if(strcmp(v[1],"-e")==0){
		sprintf(dfile,"%s.aes",v[2]);
		enc=1;
	}else if(strcmp(v[1],"-d")==0){
		sprintf(dfile,"%s.org",v[2]);
		enc=0;
	}
	if(enc<0){
		printf("error parameter\n");
		close(sfd);
		return enc;
	}
	fstat(sfd,&st);
	dfd=creat(dfile,st.st_mode);
	if(dfd){
		get_passwd(passwd);
		if(enc){
			printf("%ld bytes encoded\n",encodefile(sfd,dfd,passwd));
		}else {
			printf("%ld bytes decoded\n",decodefile(sfd,dfd,passwd));
		}
		close(dfd);
	}
	close(sfd);
	return 0;
}
