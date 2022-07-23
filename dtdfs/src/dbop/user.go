package dbop
// todo: use map to cache db operate result

import (
	_ "MySQL"
	"errors"
	"fmt"
	"strings"
	api "apiv1"
)

func ParseVisitors(recvlist string) ([]string,[]int32,error){
    strret:=strings.Split(recvlist,",")
    intret:=make([]int32,0,len(strret))
    for _,user:=range strret{
        user=strings.TrimSpace(user)
        id,err:=IsValidUser(user) // should fix to asking server later
        if err!=nil{
            return nil,nil,err
        }
        intret=append(intret,id)
    }
    return strret,intret,nil
}

func GetUserNames(uids []int32)([]string,error){
	n:=len(uids)
	if n<1{
		return nil,errors.New("empty user list in GetUserNames")
	}
	db:=GetDB()

	ret:=make([]string,n)
	query:=fmt.Sprintf("select name from users where id='%d'",uids[0])
	for i:=1;i<n;i++{
		query+=fmt.Sprintf(" or id='%d'",uids[i])
	}
	res,err:=db.Query(query)
	if err!=nil{
		return ret,err
	}
	defer res.Close()
	i:=0
	for res.Next(){
		err=res.Scan(&ret[i])
		if err!=nil{
			return nil,err
		}
		i++
	}
	if i!=n{
		return nil,errors.New("GetUserNames error in dbop, check your id list")
	}
	return ret,nil

}

func IsValidUser(user string)(int32,error){
	var ret int32 =-1
	db:=GetDB()
	query:=fmt.Sprintf("select id from users where name='%s'",user)
	res,err:=db.Query(query)
	if err!=nil{
		return ret,err
	}
	defer res.Close()
	if !res.Next(){
		return ret,errors.New("No such user ")
	}else{
		res.Scan(&ret)
	}
	return ret,nil
}

func GetUserInfoByName(name string)(*api.UserInfoData,error){
	ret,ok:=usernamecache[name]
	if ok{
		return ret,nil
	}
	ret=new (api.UserInfoData)
	ret.Name=name
	db:=GetDB()
	query:=fmt.Sprintf("select descr,id,mobile,email from users where name='%s'",name)
	res,err:=db.Query(query)
	if err!=nil{
		fmt.Println("Query error:",err)
		return ret,err
	}
	defer res.Close()
	if !res.Next(){
		ret.Id=-1
		fmt.Println("error",err)
		return ret,nil
	}else{
		res.Scan(&ret.Descr,&ret.Id,&ret.Mobile,&ret.Email)
	}
	usernamecache[name]=ret
	return ret,nil
}

func GetUserInfo(id int32)(*api.UserInfoData,error){
	ret,ok:=userinfocache[id]
	if ok{
		return ret,nil
	}
	ret=new (api.UserInfoData)
	ret.Id=id
	db:=GetDB()
	query:=fmt.Sprintf("select descr,name,mobile,email from users where id=%d",id)
	res,err:=db.Query(query)
	if err!=nil{
		return nil,err
	}
	defer res.Close()
	if !res.Next(){
		return nil,errors.New("No such user ")
	}else{
		res.Scan(&ret.Descr,&ret.Name,&ret.Mobile,&ret.Email)
	}
	userinfocache[id]=ret
	return ret,nil
}

func LookupPasswdSHA(user string)(int32,string,string,error){
	db:=GetDB()
	query:=fmt.Sprintf("select id,pwdsha256,enclocalkey from users where name='%s'",user)
	if strings.Contains(user,"@"){
		query+=fmt.Sprintf(" or email='%s'",user)
	}
	res,err:=db.Query(query)
	if err!=nil{
		return -1,"","",err
	}
	defer res.Close()
	if res.Next(){
		var key string
		var shasum string
		var id int32
		if err:=res.Scan(&id,&shasum,&key);err!=nil{
			return -1,"","",err
		}else{
			return id,shasum,key,nil
		}
	}
	return -1,"","",errors.New("No such user")
}

func NewContact(uid, cid int32)error{
    db:=GetDB()
	query:=fmt.Sprintf("select count(*) from contacts where userid=%d and contactuserid=%d",uid,cid)
	res,err:=db.Query(query)
    if err!=nil{
        fmt.Println("select from contacts error:",err)
        return err
    }
	defer res.Close()
    if res.Next(){
        var count int64
        err=res.Scan(&count)
        if err!=nil{
            return err
        }
		if count>0{
			return nil //errors.New("contact exists already")
		}
	}
    query=fmt.Sprintf("insert into contacts (userid, contactuserid) values (%d,%d)",uid,cid)
    if _, err= db.Exec(query); err == nil {
		return nil
	}else{
		return err
	}
}

func ListContacts(uid int32)([]*api.ContactInfo,error){
	db:=GetDB()
	query:=fmt.Sprintf("select contacts.contactuserid,users.name from contacts,users where contacts.userid=%d and users.id=contacts.contactuserid",uid)
	res, err := db.Query(query)
	if err != nil {
		fmt.Println("query contacts error:",query)
		return nil,err
	}
	defer res.Close()
	clist:=make([]*api.ContactInfo,0,50)
	for res.Next(){
		cinfo:=new(api.ContactInfo)
		err=res.Scan(&cinfo.UserId,&cinfo.Name)
		if err!=nil{
			return nil,err
		}
		clist=append(clist,cinfo)
	}
	return clist,nil
}

func FuzzySearch(uid int32, keyword string)([]*api.ContactInfo,error){
	db:=GetDB()
	query:=fmt.Sprintf("select contacts.contactuserid,users.name from contacts,users where contacts.userid=%d and contacts.contactuserid=users.id and users.name like '%s'",uid,"%"+keyword+"%")
	res,err:=db.Query(query)
	if err!=nil{
		fmt.Println("query error:",query)
		return nil,err
	}
	defer res.Close()
	clist:=make([]*api.ContactInfo,0,50)
	for res.Next(){
		cinfo:=new(api.ContactInfo)
		err=res.Scan(&cinfo.UserId,&cinfo.Name)
		if err!=nil{
			return nil,err
		}
		clist=append(clist,cinfo)
	}
	return clist,nil
}

func DelContact(uid,cid int32 )error{
	db:=GetDB()
	query:=fmt.Sprintf("delete from contacts where userid=%d and contactuserid=%d",uid,cid)
	if _,err:= db.Exec(query);err != nil{
		fmt.Println("DelContact error:",query)
		return err
	}
	return nil
}

