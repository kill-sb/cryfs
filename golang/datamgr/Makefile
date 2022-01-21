export GOPATH=$(PWD)

all:datamgr dtdfs server

datamgr:src/datamgr/*.go src/dbop/*.go src/coredata/*.go
datamgr:
	go build -gcflags "-N" datamgr

dtdfs:src/dtdfs/*.go
	go build dtdfs

server:src/server/*.go
	go build server
clean:
	rm -f datamgr dtdfs server

