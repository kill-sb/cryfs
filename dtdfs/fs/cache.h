#ifndef __CACHE_H__
#define __CACHE_H__
#include <stdlib.h>
#include <map>
#include <time.h>
#include "cmfs.h"

using namespace std;
#define MEM_CACHE_MAX 1024*1024*1024

#define MAX_BLOCKS MEM_CACHE_MAX/FILE_BLOCK

#ifdef __cplusplus
extern "C"{
#endif
const char* cache_getblock(int fd, off_t blk);
const char* cache_writeblock(int fd, off_t blk, const char* buf);
void cache_sync(int fd);

#ifdef __cplusplus
}
#endif

struct cache_buf{
//	const char* path; //global file map key
//	off_t blk;  // in-file block map key
	const char* fname;
	off_t blk;
	time_t visit;
	char buf[FILEBLOCK];
	int len;
	cache_buf()
	{
//		visit=time(NULL);
		fname=NULL;
	}
};

struct file_cache{
	char *fname;
	map<off_t,cache_buf*> pages;
	file_cache()
	{
		fname=NULL;
	}
	~file_cache()
	{
		if(fname)
			delete []fname;
	}
};

#endif
