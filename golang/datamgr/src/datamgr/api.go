package main

import (
	"errors"
	"fmt"
    api "apiv1"
    core "coredata"
)

func GetDataInfo_API(uuid string)(*api.EncDataInfo,error){
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
    return ack.Data,nil
}

func GetUserInfo_API(ids []int32)([]api.UserInfoData,error){
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
    return ack.Data,nil
}

func FindUserName_API(names []string)([]api.UserInfoData,error){
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
    /*
    fmt.Println("listing result:")
    for _,v:=range ack.Data{
        fmt.Println(v.Id,v.Name)
    }*/
    return ack.Data,nil
}

/*
func UpdateDataInfo_API(dinfo *core.EncryptedData,linfo *core.LoginInfo) error{
    upreq:=api.UpdateDataInfoReq{Token:linfo.Token,Uuid:dinfo.Uuid,Hash256:dinfo.HashMd5}
    ack:=new (api.IUpdateDataAck)
    err:=HttpAPIPost(&upreq,ack,"updatedata")
    if err!=nil{
        return err
    }
    if ack.Code!=0{
        return errors.New(ack.Msg)
    }
    return nil
}*/

func SendMetaToServer_API(pdata *core.EncryptedData, token string)error{
    encreq:=api.EncDataReq{Token:token,Uuid:pdata.Uuid,Descr:pdata.Descr,IsDir:pdata.IsDir,FromRCId:pdata.FromRCId,OwnerId:pdata.OwnerId,OrgName:pdata.OrgName}
    ack:=new (api.IEncDataAck)
    err:=HttpAPIPost(&encreq,ack,"newdata")
    if err!=nil{
        fmt.Println("call api error:",err)
    }
    if ack.Code!=0{
        return errors.New(ack.Msg)
    }
    return err
}

func GetShareInfo_Public_API(uuid string)(*api.ShareInfoData,error){
    req:=&api.ShareInfoReq{Token:"0",Uuid:uuid,NeedKey:0}
    ack:=api.NewShareInfoAck()
    err:=HttpAPIPost(req,ack,"getshareinfo")
    if err!=nil{
        fmt.Println("call getshareinfo error:",err)
        return nil,err
    }
	if ack.Code!=0{
		return nil,errors.New(ack.Msg)
	}
    return ack.Data,nil
}

func GetShareInfo_User_API(token string, uuid string)(*api.ShareInfoData,error){
    req:=&api.ShareInfoReq{Token:token,Uuid:uuid,NeedKey:1}
    ack:=api.NewShareInfoAck()
    err:=HttpAPIPost(req,ack,"getshareinfo")
    if err!=nil{
        fmt.Println("call getshareinfo error:",err)
        return nil,err
    }
	if ack.Code!=0{
		return nil,errors.New(ack.Msg)
	}
    return ack.Data,nil

}

func ShareData_API(token string,sinfo *core.ShareInfo)(error){
    data:=FillShareReqData(sinfo)
    req:=&api.ShareDataReq{Token:token,Data:data}
    ack:=api.NewShareAck()
    err:=HttpAPIPost(req,ack,"sharedata")
    if err!=nil{
        fmt.Println("call getshareinfo error:",err)
        return err
    }
	if ack.Code!=0{
		return errors.New(ack.Msg)
	}
    return nil
}

func CreateRunContext_API(token string, rc *api.RCInfo)(error){
	req:=&api.CreateRCReq{Token:token,Data:rc}
	ack:=api.NewRCInfoAck()
	err:=HttpAPIPost(req,ack,"createrc")
    if err!=nil{
        fmt.Println("call updaterc error:",err)
        return err
    }
    if ack.Code!=0{
        return errors.New(ack.Msg)
    }
	return nil
}

func UpdateRunContext_API(token string, rc *api.RCInfo) error{
	//ack= new (api.ISimpleAck)
	req:=&api.UpdateRCReq{Token:token,RCId:rc.RCId,OutputUuid:rc.OutputUuid,EndTime:rc.EndTime}
    ack:=new (api.ISimpleAck)
    err:=HttpAPIPost(req,ack,"updaterc")
    if err!=nil{
        fmt.Println("call updaterc error:",err)
        return err
    }
	if ack.Code!=0{
		return errors.New(ack.Msg)
	}
	return nil
}

func GetRCInfo_API(rcid int64)(*api.RCInfo,error){
    req:=&api.GetRCInfoReq{Token:"0",RCId:rcid}
    ack:=api.NewRCInfoAck()
    err:=HttpAPIPost(req,ack,"getrcinfo")
    if err!=nil{
        fmt.Println("call getrcinfo error:",err)
        return nil,err
    }
	if ack.Code!=0{
		return nil,errors.New(ack.Msg)
	}
    return ack.Data,nil
}

func Logout_API(token string)error{
	req:=&api.LoginStatReq{Token:token}
	ack:=api.NewLoginStatAck()
	err:=HttpAPIPost(req,ack,"logout")
    if err!=nil{
        fmt.Println("call logout error:",err)
        return err
    }
    if ack.Code!=0{
        return errors.New(ack.Msg)
    }
	return nil
}
