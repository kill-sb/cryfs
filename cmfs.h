#ifndef __CMFS_H__
#define __CMFS_H__

#define BLOCKSIZE 4096

#define O_WRITE(flags) ((flags) & (O_RDWR | O_WRONLY))
#define O_READ(flags)  (((flags) & (O_RDWR | O_RDONLY)) | !O_WRITE(flags))

#define U_ATIME (1 << 0)
#define U_CTIME (1 << 1)
#define U_MTIME (1 << 2)

#ifndef PATH_MAX
#define PATH_MAX 4096
#endif


#include "crypt.h"

typedef struct key_info{
    unsigned char crypt_key[AES_KEYLEN/8];
    unsigned char iv[AES_KEYLEN/8];
}KEY_INFO;


struct cmfs_options{
	char mnt_point[PATH_MAX];
	char src_dir[PATH_MAX];
	char options[1024];
	KEY_INFO keyinfo;
};

#ifdef __DEBUG
#define LOG(str) {FILE *fp=fopen("/tmp/fs.log","a+");fprintf(fp,"%s\n",str);fclose(fp);}
#else
#define LOG(str)
#endif 

int decodeblk(const char* cibuf, const char* passwd,char* plbuf, int len,int last);
int encodeblk(const char* cibuf, const char* passwd,char* plbuf, int len,int last);
//int encode(const char* src, const char* passwd, char *dst,int len);
//int decode(const char* src, const char* passwd, char* dst,int len);
#endif
