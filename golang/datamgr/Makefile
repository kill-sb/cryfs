export GOPATH=$(PWD)

all:datamgr ddfce

datamgr:src/datamgr/*.go src/dbop/*.go src/coredata/*.go
datamgr:
	go build -gcflags "-N" datamgr

ddfce:src/ddfce/*.go
	go build ddfce
clean:
	rm -f datamgr 

