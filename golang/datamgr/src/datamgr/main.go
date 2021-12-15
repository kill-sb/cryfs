package main

import(
	"fmt"
	"flag"
	"os"
	core "coredata"
)

const AES_KEY_LEN=128

var definpath , inpath string
var defoutpath,outpath string
var defuser, loginuser string
var config string

func LoadConfig(){
	definpath=os.Getenv("DATA_IN_PATH")
	defoutpath=os.Getenv("HOME")+"/.cmitdata"
	defuser=os.Getenv("DATA_DEF_USER")
}

func GetFunction() int {
	var bList,bEnc,bTrace,bShare,bMnt,bDec bool
	flag.BoolVar(&bEnc,"enc",false,"encrypt raw data")
	flag.BoolVar(&bShare,"share",false,"share data to other users")
	flag.BoolVar(&bMnt,"mnt",false,"mount encrypted data")
	flag.BoolVar(&bDec,"dec",false,"decrypted local data(test only)")
	flag.BoolVar(&bTrace,"trace",false,"trace source of data")
	flag.BoolVar(&bList,"list",false,"list local encrypted data")
	flag.StringVar(&inpath,"in",definpath,"original data path (may be a file or a directory)")
	flag.StringVar(&outpath,"out",definpath,"original data path (may be a file or a directory)")
	flag.StringVar(&loginuser,"user",defuser, "login user name")
	flag.StringVar(&config,"config","", "use config file to decribe share info")
	flag.Parse()
	ret:=core.INVALID
	count:=0
	if(bList){
		ret=core.LIST
		count++
	}
	if(bTrace){
		ret=core.TRACE
		count++
	}
	if(bDec){
		ret=core.DECODE
		count++
	}
	if bEnc{
		ret= core.ENCODE
		count++
	}
	if bShare{
		ret=core.DISTRIBUTE
		count++
	}
	if bMnt{
		ret=core.MOUNT
		count++
	}
	if count!=1{
		ret=core.INVALID
	}
	return ret
}

func main(){
	LoadConfig()
	fun:=GetFunction()
	switch fun{
	case core.ENCODE:
		doEncode()
	case core.DISTRIBUTE:
		doShare()
	case core.MOUNT:
	case core.DECODE:
		doDecode()
	case core.TRACE:
		doTrace()
	case core.LIST:
		doList()
	default:
		fmt.Println("datamgr -enc|-dec|-list|-mnt|-share|-trace  -in INPUT_PATH [-out OUTPUTPATH] [-config CONFIGFILE] (use -h for more help)")
	}
}
