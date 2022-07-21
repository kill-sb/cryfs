package main

import(
	"sort"
	"fmt"
	"io/ioutil"
	"os"
//	"errors"
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
		result=fmt.Sprintf("%d. Data Uuid :%s (Type: Local Encrypted Data)\n",index,data.Uuid)
	}else{
		result=fmt.Sprintf("Data Uuid :%s (Type: Local Encrypted Data)\n",data.Uuid)
	}
	result+=fmt.Sprintf("\tFilename :%s\n",inpath+"/"+data.Uuid)
	result+=fmt.Sprintf("\tOrgname :%s\n",data.OrgName)
	user,err:=GetUserName(data.OwnerId)
	if err==nil{
		result+=fmt.Sprintf("\tData Owner :%s\n",user)
		//result+=fmt.Sprintf("\tData Owner :%s(userid:%d)\n",user,data.OwnerId)
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

func MergeQueryObjs(token string, allobjs []*api.DataObj)(map[string]api.IFDataDesc,error){
	objmap:=make(map[string]*api.DataObj)
	objs:=make([]*api.DataObj,0,len(allobjs))
	for _,v:=range allobjs{
		if v.Type==core.RAWDATA{ // objmap is used for QueryObjs, skip RAWDATA
			continue
		}
		if _,ok:=objmap[v.Obj];!ok{
			objmap[v.Obj]=v
			objs=append(objs,v)
		}
	}
	dataret,err:=QueryObj_API(token,objs)
	if err!=nil{
		fmt.Println("MergeQueryObjs error:",err)
		return nil,err
	}
	mapret:=make(map[string]api.IFDataDesc)
	for _,v:=range dataret{
		mapret[v.GetUuid()]=v
		AddUserIdList(v.GetOwnerId())
	}
	return mapret,nil
}

func TraceRootObj(token string, root api.IFDataDesc)error{
	/*
	 1. Get traceparent,traceback,traceforword results in different obj slices
	 2. put objs into a map(objmap) with uuid-key, obj-value
	 3. put all elements in map into a slice, and use queryobjs to get objinfo
	 4. put queryobjs results into a map(infomap) with uuid-key, objinfo-value
	 5. show detail info in each element in obj slice in step1 by lookup infomap
	*/

	data:=&api.DataObj{Obj:root.GetUuid(),Type:root.GetType()}
	allobjs:=make([]*api.DataObj,0,50)

	bobjs,err:=TraceData_API(token,data,api.TRACE_BACK)
	if err!=nil{
		fmt.Println("Trace back of ",data.Obj," error:",err.Error())
		return err
	}
	allobjs=append(allobjs,bobjs...)

	fobjs,err:=TraceData_API(token,data,api.TRACE_FORWARD)
	if err!=nil{
		fmt.Println("Trace forward of ",data.Obj," error:",err.Error())
		return err
	}
	allobjs=append(allobjs,fobjs...)

	if len(allobjs)<1{
		return nil
	}

	infomap,err:=MergeQueryObjs(token,allobjs)
	if err!=nil{
		fmt.Println("TraceCSDFile error:",err.Error())
		return err
	}

	DisplayTraceResult(root,bobjs,fobjs,infomap)
	return nil
}

func TraceCSDFile(token string,fname string){
	head,err:=core.LoadShareInfoHead(fname)
	if err!=nil{
		fmt.Println("Load share info head error in TraceCSDFile:",err)
		return
	}
	uuid:=string(head.Uuid[:])
	sinfo,err:=GetShareInfo_User_API(token,uuid,0)
	if err!=nil{
		fmt.Println("GetShareInfo err:",err)
		return
	}
	AddUserIdList(sinfo.OwnerId)
	TraceRootObj(token,sinfo)
}

func TraceEncRCInfo(token string, dinfo *api.EncDataInfo){
	fmt.Println("\nData Run Context info:")
	if dinfo.FromRCId==0{
		fmt.Println("    The data is created from local plain file(s) directly, there's no run context info.")
		return
	}
	rcinfo,err:=GetRCInfo_API(dinfo.FromRCId)
	if err!=nil{
		fmt.Println("Get RCInfo error in TraceEncRCInfo: ",err)
		return
	}
	fmt.Printf("    Environment info:\n\tOS: %s,  Container Image: %s,  Client IP: %s,  Start Time: %s,  End Time: %s", rcinfo.OS,rcinfo.BaseImg,rcinfo.IPAddr,rcinfo.StartTime,rcinfo.EndTime)
	nsrc:=len(rcinfo.InputData)
	if nsrc>0{
		srcobjs:=make([]*api.DataObj,nsrc,nsrc)
		for i,v:=range rcinfo.InputData{
			srcobjs[i]=&api.DataObj{Obj:v.DataUuid,Type:v.DataType}
		}
		if infomap,err:=MergeQueryObjs("",srcobjs);err==nil{
			ClearTodoList()
			fmt.Println("\n    Parent Data Objects:")
			i:=0
			for _,v:=range infomap{
				i++
				fmt.Printf("\t%d. ",i)
				v.PrintDataInfo(0,keyword,GlobalGetUserName)
			}
		}
	}
	if len(rcinfo.ImportPlain)>0{
		fmt.Println("    Imported plain files info:")
		for i,v:=range rcinfo.ImportPlain{
			fmt.Printf("\t%d. File: %s,  Content description: %s,  Size: %d,  SHA256 sum: %s\n",i+1,v.RelName,v.FileDesc,v.Size,v.Sha256)
		}
	}else{
		fmt.Println("    Imported plain files info:  (N/A)")
	}

}

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

	AddUserIdList(dinfo.OwnerId)

	err=TraceRootObj(token,dinfo)
	if err==nil{
		TraceEncRCInfo(token,dinfo)
	}
}


func DisplayTraceResult(dinfo api.IFDataDesc,bobjs,fobjs []*api.DataObj,info map[string]api.IFDataDesc){
/*	should be done in MergeQueryObjs already, even if no yet, it will be done in GloblGetUserName as need
	for _,v:=range info{
		AddUserIdList(v.GetOwnerId())
	}
*/
	ClearTodoList()
	fmt.Println("Data Info:")
	dinfo.PrintDataInfo(1,keyword,GlobalGetUserName)

	if len(bobjs)>0{
		fmt.Println("\nTrace back result:")
	}else{
		fmt.Println("\nTrace back result:  (N/A)")
	}

	for i,v:=range bobjs{
		fmt.Printf("    %d. ",i+1)
		if v.Type==core.RAWDATA{
			result:="Data Obj: "+v.Obj+" (Type: Local Plain Data)"
			if keyword!=""{
				result=strings.Replace(result,keyword,"\033[7m"+keyword+"\033[0m", -1)
			}
			fmt.Println(result)
		}else{
			info[v.Obj].PrintDataInfo(0,keyword,GlobalGetUserName)
		}
	}

	if len(fobjs)>0{
		fmt.Println("\nTrace forward result:")
	}else{
		fmt.Println("\nTrace forward result:  (N/A)")
	}
	for i,v:=range fobjs{
		fmt.Printf("    %d. ",i+1)
		info[v.Obj].PrintDataInfo(0,keyword,GlobalGetUserName)
	}
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
	switch ftype{
	case core.ENCDATA:
		TraceEncData(linfo.Token,inpath)
	case core.CSDFILE:
		TraceCSDFile(linfo.Token,inpath)
	default:
		fmt.Println(inpath+" does not have valid data type.")
	}
}

func PrintShareDataInfo(sinfo *core.ShareInfo,index int)bool{
	var result string
	if index>0{
		result=fmt.Sprintf("\t%d. Uuid :%s (Type: Shared Data)\n",index,sinfo.Uuid)
	}else{
		result=fmt.Sprintf("\t%d. Uuid :%s (Type: ShareData)\n",index,sinfo.Uuid)
	}
	result+=fmt.Sprintf("\tFilename :%s\n",sinfo.FileUri)
	if sinfo.Sha256!=""{
		result+=fmt.Sprintf("\tSHA256sum: %s\n",sinfo.Sha256)
	}else{
		result+=fmt.Sprintf("\tSHA256sum: N/A\n")
	}
	result+=fmt.Sprintf("\tFrom data:")
	if sinfo.FromType==core.ENCDATA{
		result+="Encoded Data, "
	}else if sinfo.FromType==core.CSDFILE{
		result+="CSD File, "
	}
	result+=fmt.Sprintf("Parent uuid: %s\n",sinfo.FromUuid)
	user,err:=GetUserName(sinfo.OwnerId)
	if err==nil{
		result+=fmt.Sprintf("\tShared tag create user :%s\n",user)
//		result+=fmt.Sprintf("\tShared tag create user :%s(userid:%d)\n",user,sinfo.OwnerId)
	}
	result+=fmt.Sprintf("\tReceive users :%s\n",sinfo.Receivers)
	var perm string
	if sinfo.Perm==0{
		perm="ReadOnly"
	}else{
		perm="Resharable"
	}
	maxtime:="Infinite"
	if sinfo.MaxUse!=-1{
		maxtime=fmt.Sprintf("%d",sinfo.MaxUse)
	}
	result+=fmt.Sprintf("\tPermissions :Data Access Mode(%s), Expire Date(%s), Max Open Times(%s)\n",perm,sinfo.Expire,maxtime)

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

func GetUserName(id int32)(string,error){
	AddUserIdList(id)
	err:=ClearTodoList()
	if err!=nil{
		return "",err
	}
	return GlobalGetUserName(id),nil
}

