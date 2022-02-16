package main

import (
	_"dbop"
	"net/http"
	//"os/exec"
	"encoding/json"
	"coredata"
	"log"
//	"fmt"
	"flag"
	"sync"
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
	Log *log.Logger
	Debug	bool
}

var configfile string
var g_config ServerConfig

var fhandler http.Handler
var tokenmap map[string] *LoginUserInfo
var tokenlock sync.RWMutex

func LoadSvrConfig() *ServerConfig{
//	g_config= new (ServerConfig)
	curpath:=coredata.GetSelfPath()
	g_config.Port=":8080"
	g_config.CertPem=curpath+"/cert.pem"
	g_config.KeyPem=curpath+"/key.pem"
	g_config.Version="v1"
	g_config.LoginMethod=PASSWD

	file,err:=os.Open(configfile)
	if err==nil{
		/*
		config file format sameple:
		[server]
		version=v1
		port=8080
		cert=/etc/dtdfs_cert.pem
		key=/etc/dtdfs_key.pem
		login=PASSWD,CERT,MOBILE 
		*/
		file.Close()
	}

	return &g_config
}

func defhandler(w http.ResponseWriter, r *http.Request){
	if r.Method=="GET"{
	//			fhandler.ServeHTTP(w,r)

	}
}

func Debug(obj...interface{}){
	if g_config.Log!=nil{
		g_config.Log.Println(obj...)
	}else{
		log.Println(obj...)
	}
}

func DebugJson(tip string,obj interface{}){
	ret,err:=json.Marshal(obj)
	if err==nil{
		if g_config.Log!=nil{
			g_config.Log.Println(tip,string(ret))
		}else{
			log.Println(tip,string(ret))
		}
	}
}

func SetupHandler() error{
//	fhandler=http.FileServer(http.Dir("/root/linux-5.9.1"))
	http.HandleFunc("/", defhandler)
	prefix:="/api/"+g_config.Version+"/"
	http.HandleFunc(prefix+"login",LoginFunc) // POST
	http.HandleFunc(prefix+"getuser",GetUserFunc) // GET
	http.HandleFunc(prefix+"findusername",FindUserNameFunc) // GET
	http.HandleFunc(prefix+"newdata",NewDataFunc) // POST
	http.HandleFunc(prefix+"getshareinfo",GetShareInfoFunc) // GET
	http.HandleFunc(prefix+"sharedata",ShareDataFunc) // POST
	http.HandleFunc(prefix+"getdatainfo",GetDataInfoFunc) // GET
	http.HandleFunc(prefix+"traceback",TraceBackFunc) // GET
	http.HandleFunc(prefix+"updatedata",UpdateDataFunc) // GET
	http.HandleFunc(prefix+"traceforward",TraceForwardFunc) // GET
	http.HandleFunc(prefix+"queryobjs",QueryObjsFunc) // GET
	http.HandleFunc(prefix+"logout",LogoutFunc) // GET
	http.HandleFunc(prefix+"refreshtoken",RefreshTokenFunc) // GET

	return nil
}

func ParseArgs(){
	var logfile string
	flag.BoolVar(&g_config.Debug,"d",false,"run on debug mode")
	flag.StringVar(&configfile,"config","apisvr.cfg","name of configure file for server setup")
	flag.StringVar(&logfile,"logfile","","log file name")
	flag.Parse()
	g_config.Log=nil
//	file,err:=os.Create(logfile)
	file,err:=os.OpenFile(logfile,os.O_RDWR|os.O_APPEND|os.O_CREATE|os.O_SYNC,0644)
	if err==nil{
		g_config.Log=log.New(file,"",log.LstdFlags|log.Lshortfile)
		g_config.Log.Println("Start:")
	}
}

func main(){
	ParseArgs()
	/*if !g_config.Debug && os.Getppid() != 1{
        cmd := exec.Command(os.Args[0], os.Args[1:]...)
        cmd.Start()
        os.Exit(0)
    }*/
	LoadSvrConfig()
	err:=SetupHandler()
	if err!=nil{
		Debug("Setup handler error:",err)
		return
	}
	tokenmap=make(map[string]*LoginUserInfo)
	go TokenCacheMgr()
	err=http.ListenAndServeTLS(g_config.Port,g_config.CertPem,g_config.KeyPem,nil)
	if err!=nil{
		Debug("Listen error:",err)
	}
}
