
CC=gcc
CFLAGS=-g -Wall -D_FILE_OFFSET_BITS=64
LDFLAGS=-lfuse

OBJ=cmfs.o 
#OBJ=cmfs.o node.o dir.o

%.o: %.c
	$(CC) -c $(CFLAGS) $< -o $@

cmfs: $(OBJ)
	$(CC) $(OBJ) $(LDFLAGS) -o cmfs

.PHONY: clean
clean:
	rm -f *.o cmfs

