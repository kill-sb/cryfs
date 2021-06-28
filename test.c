#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <fcntl.h>
#include <sys/stat.h>

int main(int c , char **v)
{
	int fd;
	size_t cnt,start;
	char *p;
	char ch;
	if(c<2)
		return 1;
	fd=open(v[1],O_RDWR);
	if(fd<0) {
		perror("fail:");
		return 1;
	}
	printf("ch:");
	ch=getchar();
	printf("Start,cnt:\n");
	scanf("%ld%ld",&start,&cnt);
	p=(char*)malloc(cnt);
	memset(p,ch,cnt);

	printf("size:%d-offset:%d\n",cnt,start);
	pwrite(fd,p,cnt,start);
	close(fd);
	return 0;
}
