export GOPATH=$(PWD)

all:datamgr

datamgr:src/datamgr/*.go src/dbop/*.go src/coredata/*.go
datamgr:
	go build -gcflags "-N" datamgr
clean:
	rm -f datamgr 

