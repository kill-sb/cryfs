export GOPATH=$(PWD)


all:datamgr dtdfs apisvr 


dtdfs:src/dtdfs/*.go src/coredata/*.go
	go build dtdfs

datamgr:dtdfs src/datamgr/*.go src/dbop/*.go src/coredata/*.go src/apiv1/*.go
datamgr:
	strip dtdfs
	go build -gcflags "-N" -ldflags "-X 'main.CheckSum=`sha256sum dtdfs`'" datamgr

apisvr:src/server/*.go src/apiv1/*.go src/dbop/*.go src/coredata/*.go
	go build -o apisvr server 
clean:
	rm -f datamgr dtdfs apisvr

