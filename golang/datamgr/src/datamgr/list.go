package main

import(
	"sort"
	"fmt"
	"io/ioutil"
	"os"
	"errors"
	"strings"
	api "apiv1"
	core "coredata"
)

type FTimeSort struct{
	data []string
}

func (s *FTimeSort)Len() int{
	return len(s.data)
}

func (s *FTimeSort)Swap(i,j int){
	tmp:=s.data[i]
	s.data[i]=s.data[j]
	s.data[j]=tmp
}

func (s *FTimeSort)Less(i,j int)bool{
	finfoi,erri:=os.Stat(s.data[i])
	finfoj,errj:=os.Stat(s.data[j])
	if erri!=nil || errj!=nil {
		return true
	}
	if finfoi.ModTime().Before(finfoj.ModTime()){
		return true
	}else{
		return false
	}
}

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
	var taglist,csdlist FTimeSort
	taglist.data=make([]string,0,len(dir))
	csdlist.data=make([]string,0,len(dir))
	for _,entry:=range dir{
		if !entry.IsDir(){
			fname:=entry.Name()
			if strings.HasSuffix(fname,".tag")|| strings.HasSuffix(fname,".TAG"){
				taglist.data=append(taglist.data,inpath+"/"+fname)

			}else if strings.HasSuffix(fname,".csd") || strings.HasSuffix(fname,".CSD"){
				csdlist.data=append(csdlist.data,inpath+"/"+fname)
			}
		}
	}
	sort.Sort(&taglist)
	sort.Sort(&csdlist)
	fmt.Println("\n***********************Local Encrypted Data**************************\n")
	ListTags(taglist.data)
	fmt.Println("\n***********************Shared Data From Users**************************\n")
	ListCSDs(csdlist.data)
}

func ListTags(tags[]string){
	i:=1
	for _,tag:=range tags{
		tinfo,err:=core.LoadTagFromDisk(tag)
		if err==nil{
			edata,err:=tinfo.GetDataInfo()
			if err==nil{
		//		fmt.Printf("\t%d\n",i+1)
				if PrintEncDataInfo(edata,i){
					i++
				}
			}else{
				fmt.Println(err)
			}
		}else{
			fmt.Println(err)
		}
	}
}

func PrintEncDataInfo(data *core.EncryptedData,index int)bool{
	result:=fmt.Sprintf("\t%d\n\tData Uuid :%s\n",index,data.Uuid)
	result+=fmt.Sprintf("\tFilename :%s\n",inpath+"/"+data.Uuid)
	user,err:=GetUserName(data.OwnerId)
	if err==nil{
		result+=fmt.Sprintf("\tData Owner :%s(%d)\n",user,data.OwnerId)
	}
	if data.FromType==core.RAWDATA{
		result+=fmt.Sprintf("\tFrom Type: Plain Local File\n")
		result+=fmt.Sprintf("\tOrginal filename :%s\n",data.OrgName)
	}else{
		result+=fmt.Sprintf("\tFrom Type: Shared Data\n")
		result+=fmt.Sprintf("\tFrom Shared Data Infomation :\n\t\tUuid :%s\n\t\tFileName:%s\n",data.FromObj,strings.TrimSuffix(data.OrgName,".outdata"))
	}
	result+=fmt.Sprintf("\tDescription :%s\n",data.Descr)
	if data.IsDir==1{
		result+=fmt.Sprintf("\tIs Directory :yes\n")
	}else{
		result+=fmt.Sprintln("\tIs Directory :no\n")
	}
	if keyword!=""{
		if strings.Contains(result,keyword){
			result=strings.Replace(result,keyword,"\033[7m"+keyword+"\033[0m", -1)
		}else{
			return false
		}
	}
	fmt.Print(result)
	fmt.Println("---------------------------------------------------------------------")
	return true
}

func ListCSDs(csds[]string){
	i:=1
    for _,csd:=range csds{
		head,err:=core.LoadShareInfoHead(csd)
		if err==nil{
			sinfo,err:=GetShareInfoFromHead(head,nil)
            if err==nil{
				sinfo.FileUri=csd
		//		fmt.Printf("\t%d\n",i+1)
                if PrintShareDataInfo(sinfo,i){
					i++
				}
            }else{
				fmt.Println(err)
			}
        }else{
				fmt.Println(err)
		}
    }
}
/*
func GetDataInfo_API(uuid string)(*api.IDataInfoAck,error){
	req:=&api.GetDataInfoReq{Token:"0",Uuid:uuid}
	ack:=api.NewDataInfoAck()
	err:=HttpAPIPost(req,ack,"getdatainfo")
	if err!=nil{
		fmt.Println("call api info error:",err)
		return nil,err
	}
	if ack.Code!=0{
		fmt.Println("request error:",ack.Msg)
		return nil,errors.New(ack.Msg)
	}
	return ack,nil
}
*/
func FillEncDataInfo(adata *api.EncDataInfo)*core.EncryptedData{
    info:=new (core.EncryptedData)
    info.Uuid=adata.Uuid
    info.OwnerId=adata.OwnerId
    info.Descr=adata.Descr
    info.FromType=adata.FromType
    info.FromObj=adata.FromObj
    info.HashMd5=adata.Hash256
    info.IsDir=adata.IsDir
    info.CrTime=adata.CrTime
    info.OrgName=adata.OrgName
	info.OwnerName,_=GetUserName(info.OwnerId)
    return info
}

func GetEncDataInfo(uuid string)(*core.EncryptedData,error){
	ainfo,err:=GetDataInfo_API(uuid)
	if err!=nil{
		return nil,err
	}
	if ainfo.Code!=0{
		return nil,errors.New(ainfo.Msg)
	}
	return FillEncDataInfo(ainfo.Data),nil
}

func traceRawData(tracer []core.InfoTracer,uuid string)([]core.InfoTracer,error){
	// RAW DATA
	dinfo,err:=GetEncDataInfo(uuid)
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
	sifack,err:=GetShareInfo_Public_API(uuid)
	if err!=nil{
		fmt.Println("GetShareInfo_Public_API error in traceCSDFile:",err)
		return nil,err
	}
	if sifack.Code!=0{
//		fmt.Println("error return value--msg:",sinfo.Msg)
		return nil,errors.New(sifack.Msg)
	}
	sinfo:=FillShareInfo(sifack.Data,uuid,0,0,nil)
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
		if err:=tracer[i].PrintTraceInfo(tab,keyword);err!=nil{
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

func PrintShareDataInfo(sinfo *core.ShareInfo,index int)bool{
	result:=fmt.Sprintf("\t%d\n\tShared tag Uuid :%s\n",index,sinfo.Uuid)
	result+=fmt.Sprintf("\tFilename :%s\n",sinfo.FileUri)
	user,err:=GetUserName(sinfo.OwnerId)
	if err==nil{
		result+=fmt.Sprintf("\tShared tag create user :%s(%d)\n",user,sinfo.OwnerId)
	}
	result+=fmt.Sprintf("\tReceive users :%s\n",sinfo.Receivers)
	var perm string
	if sinfo.Perm==0{
		perm="ReadOnly"
	}else{
		perm="Resharable"
	}
	maxtime:="Infinite"
	lefttime:="Infinite"
	if sinfo.MaxUse!=-1{
		maxtime=fmt.Sprintf("%d",sinfo.MaxUse)
		lefttime=fmt.Sprintf("%d",sinfo.LeftUse)
	}
	result+=fmt.Sprintf("\tPermissions :Data Access Mode(%s), Expire Date(%s), Left/Max Open Times(%s/%s)\n",perm,sinfo.Expire,lefttime,maxtime)

	result+=fmt.Sprintf("\tOriginal filename :%s\n",sinfo.OrgName)
	if sinfo.IsDir==1{
		result+=fmt.Sprintf("\tIs Directory :yes\n")
	}else{
		result+=fmt.Sprintf("\tIs Directory :no\n")
	}
	if keyword!=""{
		if strings.Contains(result,keyword){
			result=strings.Replace(result,keyword,"\033[7m"+keyword+"\033[0m", -1)
		}else{
			return false
		}
	}
	fmt.Print(result)
	fmt.Println("-----------------------------------------------------------------------")
	return true
}

func GetUserName(id int32)(ret string,err error){
	if ret,ok:=idmap[id];ok{
		return ret,nil
	}
	ids:=[]int32{id}
	ud,err:=GetUserInfo_API(ids)
	if err!=nil{
		return "",err
	}
	ret=ud.Data[0].Name
	idmap[id]=ret
	namemap[ret]=id
	return ret,nil
}
/*
func GetUserInfo_API(ids []int32)(*api.IUserInfoAck,error){
	req:=&api.GetUserReq{Token:"0",Id:ids}
    ack:=api.NewUserInfoAck()
    err:=HttpAPIPost(req,ack,"getuser")
    if err!=nil{
        fmt.Println("call api info error:",err)
        return nil,err
    }
	if ack.Code!=0{
		fmt.Println("request error:",ack.Msg)
		return nil,errors.New(ack.Msg)
	}
    return ack,nil
}

func FindUserName_API(names []string)(*api.IUserInfoAck,error){
	req:=&api.FindUserNameReq{Token:"0",Name:names}
	ack:=api.NewUserInfoAck()
	err:=HttpAPIPost(req,ack,"findusername")
	if err!=nil{
		fmt.Println("call api info error:",err)
		return nil,err
	}
	if ack.Code!=0{
		fmt.Println("request error:",ack.Msg)
		return nil,errors.New(ack.Msg)
	}
	return ack,nil
}*/
