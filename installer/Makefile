VERSION="\"$(shell .dtdfs_files/dtdfs -v)\""

installer:install.c file.tgz
	gcc install.c -D__CUR_VERSION=$(VERSION) -O2 -o installer
	dd if=file.tgz of=installer seek=65536 oflag=seek_bytes

file.tgz:.dtdfs_files/*
	tar czvf file.tgz .dtdfs_files/

clean:
	rm -f file.tgz installer
