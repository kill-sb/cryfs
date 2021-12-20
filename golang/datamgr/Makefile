export GOPATH=$(PWD)

all:datamgr dtdfs

datamgr:src/datamgr/*.go src/dbop/*.go src/coredata/*.go
datamgr:
	go build -gcflags "-N" datamgr

dtdfs:src/dtdfs/*.go
	go build dtdfs
clean:
	rm -f datamgr dtdfs

