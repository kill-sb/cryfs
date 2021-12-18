export GOPATH=$(PWD)

all:datamgr ddfs

datamgr:src/datamgr/*.go src/dbop/*.go src/coredata/*.go
datamgr:
	go build -gcflags "-N" datamgr

ddfs:src/ddfs/*.go
	go build ddfs
clean:
	rm -f datamgr ddfs

