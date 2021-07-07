
CC=gcc
CFLAGS=-g -D_FILE_OFFSET_BITS=64 -O2 -DFILEBLOCK=4096
#CFLAGS=-g -D_FILE_OFFSET_BITS=64 -D__DEBUG -O0
LDFLAGS=-lfuse -lssl -lcrypto 

OBJ=cmfs.o crypt.o
#OBJ=cmfs.o node.o dir.o

%.o: %.c
	$(CC) -c $(CFLAGS) $< -o $@

cmfs: $(OBJ)
	$(CC) $(OBJ) $(LDFLAGS) -o cmfs

test:cmfs
	./cmfs  /root/tmp /root/plain -o kernel_cache -o auto_cache -o big_writes -o max_write=65536 -o entry_timeout=120 -o attr_timeout=120  -o kernel_cache -o auto_cache
.PHONY: clean
clean:
	rm -f *.o cmfs

