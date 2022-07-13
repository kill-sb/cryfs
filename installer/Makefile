installer:install.c file.tgz
	gcc install.c -O2 -o installer
	dd if=file.tgz of=installer seek=65536 oflag=seek_bytes

clean:
	rm -f installer
