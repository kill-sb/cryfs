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
			edata,err:=GetDataInfo(tinfo)
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
	var result string
	if index>0{
		result=fmt.Sprintf("\t%d\n\tData Uuid :%s\n",index,data.Uuid)
	}else{
		result=fmt.Sprintf("Data Info:\n\tData Uuid :%s\n",data.Uuid)
	}
	result+=fmt.Sprintf("\tFilename :%s\n",inpath+"/"+data.Uuid)
	result+=fmt.Sprintf("\tOrgname :%s\n",data.OrgName)
	user,err:=GetUserName(data.OwnerId)
	if err==nil{
		result+=fmt.Sprintf("\tData Owner :%s(%d)\n",user,data.OwnerId)
	}

	if data.FromRCId==0{
		result+="\tFrom: local plain data\n"
	}else{
		result+="\tFrom: data reprocessed in container\n"
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
			sinfo,err:=GetShareInfoFromHead(head,nil,0)
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
    info.Descr=adata.Descr
    info.IsDir=adata.IsDir
    info.OwnerId=adata.OwnerId
	info.OwnerName,_=GetUserName(info.OwnerId)
    info.FromRCId=adata.FromRCId
    info.OrgName=adata.OrgName
    info.CrTime=adata.CrTime
	if info.FromRCId>0{ // TODO need to be tested more 
		info.FromContext,_=GetRCInfo_API(info.FromRCId)
	}
	info.EncryptingKey=make([]byte,16)
    return info
}

func GetEncDataInfo(uuid string)(*core.EncryptedData,error){
	if dinfo,ok:=uuidmap[uuid];ok{
		return dinfo,nil
	}
	ainfo,err:=GetDataInfo_API(uuid)
	if err!=nil{
		return nil,err
	}
	dinfo:=FillEncDataInfo(ainfo)
	uuidmap[uuid]=dinfo
	return dinfo,nil
}

func traceRawData(tracer []core.InfoTracer,uuid string)([]core.InfoTracer,error){
	return nil,errors.New("Empty implement in traceRawData")
	// RAW DATA
/*	dinfo,err:=GetEncDataInfo(uuid)
	if err!=nil{
		fmt.Println("GetEncDataInfo error in traceRAWDATA:",err)
		return nil,err
	}
	tracer=append(tracer,dinfo)

    // should be replaced later because of multi-source processing

	if dinfo.FromType==core.RAWDATA{
		return tracer,nil
	}else if dinfo.FromType==core.CSDFILE{
		return traceCSDFile(tracer,dinfo.FromObj)
	}else{
		fmt.Println("Get unknown 'FromType' during tracing:",uuid,":",dinfo.FromType)
		return nil,errors.New("Unknown FromType")
	}*/
}


func traceCSDFile(tracer []core.InfoTracer,uuid string)([]core.InfoTracer ,error){
	return nil,errors.New("Empty implement in traceCSDFile")
	/*
	data,err:=GetShareInfo_Public_API(uuid)
	if err!=nil{
		fmt.Println("GetShareInfo_Public_API error in traceCSDFile:",err)
		return nil,err
	}
	sinfo:=FillShareInfo(data,uuid,0,0,nil)
	tracer=append(tracer,sinfo)

    // should be replaced later because of multi-source processing
	if sinfo.FromType==core.ENCDATA{
		return traceRawData(tracer,sinfo.FromUuid)
	}else if sinfo.FromType==core.CSDFILE{
		return traceCSDFile(tracer,sinfo.FromUuid)
	}else{
		fmt.Println("Get unknown 'FromType' during tracing:",uuid,":",sinfo.FromType)
		return nil,errors.New("Unknown FromType")
	}*/
}
/*
func TraceBack(token string,data *api.DataObj)([]*api.DataObj,error){
	return TraceData_API(token,data,api.TRACE_BACK)
}

func TraceForward(token string, data* api.DataObj)([]*api.DataObj,error){
	return TraceData_API(token,data,api.TRACE_FORWARD)
}

func TraceSource(token string,data* api.DataObj)([]*api.DataObj,error){

}*/

func TraceEncData(token string,fname string){
	st,err:=os.Stat(fname);
	if err!=nil{
		fmt.Println(fname," not found")
		return
	}
	uuid:=st.Name()
	if (!core.IsValidUuid(uuid)){
		fmt.Println(fname,"is not valid encoded data")
		return
	}
	dinfo,err:=GetDataInfo_API(uuid)
	if err!=nil{
		fmt.Println("Get data brief info error:",err)
		return
	}
	/*
	 1. Get traceparent,traceback,traceforword results in different obj slices
	 2. put objs into a map(objmap) with uuid-key, obj-value
	 3. put all elements in map into a slice, and use queryobjs to get objinfo
	 4. put queryobjs results into a map(infomap) with uuid-key, objinfo-value
	 5. show detail info in each element in obj slice in step1 by lookup infomap
	*/

	objmap:=make(map[string]*api.DataObj)
	infomap:=make(map[string]api.IFDataDesc)
	data:=&api.DataObj{Obj:uuid,Type:core.ENCDATA}

	// step 1 & 2
	pobjs,err:=TraceData_API(token,data,api.TRACE_PARENTS)
	if err!=nil{
		fmt.Println("Trace source of ",uuid," error:",err.Error())
		return
	}
	for _,obj:=range pobjs{
		if obj.Type>=0{
			objmap[obj.Obj]=obj
		}
	}

	bobjs,err:=TraceData_API(token,data,api.TRACE_BACK)
	if err!=nil{
		fmt.Println("Trace back of ",uuid," error:",err.Error())
		return
	}
	for _,obj:=range bobjs{
		if obj.Type>=0{
			objmap[obj.Obj]=obj
		}
	}
	fobjs,err:=TraceData_API(token,data,api.TRACE_FORWARD)
	if err!=nil{
		fmt.Println("Trace forward of ",uuid," error:",err.Error())
		return
	}
	for _,obj:=range fobjs{
		objmap[obj.Obj]=obj
	}
	cnt:=len(objmap)
	if cnt<1{
		return
	}

	// step3
	allobjs:=make([]*api.DataObj,0,len(objmap))
	for _,v:=range objmap{
		allobjs=append(allobjs,v)
	}
	retobjs,err:=QueryObj_API(token,allobjs)
	if err!=nil{
		fmt.Println("QueryObjs error:",err.Error())
		return
	}

	// step 4
	for i,v:=range retobjs{
		infomap[allobjs[i].Obj]=v
	}
	// step 5
	DisplayResult(dinfo,pobjs,bobjs,fobjs,infomap,dinfo.FromRCId)
}

func DisplayResult(dinfo* api.EncDataInfo,pobjs,bobjs,fobjs []*api.DataObj,info map[string]api.IFDataDesc,rcid int64){
	// gen userid - user map
	idmap:=make(map[int32]string)
	for _,v:=range info{
		idmap[v.GetOwnerId()]="unknown"
	}
	ids:=make([]int32,1,len(idmap)+1)
	ids[0]=dinfo.OwnerId
	for k,_:=range idmap{
		ids=append(ids,k)
	}
	users,err:=GetUserInfo_API(ids)
	if err!=nil{
		fmt.Println("GetUserInfo error:",err)
		return
	}
	for i,v:=range ids{
		idmap[v]=users[i].Name
	}
	lookupid:=func(id int32)string{
		return idmap[id]
	}
	fmt.Println("Data Info:\n")
	dinfo.PrintDataInfo(0,keyword,lookupid)
	// Show all obj's info
	fmt.Println("\nParent Objs:\n")
	for _,v:=range pobjs{
		if v.Type==core.RAWDATA{
			fmt.Println("\tLocal Plaine Data: ",v.Obj)
		}else{
			info[v.Obj].PrintDataInfo(1,keyword,lookupid)
		}
	}
	fmt.Println("\nTrace back result:\n")
	for _,v:=range bobjs{
		if v.Type==core.RAWDATA{
			fmt.Println("\tLocal Plaine Data: ",v.Obj)
		}else{
			info[v.Obj].PrintDataInfo(1,keyword,lookupid)
		}
	}
	fmt.Println("\nTrace forward result:\n")
	for _,v:=range fobjs{
		info[v.Obj].PrintDataInfo(1,keyword,lookupid)
	}
}

func TraceCSDFile(token string,fname string){

}

func doTraceAll(){
	if inpath==""{
		fmt.Println("use -in to set filename need to be traced")
		return
	}
	if loginuser==""{
		fmt.Println("use parameter -user=NAME to set login user")
        return
	}
	linfo,err:=Login(loginuser)
    if err!=nil{
        fmt.Println("login error:",err)
        return
    }
    defer Logout(linfo)

	ftype:=GetDataType(inpath)
	fmt.Println("type:",ftype)
	switch ftype{
	case core.ENCDATA:
		TraceEncData(linfo.Token,inpath)
	case core.CSDFILE:
		TraceCSDFile(linfo.Token,inpath)
	default:
		fmt.Println(inpath+" does not have valid data type.")
	}
}

/*
func doTraceAll(){
	if inpath==""{
		fmt.Println("use -in to set filename need to be traced")
		return
	}
	ftype:=GetDataType(inpath)
	var tracer =make([]core.InfoTracer,0,20)
	switch ftype{
// should be replaced later because of multi-source processing, more cases like  core.ENCDATA

	case core.ENCDATA:
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
*/

func PrintShareDataInfo(sinfo *core.ShareInfo,index int)bool{
	result:=fmt.Sprintf("\t%d\n\tShared tag Uuid :%s\n",index,sinfo.Uuid)
	result+=fmt.Sprintf("\tFilename :%s\n",sinfo.FileUri)
	result+=fmt.Sprintf("\tFrom data:")
	if sinfo.FromType==core.ENCDATA{
		result+="Encoded Data, "
	}else if sinfo.FromType==core.CSDFILE{
		result+="CSD File, "
	}
	result+=fmt.Sprintf("UUID: %s\n",sinfo.FromUuid)
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
	ret=ud[0].Name
	idmap[id]=ret
	namemap[ret]=id
	return ret,nil
}

