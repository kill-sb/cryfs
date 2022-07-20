export GOPATH=$(PWD)


all:subdir setenv datamgr dtdfs apisvr

subdir:
	make -C fs

cmfs:fs/cmfs
	ln fs/cmfs cmfs

dtdfs:src/dtdfs/*.go src/coredata/*.go 
	go build dtdfs

datamgr:dtdfs cmfs src/datamgr/*.go src/dbop/*.go src/coredata/*.go src/apiv1/*.go
datamgr:
	strip dtdfs
	go build -gcflags "-N" -ldflags "-X 'main.DtdfsSum=`sha256sum dtdfs`' -X 'main.CmfsSum=`sha256sum fs/cmfs`'" datamgr

apisvr:src/server/*.go src/apiv1/*.go src/dbop/*.go src/coredata/*.go 
	go build -o apisvr server 

testsvr:apisvr
#	ulimit -n 65535
	./apisvr -d

setenv:
	go env -w GO111MODULE=off

clean:
	cd fs && rm -f cmfs *.o
	rm -f datamgr dtdfs apisvr cmfs

