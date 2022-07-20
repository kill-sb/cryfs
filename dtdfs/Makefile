export GOPATH=$(PWD)


all:subdir setenv datamgr dtdfs apisvr

subdir:
	make -C fs

dtdfs:src/dtdfs/*.go src/coredata/*.go 
	go build dtdfs

datamgr:dtdfs src/datamgr/*.go src/dbop/*.go src/coredata/*.go src/apiv1/*.go
datamgr:
	strip dtdfs
	go build -gcflags "-N" -ldflags "-X 'main.CheckSum=`sha256sum dtdfs`'" datamgr

apisvr:src/server/*.go src/apiv1/*.go src/dbop/*.go src/coredata/*.go 
	go build -o apisvr server 

testsvr:apisvr
#	ulimit -n 65535
	./apisvr -d

setenv:
	go env -w GO111MODULE=off

clean:
	cd fs && rm -f cmfs *.o
	rm -f datamgr dtdfs apisvr

