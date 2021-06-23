
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
	if(len>4 && strcmp(buf+len-4,".cmc")==0)
	{
		buf[len-4]='\0';
		ret=1;
	}
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

static int initstat(struct node *node, mode_t mode) {
  struct stat *stbuf = &node->vstat;
  memset(stbuf, 0, sizeof(struct stat));
  stbuf->st_mode  = mode;
  stbuf->st_nlink = 0;
  update_times(node, U_ATIME | U_MTIME | U_CTIME);
  return 1;
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


//
// Filesystem entry points
//

static int memfs_getattr(const char *path, struct stat *stbuf) {
  struct node *node;
  if(!getnodebypath(path, &the_fs, &node)) {
    return -errno;
  }

  stbuf->st_mode   = node->vstat.st_mode;
  stbuf->st_nlink  = node->vstat.st_nlink;
  stbuf->st_size   = node->vstat.st_size;
  stbuf->st_blocks = node->vstat.st_blocks;
  stbuf->st_uid    = node->vstat.st_uid;
  stbuf->st_gid    = node->vstat.st_gid;
  stbuf->st_mtime  = node->vstat.st_mtime;
  stbuf->st_atime  = node->vstat.st_atime;
  stbuf->st_ctime  = node->vstat.st_ctime;

  // Directories contain the implicit hardlink '.'
  if(S_ISDIR(node->vstat.st_mode)) {
    stbuf->st_nlink++;
  }

  return 0;
}

static int memfs_readlink(const char *path, char *buf, size_t size) {
  struct node *node;
  if(!getnodebypath(path, &the_fs, &node)) {
    return -errno;
  }

  if(!S_ISLNK(node->vstat.st_mode)) {
    return -ENOLINK;
  }

  // Fuse breaks compatibility with other readlink() implementations as we cannot use the return
  // value to indicate how many bytes were written. Instead, we need to null-terminate the string,
  // unless the buffer is not large enough to hold the path. In that case, fuse will null-terminate
  // the string before passing it on.

  if(node->vstat.st_size > size) {
    memcpy(buf, node->data, size);
  } else {
    strcpy(buf, node->data);
  }

  return 0;
}

static int memfs_readdir(const char *path, void *buf, fuse_fill_dir_t filler, off_t offset, struct fuse_file_info *fi) {
  struct node *dir;
  if(!getnodebypath(path, &the_fs, &dir)) {
    return -errno;
  }

  if(!S_ISDIR(dir->vstat.st_mode)) {
    return -ENOTDIR;
  }

  filler(buf, ".",  &dir->vstat, 0);
  if(dir == the_fs.root) {
    filler(buf, "..", NULL, 0);
  } else {
    char *parent_path = safe_dirname(path);
    struct node *parent;
    getnodebypath(parent_path, &the_fs, &parent);
    free(parent_path);
    filler(buf, "..", &parent->vstat, 0);
  }

  struct direntry *entry = dir->data;
  while(entry != NULL) {
    if(filler(buf, entry->name, &entry->node->vstat, 0))
      break;
    entry = entry->next;
  }

  return 0;
}

static int memfs_mknod(const char *path, mode_t mode, dev_t rdev) {
  struct node *node;
  int res = createentry(path, mode, &node);
  if(res) return res;

  if(S_ISREG(mode)) {
    node->data = NULL;
    node->vstat.st_blocks = 0;
  } else {
    return -ENOSYS;
  }

  return 0;
}

static int memfs_mkdir(const char *path, mode_t mode) {
  struct node *node;
  int res = createentry(path, S_IFDIR | mode, &node);
  if(res) return res;

  // No entries
  node->data = NULL;

  return 0;
}

static int memfs_unlink(const char *path) {
  char *dirpath, *name;
  struct node *dir, *node;

  // Find inode
  if(!getnodebypath(path, &the_fs, &node)) {
    return -errno;
  }

  if(S_ISDIR(node->vstat.st_mode)) {
    return -EISDIR;
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

  // If the link count is zero, delete the associated data
  if(node->vstat.st_nlink == 0) {
    if(node->fd_count == 0) {
      // No open file descriptors, we can safely delete the node
      if(node->data) free(node->data);
      free(node);
    } else {
      // There are open file descriptors, schedule deletion
      node->delete_on_close = 1;
    }
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

static int memfs_chmod(const char *path, mode_t mode) {
  struct node *node;
  if(!getnodebypath(path, &the_fs, &node)) {
    return -errno;
  }

  mode_t mask = S_ISUID | S_ISGID | S_ISVTX |
                S_IRUSR | S_IWUSR | S_IXUSR |
                S_IRGRP | S_IWGRP | S_IXGRP |
                S_IROTH | S_IWOTH | S_IXOTH;

  node->vstat.st_mode = (node->vstat.st_mode & ~mask) | (mode & mask);

  update_times(node, U_CTIME);

  return 0;
}

static int memfs_chown(const char *path, uid_t uid, gid_t gid) {
  struct node *node;
  if(!getnodebypath(path, &the_fs, &node)) {
    return -errno;
  }

  node->vstat.st_uid = uid;
  node->vstat.st_gid = gid;

  update_times(node, U_CTIME);

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

static int memfs_truncate(const char *path, off_t size) {
  struct node *node;
  if(!getnodebypath(path, &the_fs, &node)) {
    return -errno;
  }

  // Calculate new block count
  blkcnt_t newblkcnt = (size + BLOCKSIZE - 1) / BLOCKSIZE;
  blkcnt_t oldblkcnt = node->vstat.st_blocks;

  if(oldblkcnt < newblkcnt) {
    // Allocate additional memory
    void *newdata = malloc(newblkcnt * BLOCKSIZE);
    if(!newdata) {
      return -ENOMEM;
    }

    memcpy(newdata, node->data, node->vstat.st_size);
    free(node->data);
    node->data = newdata;
  } else if(oldblkcnt > newblkcnt) {
    // Allocate new memory so we can free the unnecessarily large memory
    void *newdata = malloc(newblkcnt * BLOCKSIZE);
    if(!newdata) {
      return -ENOMEM;
    }

    memcpy(newdata, node->data, size);
    free(node->data);
    node->data = newdata;
  }

  // Fill additional memory with zeroes
  if(node->vstat.st_size < size) {
    memset(node->data + node->vstat.st_size, 0, node->vstat.st_size - size);
  }

  // Update file size
  node->vstat.st_size = size;
  node->vstat.st_blocks = newblkcnt;

  return 0;
}

static int memfs_write(const char *path, const char *buf, size_t size, off_t offset, struct fuse_file_info *fi) {
  struct filehandle *fh = (struct filehandle *) fi->fh;

  // Check whether the file was opened for writing
  if(!O_WRITE(fh->o_flags)) {
    return -EACCES;
  }

  struct node *node = fh->node;

  // Calculate number of required blocks
  blkcnt_t req_blocks = (offset + size + BLOCKSIZE - 1) / BLOCKSIZE;

  if(node->vstat.st_blocks < req_blocks) {
    // Allocate more memory
    void *newdata = malloc(req_blocks * BLOCKSIZE);
    if(!newdata) {
      return -ENOMEM;
    }

    // Copy old contents
    if(node->data != NULL) {
      memcpy(newdata, node->data, node->vstat.st_size);
      free(node->data);
    }

    // Update allocation information
    node->data = newdata;
    node->vstat.st_blocks = req_blocks;
  }

  // Write to file buffer
  memcpy(((char *) node->data) + offset, buf, size);

  // Update file size if necessary
  off_t minsize = offset + size;
  if(minsize > node->vstat.st_size) {
    node->vstat.st_size = minsize;
  }

  update_times(node, U_CTIME | U_MTIME);

  return size;
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


static void update_attr(const char* path, struct stat *stbuf){
}


static int cmfs_getattr(const char *path, struct stat *stbuf) {
  	int ret;
	char dst[PATH_MAX];
	get_realname(dst,path);
  	if((ret=lstat(dst,stbuf))){
	  	return -errno;
  	}
	update_attr(dst,stbuf);
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
		if((drt->d_type & DT_DIR) ==0)// nondir
			sprintf(dstname,"%s.cmc",drt->d_name);
		else 
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

	fd=open(dst,fi->flags);
	fi->fh = fd;
	return 0;
}

static int cmfs_read(const char *path, char *buf, size_t size, off_t offset, struct fuse_file_info *fi) {
	int ret,i;
	off_t iblk,totalrd=0;
	off_t startblk,endblk;// skip blocks
	char cibuf[FILEBLOCK],plbuf[FILEBLOCK];
	struct stat st; 
	off_t end;// end of file
	int rd,de;// real read bytes and decrypted plaintext length in each block 

	// Check whether the file was opened for reading
	if(!O_READ(fi->flags)) {
		return -EACCES;
	}

/*	ret=pread(fi->fh,buf,size,offset);
	for(i=0;i<ret;i++){
		if(buf[i]>='a' && buf[i]<='z')
			buf[i]-='a'-'A';
	}*/
	if(fstat(fi->fh,&st)<0)
		return -EACCES;
	if(size==0 || offset>st.st_size)
		return 0;

	end=offset+size>st.st_size?st.st_size:offset+size;
	startblk=offset/FILEBLOCK;
	endblk=end/FILEBLOCK;
	for(iblk=startblk;iblk<endblk;iblk++){
		rd=pread(fi->fh,cibuf,FILEBLOCK,i*FILEBLOCK);
		assert(rd==FILEBLOCK);// we are here means it must not be endblk
		de=decodeblk(cibuf,g_opts.keyinfo.crypt_key,plbuf,rd,0);
		assert(de==FILEBLOCK);
		if(iblk==startblk){
			int startbyte=offset%FILEBLOCK;
			memcpy(buf,plbuf+startbyte,FILEBLOCK-startbyte);
			totalrd+=(FILEBLOCK-startbyte);
		}else{
			memcpy(buf+totalrd,plbuf,FILEBLOCK);
			totalrd+=FILEBLOCK;
		}
	}
	// last block,but also may be first block
	rd=pread(fi->fh,cibuf,FILEBLOCK,endblk*FILEBLOCK);
	assert(rd==end%FILEBLOCK);
	de=decodeblk(cibuf,g_opts.keyinfo.crypt_key,plbuf,rd,1);
	if(de<=0) return 0;
	if(startblk==endblk){
		int startbyte=offset%FILEBLOCK;
		memcpy(buf,plbuf+startbyte,de-startbyte);
		totalrd+=de-startbyte;
	}
	return totalrd;
}

static int cmfs_write(const char *path, const char *buf, size_t size, off_t offset,struct fuse_file_info *fi)
{
	char *tmp;
	int ret,i;
	if(!O_WRITE(fi->flags)) {
		return -EACCES;
	}
	tmp=(char*)malloc(size);
	for(i=0;i<size;i++)
	{
		if(buf[i]>='A' && buf[i]<='Z')
			tmp[i]=buf[i]+'a'-'A';
		else
			tmp[i]=buf[i];
	}
	ret=pwrite(fi->fh,tmp,size,offset);
	free(tmp);
	return ret;
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
//  .mknod        = cmfs_mknod,
//  .mkdir        = cmfs_mkdir,
 /*
  .symlink      = cmfs_symlink,
  .unlink       = cmfs_unlink,
  .rmdir        = cmfs_rmdir,
  .rename       = cmfs_rename,
  .link         = cmfs_link,
  .chmod        = cmfs_chmod,
  .chown        = cmfs_chown,
  .truncate     = cmfs_truncate,
  .utimens      = cmfs_utimens,
  .write        = cmfs_write,
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

