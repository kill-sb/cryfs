export GOPATH=$(PWD)

all:datamgr

#datamgr:src/datamgr/main.go src/dbop/dbop.go
datamgr:
	go build -gcflags "-N" datamgr
clean:
	rm -f datamgr 

