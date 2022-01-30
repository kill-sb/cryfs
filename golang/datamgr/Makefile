export GOPATH=$(PWD)

all:datamgr dtdfs server

datamgr:src/datamgr/*.go src/dbop/*.go src/coredata/*.go src/apiv1/*.go
datamgr:
	go build -gcflags "-N" datamgr

dtdfs:src/dtdfs/*.go
	go build dtdfs

server:src/server/*.go src/apiv1/*.go src/dbop/*.go
	go build server
clean:
	rm -f datamgr dtdfs server

