
#define FUSE_USE_VERSION 26
#include <assert.h>
#include <fuse.h>
#include <stdlib.h>
#include <getopt.h>
#include <stdio.h>
#include <string.h>
#include <unistd.h>
#include <fcntl.h>
#include <sys/stat.h>
#include <errno.h>
#include <dirent.h>
#include <sys/time.h>
#include <libgen.h>
#include <stddef.h>

#include "cmfs.h"

static struct cmfs_options g_opts;

char * safe_dirname(const char *msg) {
  char *buf = strdup(msg);
  char *dir = dirname(buf);
  char *res = strdup(dir);
  free(buf);
  return res;
}

char * safe_basename(const char *msg) {
  char *buf = strdup(msg);
  char *nam = basename(buf);
  char *res = strdup(nam);
  free(buf);
  return res;
}

static int get_realname(char *buf, const char* path)
{
	int len;
	int ret=0;
	sprintf(buf,"%s%s",g_opts.src_dir,path);
	len=strlen(buf);
	return ret;
}


static int readblk(int fd,off_t blk,char *buf, int needupad)
{
	int rd,decode;
	char cibuf[FILEBLOCK];
	rd=pread(fd,cibuf,FILEBLOCK,blk*FILEBLOCK);
	if (rd<=0) return rd;
	decode=decodeblk(cibuf,g_opts.keyinfo.crypt_key,buf,rd,needupad);
	return decode;
}

static int writeblk(int fd, off_t blk, const char *buf, int start,int slen /* plaintext length*/,int needpad)
{
	char cibuff[FILEBLOCK+AESBLOCK]; // may need another full AESBLOCK to pad
	int elen=encodeblk(buf,g_opts.keyinfo.crypt_key,cibuff,slen,needpad);
	if(elen>start){
		elen-=start;
	}else{
		char log[1024];
		sprintf(log,"writeblk: start- %d, len- %d, elen- %d");
		LOG(log);
		return  0;
	}
	return pwrite(fd,cibuff+start,elen,blk*FILEBLOCK+start);
}

static size_t get_realsize(const char* realpath, size_t srclen){
	int endblk=srclen/FILEBLOCK;
	char plain[FILEBLOCK];
	int length;
	int left=srclen%FILEBLOCK;
	int fd;
	if(srclen==0) return 0;
	fd=open(realpath,O_RDONLY);
	if(fd<0) return srclen;
	if(left==0){
		left=FILEBLOCK;
	   	endblk--;
	}
	length=readblk(fd,endblk,plain,1);
	close(fd);

	if(length<left && left-length<=AESBLOCK)
		srclen-=left-length;
	return srclen;
}

static size_t get_realsize_fd(int fd, size_t srclen)
{
	int endblk=srclen/FILEBLOCK;
	char plain[FILEBLOCK];
	int length;
	int left=srclen%FILEBLOCK;
	if(srclen==0) return 0;
	if(fd<0) return srclen;
	if(left==0){
		left=FILEBLOCK;
	   	endblk--;
	}
	length=readblk(fd,endblk,plain,1);

	if(length<left && left-length<=AESBLOCK)
		srclen-=left-length;
	return srclen;

}

static int cmfs_utimens(const char *path, const struct timespec ts[2]) {
	
	char dst[PATH_MAX];
	int fd;
	int ret;
	get_realname(dst,path);
	fd=open(dst,O_WRONLY);
	if(fd<0) return -1;
	ret=futimens(fd,ts);
	close(fd);
  	return ret;
}

static int cmfs_unlink(const char *path) {

	char dst[PATH_MAX];
	get_realname(dst,path);
	if(unlink(dst)<0)
		return -errno;
	return 0;
}


static int cmfs_symlink(const char *from, const char *to) {
	char dst[PATH_MAX];
	get_realname(dst,to);
	if(symlink(from,dst)<0)
		return -errno;
	return 0;
}

static int cmfs_release(const char *path, struct fuse_file_info *fi) {
	if(fi->fh>=0)
		close(fi->fh);
 	return 0;
}

static int cmfs_chmod(const char *path, mode_t mode) {
	char dst[PATH_MAX];
	get_realname(dst,path);
	if(chmod(dst,mode)<0)
		return -errno;
	return 0;
}

static int cmfs_mknod(const char *path, mode_t mode, dev_t rdev) {
	char dst[PATH_MAX];
	get_realname(dst,path);
	if (mknod(dst,mode,rdev)<0)
		return -errno;
	return 0;
}

static int cmfs_rename(const char *from, const char *to) {
	char dstfrom[PATH_MAX],dstto[PATH_MAX];
	get_realname(dstfrom,from);
	get_realname(dstto,to);
  	if(rename(dstfrom,dstto)<0)
		return -errno;
	return 0;
}

static int cmfs_chown(const char *path, uid_t uid, gid_t gid) {
	char dst[PATH_MAX];
	get_realname(dst,path);
	if(lchown(dst,uid,gid)<0)
		return -errno;
	return 0;
}

static int cmfs_rmdir(const char *path) {
	char dst[PATH_MAX];
	get_realname(dst,path);
	if (rmdir(dst)<0)
		return -errno;
	return 0;
}

static int cmfs_getattr(const char *path, struct stat *stbuf) {
  	int ret;
	char dst[PATH_MAX];
	get_realname(dst,path);
  	if((ret=lstat(dst,stbuf))){
	  	return -errno;
  	}
	if(S_ISREG(stbuf->st_mode))
		stbuf->st_size=get_realsize(dst,stbuf->st_size);
  	return 0;
}

static int cmfs_mkdir(const char *path, mode_t mode) {
	char dst[PATH_MAX];
	get_realname(dst,path);
	if(mkdir(dst,mode)<0)
		return -errno;
	return 0;
}

static int cmfs_readdir(const char *path, void *buf, fuse_fill_dir_t filler, off_t offset, struct fuse_file_info *fi) {
  char dstname[PATH_MAX];
  DIR* dir;
  struct dirent * drt;
  sprintf(dstname,"%s%s",g_opts.src_dir,path);
  dir=opendir(dstname);
  if(!dir){
	  return ENOTDIR;
  }
  	for(drt=readdir(dir);drt!=NULL;drt=readdir(dir))
  	{
		strcpy(dstname,drt->d_name);
		filler(buf,dstname,NULL,0);
	}
  closedir(dir);
  return 0;
}

static int cmfs_readlink(const char *path, char *buf, size_t size) {
  	int ret;
	char dst[PATH_MAX],link[PATH_MAX];
	
	get_realname(dst,path);
	ret=readlink(dst,link,size);
	if (ret<=0){
		return -errno;
	}
	if(ret>size-1)
		ret=size-1;

	memcpy(buf,link,ret);
	buf[ret]='\0';
	return 0;
}

static int cmfs_open(const char *path, struct fuse_file_info *fi) {
	char dst[PATH_MAX];
	struct stat st;
	int fd;
	int ret;
	get_realname(dst,path);

	ret=stat(dst,&st);
	if(ret){
		return -errno;
	}
	if(S_ISDIR(st.st_mode)) {

		return -EISDIR;
	}

	fd=open(dst,O_RDWR);
	fi->fh = fd;
	return 0;
}

static int cmfs_read(const char *path, char *buf, size_t size, off_t offset, struct fuse_file_info *fi) {
	int ret,i;
	off_t iblk,totalrd=0,end;
	off_t lastfileblk;
	off_t startblk,endblk;// start from 0(consider as skip blocks)
	char cibuf[FILEBLOCK],plbuf[FILEBLOCK];
	struct stat st; 
	int firstread;
	int startbyte,lastbyte;
	int startcpy,endcpy; // start and end byte in memcpy between return buf --"buf" and decrypted block --"plbuf"

	int rd; // real read bytes 
	int de; // decrypted plaintext length (in 1k block) 
	

	// Check whether the file was opened for reading
	if(!O_READ(fi->flags)) {
		return -EACCES;
	}

//	return pread(fi->fh,buf,size,offset);
	
	if(fstat(fi->fh,&st)<0)
		return -EACCES;
	if(size<=0 || offset>=st.st_size)
		return 0;
	lastfileblk=st.st_size/FILEBLOCK;	
	if(st.st_size%FILEBLOCK==0)
		lastfileblk--;
	startblk=offset/FILEBLOCK;
	startbyte=offset%FILEBLOCK;
	end=offset+size;
	if(end>st.st_size)
		end=st.st_size;

	endblk=end/FILEBLOCK;
	lastbyte=end%FILEBLOCK;
	if(lastbyte==0)// offset <st_size, so treat the position as last byte of last block
	{
		lastbyte=FILEBLOCK;
		endblk--;
	}

	// 1. process first block(start with offset%FILEBLOCK) and check if is the last block.
	// 2. process mid blocks which are all full blocks
	// 3. process last block 

	if(startblk==endblk)
	{
		if((rd=readblk(fi->fh,startblk,plbuf,1))<=0)
			return rd;
		if(lastbyte>rd)
			lastbyte=rd;
		totalrd=lastbyte-startbyte;
		memcpy(buf+startbyte,plbuf,totalrd);
		return totalrd;
		
	}
	if((rd=readblk(fi->fh,startblk,plbuf,0))<=0)
	{
		assert("readblk error");
		return rd;
	}

	// read full block at first
	totalrd+=(FILEBLOCK-offset%FILEBLOCK);
	memcpy(buf+startbyte,plbuf,totalrd);

	// process mid blocks,simply fully  read/copy
	for(iblk=startblk+1;iblk<endblk;iblk++)
	{
		if((rd=readblk(fi->fh,iblk,plbuf,0)<FILEBLOCK))
			return -1; // error occured
		memcpy(buf+totalrd,plbuf,FILEBLOCK);
		totalrd+=FILEBLOCK;
	}

	// process last block
	if (endblk==lastfileblk)
		rd=readblk(fi->fh,endblk,plbuf,1);	
	else
		rd=readblk(fi->fh,endblk,plbuf,0);
	if (rd<0)	
	{
		return -1;
	}
	if (rd) // the last block may has a AESBLOCK size and totally filled with padding data, so readblk will return 0
	{
		if(rd<lastbyte)
			lastbyte=rd;
		memcpy(buf+totalrd,plbuf,lastbyte);
		totalrd+=lastbyte;
	}
	return totalrd;
}

static int extend_file(int fd, off_t size,off_t fsize);
static int native_write(const char *buf, size_t size, off_t offset,int fd)
{
	int ret,i;
	off_t iblk,totalwr=0,end;
	off_t lastfileblk;
	int firstbyte,endbyte;// byte offset in block
	off_t startblk,endblk;// start from 0(consider as skip blocks)
	char cibuf[FILEBLOCK],plbuf[FILEBLOCK]={0};
	struct stat st; 
	int startcpy,endcpy; // start and end byte in memcpy between return buf --"buf" and decrypted block --"plbuf"
	size_t realsize;

	int rd; // real read bytes 
	int de; // decrypted plaintext length (in 1k block) 
	

	//ret=pwrite(fi->fh,buf,size,offset);

	if(fstat(fd,&st)<0)
		return -EACCES;
	if(size<=0)
		return 0;
	realsize=get_realsize_fd(fd,st.st_size);
	//  when current blk>lastfileblk => needrd==0
	lastfileblk=st.st_size/FILEBLOCK;	
	if(st.st_size>0 && st.st_size%FILEBLOCK==0)
		lastfileblk--;
	startblk=offset/FILEBLOCK;
	end=offset+size;
	endblk=end/FILEBLOCK; 

	firstbyte=offset%FILEBLOCK;
	endbyte=end%FILEBLOCK;
	if(end>0 && endbyte==0)
	{
		endblk--;
		endbyte=FILEBLOCK;
	}


// static int writeblk(int fd, off_t blk, const char *buf, int slen ,int needpad); // buf should start from beginning of a block,but not byte start to be encrypted, slen is whole buf size(startbyte+size  or FILEBLOCK),because left bytes should be all reencrypted

	// first step , process first block

	if(startblk==endblk){ // only one block,use "size" in memcpy
		if(startblk>lastfileblk) // need not readblk
		{
			//extend_file(int fd, off_t size,off_t fsize);
			extend_file(fd,end,realsize);
			memcpy(plbuf+firstbyte,buf,size);
			writeblk(fd,startblk,plbuf,firstbyte,endbyte,1);
		}else if(startblk==lastfileblk){
			if((rd=readblk(fd,startblk,plbuf,1))<=0)
				memset(plbuf,0,FILEBLOCK);
			memcpy(plbuf+firstbyte,buf,size);
			if(rd>=endbyte) // left bytes need to be reencrypted, but do not need repadding
				writeblk(fd,startblk,plbuf,firstbyte,rd,1);
			else // overwrite the end , need repadding
				writeblk(fd,startblk,plbuf,firstbyte,endbyte,1);
		}else{// not lastfileblock
			if((rd=readblk(fd,startblk,plbuf,0))<=0) 
				memset(plbuf,0,FILEBLOCK);
			memcpy(plbuf+firstbyte,buf,size);
			writeblk(fd,startblk,plbuf,firstbyte,FILEBLOCK,0);
		}
		return size;
	}else{ // need not pad, more blocks will follow
		if(startblk>lastfileblk){
			extend_file(fd,end,realsize);
			memcpy(plbuf+firstbyte,buf,FILEBLOCK-firstbyte);
			writeblk(fd,startblk,plbuf,firstbyte,FILEBLOCK,0);
		}else if(startblk==lastfileblk){
			if((rd=readblk(fd,startblk,plbuf,1))<=0)
				memset(plbuf,0,FILEBLOCK);
			memcpy(plbuf+firstbyte,buf,FILEBLOCK-firstbyte);
			writeblk(fd,startblk,plbuf,firstbyte,FILEBLOCK,0);
		}else{ // mid blocks of file
			if((rd=readblk(fd,startblk,plbuf,0))<=0)
				memset(plbuf,0,FILEBLOCK);
			memcpy(plbuf+firstbyte,buf,FILEBLOCK-firstbyte);
			writeblk(fd,startblk,plbuf,firstbyte,FILEBLOCK,0);
		}
	}

	// mid block, write whole block
	for (iblk=startblk+1;iblk<endblk;iblk++){		
		memcpy(plbuf,buf+(FILEBLOCK-firstbyte)+(iblk-startblk-1)*FILEBLOCK,FILEBLOCK);
		writeblk(fd,iblk,plbuf,0,FILEBLOCK,0);
	}

	// last block -- endblk, and must not be firstblk
	memset(plbuf,0,FILEBLOCK);
	if(endblk>lastfileblk){// simply memcpy
		memcpy(plbuf,buf+(FILEBLOCK-firstbyte)+(endblk-startblk-1)*FILEBLOCK,endbyte);
		writeblk(fd,endblk,plbuf,0,endbyte,1);
	}else if(endblk==lastfileblk){
		rd=readblk(fd,endblk,plbuf,1);
		memcpy(plbuf,buf+(FILEBLOCK-firstbyte)+(endblk-startblk-1)*FILEBLOCK,endbyte);
		if(rd>endbyte)
			writeblk(fd,endblk,plbuf,0,rd,1);
		else
			writeblk(fd,endblk,plbuf,0,endbyte,1);
	}else{ // in mid-file blocks
		rd=readblk(fd,endblk,plbuf,0);
		memcpy(plbuf,buf+(FILEBLOCK-firstbyte)+(endblk-startblk-1)*FILEBLOCK,endbyte);
		writeblk(fd,endblk,plbuf,0,FILEBLOCK,0);// notice, here should write FILEBLOCK , not endbyte.because left bytes should be recrypted
	}


	return size;
}

static int cmfs_write(const char *path, const char *buf, size_t size, off_t offset,struct fuse_file_info *fi)
{
	if(!O_WRITE(fi->flags) ) {
		return -EACCES;
	}

	return native_write(buf,size,offset,fi->fh);
}

static int shrink_file(int fd, off_t size,off_t fsize){
	off_t blk=size/FILEBLOCK,lastfileblk=fsize/FILEBLOCK;
	int endbyte=size%FILEBLOCK;
	char plbuf[FILEBLOCK];
	int rd;
	int ret;
	if(endbyte==0) /* assert (size>0) */ 
	{
		endbyte=FILEBLOCK;
		blk--;
	}
	if(fsize%FILEBLOCK==0)
		lastfileblk--;
	if(blk==lastfileblk)
		rd=readblk(fd,blk,plbuf,1);
	else 
		rd=readblk(fd,blk,plbuf,0);
	if(rd<=0)
		return -1;
	ret=ftruncate(fd,size);
	if (ret==0)
		writeblk(fd,blk,plbuf,0,endbyte,1);
	return ret;
}

static int extend_file(int fd, off_t size,off_t fsize){
	char buf[FILEBLOCK]={0};
	off_t len=size-fsize;// len>0
	off_t startblk=fsize/FILEBLOCK, endblk=size/FILEBLOCK,iblk;
	int startbyte=fsize%FILEBLOCK, endbyte=size%FILEBLOCK;
	if(endbyte==0){
		endbyte=FILEBLOCK;
		endblk--;
	}

//static int native_write(const char *buf, size_t size, off_t offset,int fd)
	if(startblk==endblk)
		native_write(buf,len,fsize,fd);
	else{// more than 1 block
		native_write(buf,FILEBLOCK-startbyte,fsize,fd);
		for(iblk=startblk+1;iblk<endblk;iblk++)
			native_write(buf,FILEBLOCK,iblk*FILEBLOCK,fd);
		native_write(buf,endbyte,endblk*FILEBLOCK,fd);
	}
	return 0;
}

static int cmfs_truncate(const char *path, off_t size) {
	char dst[PATH_MAX];
	int ret;
	struct stat st;
	off_t fsize;
	int fd;
	get_realname(dst,path);
	if(size<=0)
		return truncate(dst,size);
	ret=stat(dst,&st);
	if(ret){
		return -errno;
	}
	if(S_ISDIR(st.st_mode)) {
		return -EISDIR;
	}
    fsize=get_realsize(dst,st.st_size);
	if((fd=open(dst,O_RDWR))<0)
		return -1;
	if(size<fsize){
		ret=shrink_file(fd,size,fsize);
	}else if (size>fsize){
		ret=extend_file(fd,size,fsize);
	}
	close(fd);
	return ret;
}


static int cmfs_create(const char *path, mode_t mode, struct fuse_file_info *fi)
{
    int retstat = 0;
    char fpath[PATH_MAX];

	sprintf(fpath,"%s%s",g_opts.src_dir,path);
	
    fi->fh = creat(fpath, mode);
    if (fi->fh < 0)
        retstat = -errno;
	else
    	retstat=fi->fh;

    return retstat;
}

static struct fuse_operations cmfs_oper = {
  .getattr      = cmfs_getattr,
  .readlink     = cmfs_readlink,
  .readdir      = cmfs_readdir,
  .open         = cmfs_open,
  .read         = cmfs_read,
  .write		= cmfs_write,
//  .create		= cmfs_create,
  .unlink       = cmfs_unlink,
  .chmod        = cmfs_chmod,
  .chown        = cmfs_chown,
  .mkdir        = cmfs_mkdir,
  .mknod        = cmfs_mknod,
  .truncate     = cmfs_truncate,
  .rmdir        = cmfs_rmdir,
  .symlink      = cmfs_symlink,
//  .link			= cmfs_link,
  .rename       = cmfs_rename,
  .release      = cmfs_release,
  .utimens      = cmfs_utimens,
};

void cmfs_init(struct fuse_args* args)
{
    char *pass;
    int i;
    memset(g_opts.keyinfo.crypt_key,0,AESBLOCK);
    pass=getpass("Input passwd:");
    for (i=0;i<AESBLOCK-1 && pass[i]!='\0';i++)
	{ 	
		g_opts.keyinfo.crypt_key[i]=pass[i];
	}
	memcpy(g_opts.keyinfo.iv,g_opts.keyinfo.crypt_key,AESBLOCK);
}

//
// Application entry point
//

int main(int argc, char *argv[]) {
//	parse_option//	
	struct fuse_args args = FUSE_ARGS_INIT(argc, argv);
/*	if(g_opts.mnt_point==NULL || g_opts.src_dir==NULL){
		printf("Too few arguments.\n");
		return -1;
	}*/

	if( argc<3){
		printf("Usage: cmfs <srcdir> <mntpoint> [options]\n");
  		return fuse_main(argc, argv, &cmfs_oper, NULL);
	}
	g_opts.src_dir=(char*)malloc(PATH_MAX+1);
	strcpy(g_opts.src_dir,argv[1]);
	g_opts.mnt_point=(char*)malloc(PATH_MAX+1);
	strcpy(g_opts.mnt_point,argv[2]);
	argv[1]=argv[0];
	

  	cmfs_init(&args);
  	return fuse_main(argc-1, argv+1, &cmfs_oper, NULL);
}

