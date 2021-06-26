#include <string.h>
#include <assert.h>
#include <stdio.h>
#include <openssl/aes.h>
#include <stdlib.h>
#include <unistd.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <errno.h>
#include "cmfs.h"


const char *get_passwd(char *buf /*16bytes*/)
{
        char *pass;
        int i;
        memset(buf,0,AESBLOCK);
        pass=getpass("Input passwd:");
        for (i=0;i<16 && pass[i]!='\0'; i++)
                buf[i]=pass[i];
        return buf;
}

static int pad_buf(unsigned char* dst,int orgbytes/* offset in dst */) // return length  after pad
{
        int i;
        int padbytes=AESBLOCK-orgbytes%AESBLOCK;
        for(i=0;i<padbytes;i++){
                dst[orgbytes+i]=padbytes;
        }
        return padbytes+orgbytes;
}

static int unpad_buf(const unsigned char *src,int slen) // return original length,-1 on error
{
        unsigned int padsize=src[slen-1];
        if((slen-=padsize)<0)
        {
                printf("Error padd\n");
                return -1;
        }
    //    memcpy(dst,src,slen);
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

int decodeblk(const char* cibuf, const char* passwd,char* plbuf, int len,int last){
	int orglen=len;
//	char unpad[FILEBLOCK];
	decode(cibuf,passwd,plbuf,len);
	if(last)
		orglen=unpad_buf(plbuf,len);
	else
		assert(len==FILEBLOCK);
	return orglen;
}

int encodeblk(const char* plbuf, const char* passwd, char* cibuf, int len, int last)
{
	unsigned char padbuf[FILEBLOCK+AESBLOCK];
	if(last) {
		memcpy(padbuf,plbuf,len);
		len=pad_buf(padbuf,len);
		if(len<=FILEBLOCK)
			encode(padbuf,passwd,cibuf,len);
		else{
			encode(padbuf,passwd,cibuf,FILEBLOCK);
			encode(padbuf+FILEBLOCK,passwd,cibuf+FILEBLOCK,len-FILEBLOCK);
		}
	}else
		encode(plbuf,passwd,cibuf,len);
	return len;
}
