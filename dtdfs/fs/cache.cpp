#include "cache.h"
#include <iostream>

using namespace std;

extern "C"
void cache_sync(int fd)
{
	cout<<"test cpp"<<endl;	
}


