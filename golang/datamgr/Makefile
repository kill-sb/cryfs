export GOPATH=$(PWD)

all:datamgr dtdfs apisvr 

datamgr:src/datamgr/*.go src/dbop/*.go src/coredata/*.go src/apiv1/*.go
datamgr:
	go build -gcflags "-N" datamgr

dtdfs:src/dtdfs/*.go src/coredata/*.go
	go build dtdfs

apisvr:src/server/*.go src/apiv1/*.go src/dbop/*.go src/coredata/*.go
	go build -o apisvr server 
clean:
	rm -f datamgr dtdfs apisvr

