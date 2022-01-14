package main

import (
	_"dbop"
	"net/http"
	"fmt"
	"flag"
	"os"
)


const(
	PASSWD=1
	CERT=1<<1
	MOBILE=1<<2
)

type ServerConfig struct{
	Port string
	CertPem string
	KeyPem string
	Version string
	LoginMethod	int
}

var configfile string
var g_config *ServerConfig

func LoadSvrConfig() *ServerConfig{
	g_config= new (ServerConfig)
	g_config.Port=":8080"
	g_config.CertPem="cert.pem"
	g_config.KeyPem="key.pem"
	g_config.Version="v1"
	g_config.LoginMethod=PASSWD

	flag.StringVar(&configfile,"config","server.cfg","name of configure file for server setup")
	file,err:=os.Open(configfile)
	if err==nil{
		defer file.Close()
		/*
		config file format sameple:
		[server]
		version=v1
		port=8080
		cert=/etc/dtdfs_cert.pem
		key=/etc/dtdfs_key.pem
		login=PASSWD,CERT,MOBILE 
		*/
	}
	return g_config
}


func SetupHandler(cfg *ServerConfig) error{
//	http.HandleFunc("/", defhandler)
	prefix:="/api/"+cfg.Version+"/"
	http.HandleFunc(prefix+"login",LoginFunc) // POST
	http.HandleFunc(prefix+"getuser",GetUserFunc) // GET
	http.HandleFunc(prefix+"newdata",NewDataFunc) // POST
	http.HandleFunc(prefix+"sharedata",ShareDataFunc) // POST
	http.HandleFunc(prefix+"getdatainfo",GetDataInfoFunc) // GET
	http.HandleFunc(prefix+"traceback",TraceBackFunc) // GET
	http.HandleFunc(prefix+"traceforward",TraceForwardFunc) // GET

	return nil
}

func main(){
	LoadSvrConfig()
	err:=SetupHandler(g_config)
	if err!=nil{
		fmt.Println("Setup handler error:",err)
		return
	}
	err=http.ListenAndServeTLS(g_config.Port,g_config.CertPem,g_config.KeyPem,nil)
	if err!=nil{
		fmt.Println("Listen error:",err)
	}
}