
CC=gcc
CFLAGS=-g -D_FILE_OFFSET_BITS=64 -O2
#CFLAGS=-g -D_FILE_OFFSET_BITS=64 -D__DEBUG -O0
LDFLAGS=-lfuse -lssl -lcrypto 

OBJ=cmfs.o crypt.o
#OBJ=cmfs.o node.o dir.o

%.o: %.c
	$(CC) -c $(CFLAGS) $< -o $@

cmfs: $(OBJ)
	$(CC) $(OBJ) $(LDFLAGS) -o cmfs

test:cmfs
	./cmfs  /tmp/aes /mnt -o kernel_cache -o auto_cache

.PHONY: clean
clean:
	rm -f *.o cmfs

