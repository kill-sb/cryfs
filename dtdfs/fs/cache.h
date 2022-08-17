#ifndef __CACHE_H__
#define __CACHE_H__
#include <stdlib.h>

#include "cmfs.h"

#define MEM_CACHE_MAX 1024*1024*1024

#define MAX_BLOCKS MEM_CACHE_MAX/FILE_BLOCK

const char* cache_getblock(int fd, off_t blk);
const char* cache_writeblock(int fd, off_t blk, const char* buf);
#ifdef __cplusplus
extern "C"{
#endif
void cache_sync(int fd);
#ifdef __cplusplus
}
#endif

#endif
