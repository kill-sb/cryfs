package main

import(
	"fmt"
	"syscall"
	"strings"
	"errors"
	"flag"
	"os"
	core "coredata"
)

const AES_KEY_LEN=128
const version="Data Defense 0.95"

type SpaceInStr struct{
	val string
}

func (d *SpaceInStr)String()string{
	return d.val
}

func (d *SpaceInStr)Set(str string) error{
	d.val=str
	return nil
}

var definpath , inpath string
var defoutpath,outpath , oname string
var mntimport string
var defuser, loginuser string
var datadesc string
var config string
var keyword string
var podimg string
var apisvr string
var DtdfsSum string
var CmfsSum string
var ouid,ogid int

var namemap map[string]int32
//var idmap map[int32]string

var uuidmap map[string]*core.EncryptedData
var useridmap map[int32] string

func ComUser(){
	syscall.Setuid(ouid)
	syscall.Setgid(ogid)
}

func LoadConfig(){
	definpath=os.Getenv("DATA_IN_PATH")
//	defoutpath=os.Getenv("HOME")+"/.cmitdata"
	defuser=os.Getenv("DATA_DEF_USER")
	datadesc=""
}

func GetFunction() int {
	//var bList,bEnc,bTrace,bShare,bMnt,bDec,bLogin bool
	var bList,bEnc,bTrace,bShare,bMnt,bDec,bVer bool
	flag.BoolVar(&bEnc,"enc",false,"encrypt raw data")
	flag.BoolVar(&bShare,"share",false,"share data to other users")
	flag.BoolVar(&bMnt,"mnt",false,"mount encrypted data")
	flag.BoolVar(&bDec,"dec",false,"decrypted local data(this function is for current test only, it will be removed in release edtion)")
	flag.BoolVar(&bTrace,"trace",false,"trace details of data")
	flag.BoolVar(&bList,"list",false,"list local encrypted data")
//	flag.BoolVar(&bLogin,"login",false,"login and get a token")
	flag.StringVar(&datadesc,"desc","","Data description")
	flag.StringVar(&inpath,"in",definpath,"original data path (may be a file or a directory)")
	flag.StringVar(&outpath,"out",defoutpath,"output data path")
	flag.StringVar(&oname,"oname","","output new data org-name(default named with uuid)")
	flag.StringVar(&podimg,"img","cmit","container base image")
	flag.StringVar(&mntimport,"import","", "import plain data dir into container")
	flag.StringVar(&apisvr,"apisvr","apisvr:8080", "api server address (ip:port)")
	flag.StringVar(&loginuser,"user",defuser, "login user name")
//	flag.StringVar(&config,"config","", "use config file to decribe share info")
	flag.StringVar(&keyword,"search","", "used with -list or -trace.(When used with -list,search data records contain the keyword only, and when used with -trace, highlight the keyword)")
	flag.BoolVar(&bVer,"v",false,"display version")
	var ougid uint64
	flag.Uint64Var(&ougid,"l",0,"log mode")
	flag.Parse()
	ouid=int(ougid&0xffffffff)
	ogid=int((ougid>>32)&0xffffffff)
	ret:=core.INVALID
	count:=0

/*	if bLogin{
		ret=core.LOGIN
		count++
	}*/
	if(bList){
		ComUser()
		ret=core.LIST
		count++
	}
	if(bTrace){
		ComUser()
		ret=core.TRACE
		count++
	}
	if(bDec){
		ComUser()
		ret=core.DECODE
		count++
	}
	if bEnc{
		ComUser()
		ret= core.ENCODE
		count++
	}
	if bShare{
		ComUser()
		ret=core.DISTRIBUTE
		count++
	}
	if bMnt{
		ret=core.MOUNT
		count++
	}
	if bVer{
		ComUser()
		ret=core.VERSION
		count++
	}
	if count!=1{
		ret=core.INVALID
	}
	return ret
}

func PathFromPid(pid int)(string,error){
	pidfile:=fmt.Sprintf("/proc/%d/exe",pid)
	str,err:=os.Readlink(pidfile)
	if err!=nil{
//		fmt.Println("Readlink error")
		return "",errors.New("'datamgr' is an inner module, use 'dtdfs' instead")
	}
	return str,nil
}

func CheckSums()error{
	ppid:=os.Getppid()
	fpath,err:=PathFromPid(ppid)
	if err!=nil{
		return nil
	}
	if strings.HasSuffix(fpath,"/bash") || strings.HasSuffix(fpath,"/dash"){
		stfile:=fmt.Sprintf("/proc/%d/stat",ppid)
		f,err:=os.Open(stfile)
		defer f.Close()
		var tmp string
		_,err=fmt.Fscanf(f,"%s%s%s%d",&tmp,&tmp,&tmp,&ppid)
		if err!=nil{
			return err
		}
		fpath,err=PathFromPid(ppid)
		if err!=nil{
			return err
		}
	}
	if strings.HasSuffix(fpath,"/dtdfs"){
		sum:=strings.Split(DtdfsSum," ")[0]
		result,err:=GetFileSha256(fpath)
		if err==nil && result==sum{
			return nil
		}else{
			return errors.New("Invalid dtdfs file")
		}
	}
	return errors.New("'datamgr' is an inner module, use 'dtdfs' instead")
}

func testlogin(){
	token,err:=doAuth(loginuser)
	if err!=nil {
		fmt.Println("Login error:",err)
	}else{
		if token.Code==0{
			fmt.Println("test login ok:",token,",data:",token.Data)
		}else{
			fmt.Println("login failed:",token.Msg)
		}
	}
}

func main(){
	if err:=CheckSums();err!=nil{
		fmt.Println(err)
		return
	}
	LoadConfig()
	namemap=make(map[string]int32)

	uuidmap =make(map[string]*core.EncryptedData)
	useridmap=make(map[int32]string)
	fun:=GetFunction()
	if inpath!=""{
		inpath=strings.TrimSuffix(inpath,"/")
	}
	if mntimport!=""{
		mntimport=strings.TrimSuffix(mntimport,"/")
	}
	if outpath!=""{
		outpath=strings.TrimSuffix(outpath,"/")
	}
	switch fun{
/*	case core.LOGIN:
//		doAuth(loginuser)
		testlogin()
		*/
	case core.ENCODE:
		doEncode()
	case core.DISTRIBUTE:
		doShare()
	case core.LIST:
		doList()
	case core.MOUNT:
		doMount()
	case core.TRACE:
		doTraceAll()
	case core.DECODE:
		doDecode()
	case core.VERSION:
		fmt.Println(version)
	default:
		fmt.Println("dtdfs(data defense) -enc|-list|-mnt|-share|-trace  -in INPUT_PATH [-out OUTPUTPATH] [-oname outdata_orgname] [-desc \"data_description\"] [-img container_image] [-import import_tool_path] [-user USERNAME] [-search KEYWORD]\nuse -h for more help")
	}
}
