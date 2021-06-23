#ifndef __CMFS_H__
#define __CMFS_H__

#define BLOCKSIZE 4096

#define O_WRITE(flags) ((flags) & (O_RDWR | O_WRONLY))
#define O_READ(flags)  (((flags) & (O_RDWR | O_RDONLY)) | !O_WRITE(flags))

#define U_ATIME (1 << 0)
#define U_CTIME (1 << 1)
#define U_MTIME (1 << 2)

#define AES_KEYLEN 128 
#define AESBLOCK (AES_KEYLEN/8)
#define FILEBLOCK 1024

#ifndef PATH_MAX
#define PATH_MAX 4096
#endif


typedef struct key_info{
    unsigned char crypt_key[AESBLOCK];
    unsigned char iv[AESBLOCK];
}KEY_INFO;


struct cmfs_options{
	char *mnt_point;
	char *src_dir;
	char *options;
	KEY_INFO keyinfo;
};

int decodeblk(const char* cibuf, const char* passwd,char* plbuf, int len,int last);
//int encode(const char* src, const char* passwd, char *dst,int len);
//int decode(const char* src, const char* passwd, char* dst,int len);
#endif
