package main

import(
	"fmt"
	"io/ioutil"
	"errors"
	"strings"
	"dbop"
	core "coredata"
)

func doList(){
/*	linfo,err:=Login(loginuser)
	if err!=nil{
		fmt.Println("Login error:",err)
		return
	}
	*/
	if inpath==""{
		fmt.Println("-in need to be set to a directory search in")
		return
	}
	dir,err:=ioutil.ReadDir(inpath)
	if err!=nil{
		fmt.Println("Read dir error:",err)
		return
	}
	tags:=make([]string,0,len(dir))
	csds:=make([]string,0,len(dir))
	for _,entry:=range dir{
		if !entry.IsDir(){
			fname:=entry.Name()
			if strings.HasSuffix(fname,".tag")|| strings.HasSuffix(fname,".TAG"){
				tags=append(tags,inpath+"/"+fname)

			}else if strings.HasSuffix(fname,".csd") || strings.HasSuffix(fname,".CSD"){
				csds=append(csds,inpath+"/"+fname)
			}
		}
	}
	fmt.Println("\n***********************Local Encrypted Data**************************\n")
	ListTags(tags)
	fmt.Println("\n***********************Shared Data From Users**************************\n")
	ListCSDs(csds)
}

func ListTags(tags[]string){
	for i,tag:=range tags{
		tinfo,err:=core.LoadTagFromDisk(tag)
		if err==nil{
			edata,err:=tinfo.GetDataInfo()
			if err==nil{
				fmt.Printf("\t%d\n",i+1)
				PrintEncDataInfo(edata)
			}else{
				fmt.Println(err)
			}
		}else{
			fmt.Println(err)
		}
	}
}

func PrintEncDataInfo(data *core.EncryptedData){
	fmt.Println("\tData Uuid :",data.Uuid)
	fmt.Println("\tFilename :",inpath+"/"+data.Uuid)
	user,err:=dbop.GetUserName(data.OwnerId)
	if err==nil{
		fmt.Printf("\tData Owner :%s(%d)\n",user,data.OwnerId)
	}
	if data.FromType==core.RAWDATA{
		fmt.Println("\tFrom Type: Plain Local File")
		fmt.Println("\tOrginal filename :",data.OrgName)
	}else{
		fmt.Println("\tFrom Type: Shared Data")
		fmt.Println("\tFrom Shared Data Infomation :\n\t\tUuid :"+data.FromObj+"\n\t\tFileName :"+strings.TrimSuffix(data.OrgName,".outdata"))
	}
	fmt.Println("\tDescription :",data.Descr)
	if data.IsDir==1{
		fmt.Println("\tIs Directory :yes")
	}else{
		fmt.Println("\tIs Directory :no")
	}
	fmt.Println("---------------------------------------------------------------------")
}

func ListCSDs(csds[]string){
    for i,csd:=range csds{
		head,err:=core.LoadShareInfoHead(csd)
		if err==nil{
			sinfo,err:=dbop.LoadShareInfo(head)
            if err==nil{
				sinfo.FileUri=csd
				fmt.Printf("\t%d\n",i+1)
                PrintShareDataInfo(sinfo)
            }else{
				fmt.Println(err)
			}
        }else{
				fmt.Println(err)
		}
    }
}

func traceRawData(tracer []core.InfoTracer,uuid string)([]core.InfoTracer,error){
	// RAW DATA
	dinfo,err:=dbop.GetEncDataInfo(uuid)
	if err!=nil{
		fmt.Println("GetEncDataInfo error in traceRAWDATA:",err)
		return nil,err
	}
	tracer=append(tracer,dinfo)
	if dinfo.FromType==core.RAWDATA{
		return tracer,nil
	}else if dinfo.FromType==core.CSDFILE{
		return traceCSDFile(tracer,dinfo.FromObj)
	}else{
		fmt.Println("Get unknown 'FromType' during tracing:",uuid,":",dinfo.FromType)
		return nil,errors.New("Unknown FromType")
	}
}

func traceCSDFile(tracer []core.InfoTracer,uuid string)([]core.InfoTracer ,error){
	sinfo,err:=dbop.GetBriefShareInfo(uuid)
	if err!=nil{
		fmt.Println("GetBriefShareInfo error in traceCSDFile:",err)
		return nil,err
	}
	tracer=append(tracer,sinfo)
	if sinfo.FromType==core.RAWDATA{
		return traceRawData(tracer,sinfo.FromUuid)
	}else if sinfo.FromType==core.CSDFILE{
		return traceCSDFile(tracer,sinfo.FromUuid)
	}else{
		fmt.Println("Get unknown 'FromType' during tracing:",uuid,":",sinfo.FromType)
		return nil,errors.New("Unknown FromType")
	}
}

func doTraceAll(){
	if inpath==""{
		fmt.Println("use -in to set filename need to be traced")
		return
	}
	ftype:=GetDataType(inpath)
	var tracer =make([]core.InfoTracer,0,20)
	switch ftype{
	case core.RAWDATA:
	    if tag,err:=core.LoadTagFromDisk(inpath);err!=nil{
			fmt.Println("Load tag info error in traceAll:",err)
			return
		}else{
			tracer,err=traceRawData(tracer,string(tag.Uuid[:]))
			if err!=nil{
				fmt.Println("trace Rawdata error:",err)
				return
			}
		}
	case core.CSDFILE:
		if head,err:=core.LoadShareInfoHead(inpath);err!=nil{
			fmt.Println("Load share info head error in traceAll:",err)
			return
		}else{
			tracer,err=traceCSDFile(tracer,string(head.Uuid[:]))
		}
	default:
		fmt.Println("Unknown data type.")
		return
	}
	length:=len(tracer)
	for i:=length-1;i>=0;i--{
		tab:=length-1-i
		if err:=tracer[i].PrintTraceInfo(tab);err!=nil{
			return
		}else{
			if i!=0{
				for j:=0;j<=tab;j++{
					fmt.Print("\t")
				}
				fmt.Println("|")
				for j:=0;j<=tab;j++{
					fmt.Print("\t")
				}
				fmt.Println("|")
			}
		}
	}
}

func doTrace(){
	if inpath==""{
		fmt.Println("use -in to set .csd file full pathname")
		return
	}
	head,err:=core.LoadShareInfoHead(inpath)
	var sinfo *core.ShareInfo
	list:=make ([]*core.ShareInfo,0,50)
	if err==nil{
		sinfo,err=dbop.LoadShareInfo(head)
		if err!=nil{
			fmt.Println("load share info from head error",err)
			return
		}
		for ;sinfo.FromType==core.CSDFILE;{
			list=append(list,sinfo)
			sinfo,err=dbop.GetBriefShareInfo(sinfo.FromUuid)
		}

	}else{
		fmt.Println("parse csd file error:",err)
	}
	list=append(list,sinfo)
	orgdata,err:=dbop.GetEncDataInfo(sinfo.FromUuid)
	if err!=nil{
		fmt.Println("load data from db error:",err)
		return
	}
	dtuser,err:=dbop.GetUserName(orgdata.OwnerId)
	if err!=nil{
		fmt.Println("unknown user id :",orgdata.OwnerId)
		return
	}

	fmt.Println("--------------------orginal data info-----------------------------")
	fmt.Println("original file:",orgdata.OrgName,",  owner:",dtuser,", file uuid :",orgdata.Uuid)
	fmt.Println("\n---------------------file spread info------------------------------")
	var space=4;
	for i:=len(list)-1;i>=0;i--{
		for j:=0;j<space;j++{
			fmt.Print(" ")
		}
		user,err:=dbop.GetUserName(list[i].OwnerId)
		if err!=nil{
			fmt.Println("unknown user id :",list[i].OwnerId)
			return
		}
		fmt.Println(user,"-->",list[i].Receivers,"at ",list[i].CrTime,",permission",list[i].Perm,",","uuid:",list[i].Uuid)
		space+=4;
	}
}

func PrintShareDataInfo(sinfo *core.ShareInfo){
	fmt.Println("\tShared tag Uuid :",sinfo.Uuid)
	fmt.Println("\tFilename :",sinfo.FileUri)
	user,err:=dbop.GetUserName(sinfo.OwnerId)
	if err==nil{
		fmt.Printf("\tShared tag create user :%s(%d)\n",user,sinfo.OwnerId)
	}
	fmt.Println("\tReceive users :",sinfo.Receivers)
	var perm string
	if sinfo.Perm==0{
		perm="ReadOnly"
	}else{
		perm="Resharable"
	}
	fmt.Println("\tPermission :",perm)
	orgname,err:=dbop.GetOrgFileName(sinfo)
	if err==nil{
		fmt.Println("\tOriginal filename :",orgname)
	}
	if sinfo.IsDir==1{
		fmt.Println("\tIs Directory :yes")
	}else{
		fmt.Println("\tIs Directory :no")
	}

	fmt.Println("-----------------------------------------------------------------------")
}
