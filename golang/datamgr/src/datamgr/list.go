package main

import(
	"fmt"
	"io/ioutil"
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
	ListTags(tags)
	fmt.Println("\n*************************************************\n")
	ListCSDs(csds)
}

func ListTags(tags[]string){
	first:=true
	for _,tag:=range tags{
		tinfo,err:=core.LoadTagFromDisk(tag)
		if err==nil{
			edata,err:=tinfo.GetDataInfo()
			if err==nil{
				if first{
					fmt.Println("Encrypted local data:")
					first=false
				}
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

	fmt.Println("\tData filename :",data.Path)
	fmt.Println("\tOrginal filename :",data.FromObj)
	fmt.Println("\tDescription :",data.Descr)
	if data.IsDir==1{
		fmt.Println("\tIs Directory :yes")
	}else{
		fmt.Println("\tIs Directory :no")
	}
	fmt.Println("------------------------")
}

func ListCSDs(csds[]string){
	first:=true
    for _,csd:=range csds{
		head,err:=core.LoadShareInfoHead(csd)
		if err==nil{
			sinfo,err:=dbop.LoadShareInfo(head)
            if err==nil{
				sinfo.FileUri=csd
				if first{
					fmt.Println("Shared data from users:")
					first=false
				}
                PrintShareDataInfo(sinfo)
            }else{
				fmt.Println(err)
			}
        }else{
				fmt.Println(err)
		}
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
	fmt.Println("\tExpire datetime :",sinfo.Expire)
	orgname,err:=dbop.GetOrgFileName(sinfo)
	if err==nil{
		fmt.Println("\tOrignal filename :",orgname)
	}
	fmt.Println("------------------------")
}
