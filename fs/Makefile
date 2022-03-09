
CC=gcc
CFLAGS=-D_FILE_OFFSET_BITS=64 -O2
#CFLAGS=-g -D_FILE_OFFSET_BITS=64 -D__DEBUG -O0
LDFLAGS=-lfuse -lssl -lcrypto 

OBJ=cmfs.o crypt.o
#OBJ=cmfs.o node.o dir.o

%.o: %.c
	$(CC) -c $(CFLAGS) $< -o $@

cmfs: $(OBJ)
	$(CC) $(OBJ) $(LDFLAGS) -o cmfs

demo:cmfs

	./cmfs /root/encrypt /root/plain -o big_writes -o max_write=65536 -o entry_timeout=120 -o attr_timeout=120  -o kernel_cache -o auto_cache

install:cmfs
	/bin/cp cmfs /usr/local/bin
test:cmfs
#	./cmfs  /tmp/mnt  /mnt -o big_writes -o max_write=65536 -o entry_timeout=120 -o attr_timeout=120
	./cmfs  /root/mnt  /mnt -o big_writes -o max_write=65536 -o entry_timeout=120 -o attr_timeout=120  -o kernel_cache -o auto_cache
#	./cmfs  /tmp/mnt  /mnt  -o kernel_cache -o auto_cache
.PHONY: clean
clean:
	rm -f *.o cmfs

