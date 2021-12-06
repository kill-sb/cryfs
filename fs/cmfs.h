#ifndef __CMFS_H__
#define __CMFS_H__

#define O_WRITE(flags) ((flags) & (O_RDWR | O_WRONLY))
#define O_READ(flags)  (((flags) & (O_RDWR | O_RDONLY)) | !O_WRITE(flags))

#define U_ATIME (1 << 0)
#define U_CTIME (1 << 1)
#define U_MTIME (1 << 2)

#ifndef AES_KEYLEN
#define AES_KEYLEN 128 
#endif 

#ifndef AESBLOCK
#define AESBLOCK 16 
#endif

#ifndef FILEBLOCK
#define FILEBLOCK 1024
#endif

#ifndef PATH_MAX
#define PATH_MAX 4096
#endif


typedef struct key_info{
    unsigned char crypt_key[AESBLOCK];
    unsigned char iv[AESBLOCK];
}KEY_INFO;


struct cmfs_options{
	char mnt_point[PATH_MAX];
	char src_dir[PATH_MAX];
	char options[1024];
	KEY_INFO keyinfo;
};

#ifdef __DEBUG
#define LOG(str) {FILE *fp=fopen("/tmp/fs.log","a+");time_t t=time(NULL);fprintf(fp,"%s-%s:%d--> %s\n",ctime(&t),__FILE__,__LINE__,str);fclose(fp);}
#else
#define LOG(str)
#endif 

int decodeblk(const char* cibuf, const char* passwd,char* plbuf, int len,int last);
int encodeblk(const char* cibuf, const char* passwd,char* plbuf, int len,int last);
//int encode(const char* src, const char* passwd, char *dst,int len);
//int decode(const char* src, const char* passwd, char* dst,int len);
#endif
