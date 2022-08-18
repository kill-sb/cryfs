#include <map>
#include <sys/time.h>
#include <string>
#include <unistd.h>
#include <stdlib.h>
#include <stdio.h>
#include <iostream>
#include <string.h>

using namespace std;
#include "cache.h"
typedef unsigned long ulong;

struct cache_buf{
	const char* fname;
	off_t blk;
	ulong visit;
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



map<ulong,cache_buf*>lru_map;
map<string, file_cache*> g_cache;

unsigned long memuse=0;

static ulong getusec()
{	
	struct timeval now;
	gettimeofday(&now,NULL);
	return now.tv_sec*1000*1000+now.tv_usec;

}

static int updatevisit(cache_buf* cb)
{
	map<ulong,cache_buf*>::iterator pos=lru_map.find(cb->visit);
	if(pos==lru_map.end())
		return -1;
	lru_map.erase(pos);
	cb->visit=getusec();
	lru_map[cb->visit]=cb;
	return 0;
}


static const char* getpath(int fd,char *buf)// PATH_MAX len
{
	static char path[PATH_MAX];
	sprintf(path,"/proc/%d/fd/%d",getpid(),fd);
	int ret=readlink(path,buf,PATH_MAX);
	if (ret>0)
		return buf;
	return NULL;
}

// try to free memory with LRU algo
int cleanup()
{
	return -1;
}

extern "C"
int add_cache(int fd,off_t blk,const char* buf, int buf_len)
{
	if (memuse>=MEM_CACHE_MAX)
	{
		if (cleanup()<0)
			return -1;
	}
	char path[PATH_MAX];
	const char* fname=getpath(fd,path);
	cache_buf *cur=NULL;
	if (fname==NULL) return -1;
	file_cache *fc=g_cache[fname];
	if (fc==NULL)
	{
		fc=new file_cache;
		if (fc==NULL) return -1;	
		fc->fname=new char[strlen(fname)+1];
		strcpy(fc->fname,fname);
		g_cache[fname]=fc;
	}else{
		cur=fc->pages[blk];
	}
	
	if(cur==NULL)
	{
		cur=new cache_buf;
		if (cur==NULL)
		{
			delete fc;
			return -1;
		}
		cur->fname=fc->fname;
		cur->blk=blk;
		fc->pages[blk]=cur;
		memuse+=sizeof(cache_buf);
	}
	cur->len=buf_len;
	memcpy(cur->buf,buf,buf_len);
	updatevisit(cur);
	return 0;
}	

extern "C"
int searchcache(int fd, off_t blk,char *buf,int *len)
{
	char path[PATH_MAX];
	const char* fname=getpath(fd,path);
	if (fname==NULL) return -1;
	file_cache *fc=g_cache[fname];
	if (fc==NULL)
		return 0;
	cache_buf *get=fc->pages[blk];
	if(get==NULL)
		return 0;
	*len=get->len;
	if(get->len<=0) // error
		return get->len;
	memcpy(buf,get->buf,*len);
	updatevisit(get);
	return *len; // return >0, buf has been copied
}

extern "C"
void cache_sync(int fd)
{
	// invoked after close,fsync,
//	cout<<"test cpp"<<endl;	
}

extern "C"
void drop_cache(const char* fname,off_t block)
{
	// invoked after unlink, cleanup
//	file_cache* fc=			
}


