
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

#define CMFS_OPT(t, p, v) { t, offsetof(struct cmfs_options, p), v }
/*
#define OPTION(t, p)                           \
    { t, offsetof(struct cmfs_options, p), 1 }
static const struct fuse_opt option_spec[] = {
	OPTION("--name=%s", filename),
	OPTION("--contents=%s", contents),
	OPTION("-h", show_help),
	OPTION("--help", show_help),
	FUSE_OPT_END
};
*/
//
// Utility functions
//

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
/*	if(len>4 && strcmp(buf+len-4,".cmc")==0)
	{
		buf[len-4]='\0';
		ret=1;
	}*/
	return ret;
}


/*
int getnodebypath(const char *path, struct filesystem *fs, struct node **node) {
  return getnoderelativeto(path, fs->root, node);
}

static void update_times(struct node *node, int which) {
  time_t now = time(0);
  if(which & U_ATIME) node->vstat.st_atime = now;
  if(which & U_CTIME) node->vstat.st_ctime = now;
  if(which & U_MTIME) node->vstat.st_mtime = now;
}

static int createentry(const char *path, mode_t mode, struct node **node) {
  char *dirpath = safe_dirname(path);

  // Find parent node
  struct node *dir;
  int ret = getnodebypath(dirpath, &the_fs, &dir);
  free(dirpath);
  if(!ret) {
    return -errno;
  }

  // Create new node
  *node = malloc(sizeof(struct node));
  if(!*node) {
    return -ENOMEM;
  }

  (*node)->fd_count = 0;
  (*node)->delete_on_close = 0;

  // Initialize stats
  if(!initstat(*node, mode)) {
    free(*node);
    return -errno;
  }

  struct fuse_context *ctx = fuse_get_context();
  (*node)->vstat.st_uid = ctx->uid;
  (*node)->vstat.st_gid = ctx->gid;

  // Add to parent directory
  if(!dir_add_alloc(dir, safe_basename(path), *node, 0)) {
    free(*node);
    return -errno;
  }

  return 0;
}

static int memfs_rmdir(const char *path) {
  char *dirpath, *name;
  struct node *dir, *node;

  // Find inode
  if(!getnodebypath(path, &the_fs, &node)) {
    return -errno;
  }

  if(!S_ISDIR(node->vstat.st_mode)) {
    return -ENOTDIR;
  }

  // Check if directory is empty
  if(node->data != NULL) {
    return -ENOTEMPTY;
  }

  dirpath = safe_dirname(path);

  // Find parent inode
  if(!getnodebypath(dirpath, &the_fs, &dir)) {
    free(dirpath);
    return -errno;
  }

  free(dirpath);

  name = safe_basename(path);

  // Find directory entry in parent
  if(!dir_remove(dir, name)) {
    free(name);
    return -errno;
  }

  free(name);

  free(node);

  return 0;
}

static int memfs_symlink(const char *from, const char *to) {
  struct node *node;
  int res = createentry(to, S_IFLNK | 0766, &node);
  if(res) return res;

  node->data = strdup(from);
  node->vstat.st_size = strlen(from);

  return 0;
}

// TODO: Adapt to description: https://linux.die.net/man/2/rename
static int memfs_rename(const char *from, const char *to) {
  char *fromdir, *fromnam, *todir, *tonam;
  struct node *node, *fromdirnode, *todirnode;

  if(!getnodebypath(from, &the_fs, &node)) {
    return -errno;
  }

  fromdir = safe_dirname(from);

  if(!getnodebypath(fromdir, &the_fs, &fromdirnode)) {
    free(fromdir);
    return -errno;
  }

  free(fromdir);

  todir = safe_dirname(to);

  if(!getnodebypath(todir, &the_fs, &todirnode)) {
    free(todir);
    return -errno;
  }

  free(todir);

  tonam = safe_basename(to);

  // TODO: When replacing, perform the same things as when unlinking
  if(!dir_add_alloc(todirnode, tonam, node, 1)) {
    free(tonam);
    return -errno;
  }

  free(tonam);

  fromnam = safe_basename(from);

  if(!dir_remove(fromdirnode, fromnam)) {
    free(fromnam);
    return -errno;
  }

  free(fromnam);

  return 0;
}

static int memfs_link(const char *from, const char *to) {
  char *todir, *tonam;
  struct node *node, *todirnode;

  if(!getnodebypath(from, &the_fs, &node)) {
    return -errno;
  }

  todir = safe_dirname(to);

  if(!getnodebypath(todir, &the_fs, &todirnode)) {
    free(todir);
    return -errno;
  }

  free(todir);

  tonam = safe_basename(to);

  if(!dir_add_alloc(todirnode, tonam, node, 0)) {
    free(tonam);
    return -errno;
  }

  free(tonam);

  return 0;
}

static int memfs_utimens(const char *path, const struct timespec ts[2]) {
  struct node *node;
  if(!getnodebypath(path, &the_fs, &node)) {
    return -errno;
  }

  node->vstat.st_atime = ts[0].tv_sec;
  node->vstat.st_mtime = ts[1].tv_sec;

  return 0;
}


static int memfs_release(const char *path, struct fuse_file_info *fi) {
  struct filehandle *fh = (struct filehandle *) fi->fh;

  // If the file was deleted but we could not free it due to open file descriptors,
  // free the node and its data after all file descriptors have been closed.
  if(--fh->node->fd_count == 0) {
    if(fh->node->delete_on_close) {
      if(fh->node->data) free(fh->node->data);
      free(fh->node);
    }
  }

  // Free "file handle"
  free(fh);

  return 0;
}
*/

static int readblk(int fd,off_t blk,char *buf, int needupad)
{
	int rd,decode;
	char cibuf[FILEBLOCK];
	rd=pread(fd,cibuf,FILEBLOCK,blk*FILEBLOCK);
	if (rd<=0) return rd;
	decode=decodeblk(cibuf,g_opts.keyinfo.crypt_key,buf,rd,needupad);
	return decode;
}

static int writeblk(int fd, off_t blk, const char *buf, int startbyte, int slen /* plaintext length*/,int needpad)
{
	char cibuff[FILEBLOCK+AESBLOCK]; // may need another full AESBLOCK to pad
	int elen=encodeblk(buf,g_opts.keyinfo.crypt_key,cibuff,slen+startbyte,needpad);
	return pwrite(fd,cibuff+startbyte,elen-startbyte,blk*FILEBLOCK+startbyte);
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

static int cmfs_unlink(const char *path) {

	char dst[PATH_MAX];
	get_realname(dst,path);
	return unlink(dst);
}

static int cmfs_chmod(const char *path, mode_t mode) {
	char dst[PATH_MAX];
	get_realname(dst,path);
	return chmod(dst,mode);
}

static int cmfs_mknod(const char *path, mode_t mode, dev_t rdev) {
	char dst[PATH_MAX];
	get_realname(dst,path);
	return mknod(dst,mode,rdev);
}

static int cmfs_truncate(const char *path, off_t size) {
	char dst[PATH_MAX];
	get_realname(dst,path);
	if(size==0)
		return truncate(dst,0);
	return -1;
}


static int cmfs_chown(const char *path, uid_t uid, gid_t gid) {
	char dst[PATH_MAX];
	get_realname(dst,path);
	return chown(dst,uid,gid);
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
	return mkdir(dst,mode);
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
		return ret;
	}
	if(ret>size)
		memcpy(buf,link,ret);
	else
		strcpy(buf,link);

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

	fd=open(dst,fi->flags|O_RDONLY);
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
	end=offset+size;
	if(end>st.st_size)
		end=st.st_size;
	endblk=end/FILEBLOCK;
	if(end%FILEBLOCK==0)
		endblk--;

	// 1. process first block(start with offset%FILEBLOCK) and check if is the last block.
	// 2. process mid blocks which are all full blocks
	// 3. process last block 

	if(startblk==endblk)
	{
		if((rd=readblk(fi->fh,startblk,plbuf,1))<=0)
			return rd;
		if(end>rd)
			end=rd;
		totalrd=end-offset;
		memcpy(buf+offset%FILEBLOCK,plbuf,totalrd);
		return totalrd;
		
	}
	if((rd=readblk(fi->fh,startblk,plbuf,0))<=0)
	{
		assert("readblk error");
		return rd;
	}

	// read full block at first
	totalrd+=(FILEBLOCK-offset%FILEBLOCK);
	memcpy(buf+offset%FILEBLOCK,plbuf,totalrd);

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
		int lastbytes=end%FILEBLOCK;
		if(lastbytes==0)
			lastbytes=FILEBLOCK;
		if(rd<lastbytes)	
			lastbytes=rd;
		memcpy(buf+totalrd,plbuf,lastbytes);
		totalrd+=lastbytes;
	}
	return totalrd;
}

static int cmfs_write(const char *path, const char *buf, size_t size, off_t offset,struct fuse_file_info *fi)
{
	int ret,i;
	off_t iblk,totalwr=0,end;
	off_t lastfileblk;
	int firstbyte,endbyte;// byte offset in block
	off_t startblk,endblk;// start from 0(consider as skip blocks)
	char cibuf[FILEBLOCK],plbuf[FILEBLOCK]={0};
	struct stat st; 
	int startcpy,endcpy; // start and end byte in memcpy between return buf --"buf" and decrypted block --"plbuf"

	int rd; // real read bytes 
	int de; // decrypted plaintext length (in 1k block) 
	
	if(!O_WRITE(fi->flags) ) {
		return -EACCES;
	}

	//ret=pwrite(fi->fh,buf,size,offset);


	if(fstat(fi->fh,&st)<0)
		return -EACCES;
	if(size<=0)
		return 0;

	//  when current blk>lastfileblk => needrd==0
	lastfileblk=st.st_size/FILEBLOCK;	
	if(st.st_size>0 && st.st_size%FILEBLOCK==0)
		lastfileblk--;
	startblk=offset/FILEBLOCK;
	end=offset+size;
	endblk=end/FILEBLOCK; 

	firstbyte=offset%FILEBLOCK;
	endbyte=end%FILEBLOCK;
	if(endbyte==0)
		endblk--;

	/* brief write pseudo procedure:
	 	process_first_blk:
			if firstblk>lastfileblk{ // needn't read
				fillzero(buf); // fill zero before startbyte
				if firstblk==endblk{
					writeblk(buf,needpad);
					return;
				}else{
					writeblk(buf,nopad);
				}
			}else{
		   		if firstblk==lastfileblk
					readblk(buf,needpad)
				else 
					readblk(buf,nopad)
				updatedata(buf);
				if firstblk==endblk
					writeblk(buf,needpad);
				else
					writeblk(buf,nopad);
			}


		process_mid_blks:
			writeblk(buf,nopad);


		process_last_blk:
			if lastblk>lastfileblk{ // needn't read
				writeblk(buf,needpad);
			}
			else{
	   			if lastblk==lastfileblk
					readblk(buf,needpad)
				else 
					readblk(buf,nopad)
				updatedata(buf);
				writeblk(buf,needpad);
			}

	*/

// static int writeblk(int fd, off_t blk, const char *buf, int startbyte, int slen /* plaintext length*/,int needpad); // buf should start from beginning of a block,but not byte start to be encrypted

	if(startblk==endblk){ // only one block,use "size" in memcpy
		if(startblk>lastfileblk) // need not readblk
		{
			// memset(plbuf,0,FILEBLOCK);
			memcpy(plbuf+firstbyte,buf,size);
			totalwr+=writeblk(fi->fh,startblk,plbuf,firstbyte,size,1);
		}else if(startblk==lastfileblk){
			if((rd=readblk(fi->fh,startblk,plbuf,1))<=0)
				memset(plbuf,0,FILEBLOCK);
			memcpy(plbuf+firstbyte,buf,size);
			if(rd>=endbyte) // left bytes need to be reencrypted, but do not need repadding
				totalwr+=writeblk(fi->fh,startblk,plbuf,firstbyte,rd-firstbyte,0);
			else // overwrite the end , need repadding
				totalwr+=writeblk(fi->fh,startblk,plbuf,firstbyte,endbyte-firstbyte,1);
		}else{// not last fileblock
			if((rd=readblk(fi->fh,startblk,plbuf,1))<=0) 
				memset(plbuf,0,FILEBLOCK);
			memcpy(plbuf+firstbyte,buf,size);
			totalwr+=writeblk(fi->fh,startblk,plbuf,firstbyte,FILEBLOCK-firstbyte,0);
		}
	}
	/*
	else{ // not last block,memcpy from startbyte to BlockEnd
		if(startblk>lastfileblk)
		{
			memcpy(plbuf+firstbyte,buf,FILEBLOCK-firstbyte);
			writeblk(fi->fh,startblk,plbuf,FILEBLOCK,0);
		}else if(startblk==lastfileblk){
			if((rd=readblk(fi->fh,startblk,plbuf,1))<0) return -1;
			memcpy(plbuf+firstbyte,buf,FILEBLOCK-firstbyte);
			writeblk(fi->fh,startblk,plbuf,
		}
	}

*/

	///////////
	/*
	if(startblk==endblk)
	{
		if((rd=readblk(fi->fh,startblk,plbuf,1))<=0)
			return rd;
		if(end>rd)
			end=rd;
		totalrd=end-offset;
		memcpy(buf+offset%FILEBLOCK,plbuf,totalrd);
		return totalrd;
		
	}
	if((rd=readblk(fi->fh,startblk,plbuf,0))<=0)
	{
		assert("readblk error");
		return rd;
	}

	// read full block at first
	totalrd+=(FILEBLOCK-offset%FILEBLOCK);
	memcpy(buf+offset%FILEBLOCK,plbuf,totalrd);

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
		int lastbytes=end%FILEBLOCK;
		if(lastbytes==0)
			lastbytes=FILEBLOCK;
		if(rd<lastbytes)	
			lastbytes=rd;
		memcpy(buf+totalrd,plbuf,lastbytes);
		totalrd+=lastbytes;
	}*/

	return size;
}

static int cmfs_create(const char *path, mode_t mode, struct fuse_file_info *fi)
{
    int retstat = 0;
    char fpath[PATH_MAX];
    int fd;

	sprintf(fpath,"%s%s",g_opts.src_dir,path);
	
    fd = creat(fpath, mode);
    if (fd < 0)
        retstat = -errno;

    fi->fh = fd;

    return retstat;
}

static struct fuse_operations cmfs_oper = {
  .getattr      = cmfs_getattr,
  .readlink     = cmfs_readlink,
  .readdir      = cmfs_readdir,
  .open         = cmfs_open,
  .read         = cmfs_read,
  .write		= cmfs_write,
  .create		= cmfs_create,
  .unlink       = cmfs_unlink,
  .chmod        = cmfs_chmod,
  .chown        = cmfs_chown,
  .mkdir        = cmfs_mkdir,
  .mknod        = cmfs_mknod,
  .truncate     = cmfs_truncate,
 /*
  .symlink      = cmfs_symlink,
  .rmdir        = cmfs_rmdir,
  .rename       = cmfs_rename,
  .link         = cmfs_link,
  .utimens      = cmfs_utimens,
  .release      = cmfs_release*/
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

