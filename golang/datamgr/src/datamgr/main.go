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

func LoadConfig(){
	definpath=os.Getenv("DATA_IN_PATH")
	defoutpath=os.Getenv("HOME")+"/.cmitdata"
	defuser=os.Getenv("DATA_DEF_USER")
}

func GetFunction() int {
	var bList,bEnc,bTrace,bShare,bMnt,bDec bool
	flag.BoolVar(&bEnc,"e",false,"encrypt raw data")
	flag.BoolVar(&bShare,"s",false,"share data to other users")
	flag.BoolVar(&bMnt,"m",false,"mount encrypted data")
	flag.BoolVar(&bDec,"d",false,"decrypted local data(test only)")
	flag.BoolVar(&bTrace,"t",false,"trace source of data")
	flag.BoolVar(&bList,"l",false,"list local encrypted data")
	flag.StringVar(&inpath,"in",definpath,"original data path (may be a file or a directory)")
	flag.StringVar(&outpath,"out",definpath,"original data path (may be a file or a directory)")
	flag.StringVar(&loginuser,"user",defuser, "login user name")
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
	default:
		fmt.Println("Error parameters,use -h or --help for usage")
	}
}
