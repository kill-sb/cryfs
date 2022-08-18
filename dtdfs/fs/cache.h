#ifndef __CACHE_H__
#define __CACHE_H__
#include <stdlib.h>
#include <time.h>
#include "cmfs.h"

#define MEM_CACHE_MAX 1024*1024*1024

#define MAX_BLOCKS MEM_CACHE_MAX/FILE_BLOCK

#ifdef __cplusplus
extern "C"{
#endif
const char* cache_getblock(int fd, off_t blk);
const char* cache_writeblock(int fd, off_t blk, const char* buf);
void cache_sync(int fd);
int addcache(int fd, off_t blk,const char* buf, int buf_len);
int searchcache(int fd, off_t blk, char* buf, int* len);
#ifdef __cplusplus
}
#endif

#endif
