package apiv1

import (
	"fmt"
	"strings"
)

const INIT_MSG string ="Invalid parameter"

const (
    RAWDATA=iota-1
    ENCDATA
    CSDFILE
    UNKNOWN=0xff
)

const (
	REFUSE=iota-1
	WAITING
	AGREE
)

func NewSimpleAck() *ISimpleAck{
	sa:=new (ISimpleAck)
	sa.Code=-1
	sa.Msg=INIT_MSG
	return sa
}

func NewShareAck() *IShareDataAck{
//	data:=new (ShareDataAck)
	sda:=new (IShareDataAck)
	sda.Code=-1
	sda.Msg=INIT_MSG
//	sda.Data=data
	return sda
}

func NewDataInfoAck() *IDataInfoAck{
	data:=new (EncDataInfo)
	dia:=new (IDataInfoAck)
	dia.Code=-1
	dia.Msg=INIT_MSG
	dia.Data=data
	return dia
}

func NewRCInfoAck() *IRCInfoAck{
	data:=new (RCInfo)
	ria:=new (IRCInfoAck)
	ria.Code=-1
	ria.Msg=INIT_MSG
	ria.Data=data
	return ria
}

func NewTraceAck()*ITraceAck{
	data:=make([]*DataObj,0,20)
	tfa:=new (ITraceAck)
	tfa.Code=-1
	tfa.Msg=INIT_MSG
	tfa.Data=data
	return tfa

}
/*
func NewUpdateDataAck()*IUpdateDataAck{
	ack:=new (IUpdateDataAck)
	ack.Code=-1
	ack.Msg=INIT_MSG
	return ack
}*/

func NewUserInfoAck() *IUserInfoAck{
	ack:=new (IUserInfoAck)
	ack.Code=-1
	ack.Msg=INIT_MSG
	ack.Data=make([]UserInfoData,0,20)
	return ack
}
func NewShareInfoAck()*IShareInfoAck{
    data:=new (ShareInfoData)
	data.RcvrIds=make([]int32,0,20)
	data.Receivers=make([]string,0,20)
	ack:=new (IShareInfoAck)
	ack.Msg="Invalid Parameter"
	ack.Data=data
    return ack
}

func NewDataAck() *IEncDataAck{
//	data:=new (EncDataAck)
	eda:=new (IEncDataAck)
	eda.Code=-1
	eda.Msg=INIT_MSG
//	eda.Data=data
	return eda
}

func NewToken()*ITokenInfo{
    data:=&TokenInfo{Id:-1,Token:"nil",Key:"nil"}
	token:=new (ITokenInfo)
	token.Code=-1
	token.Msg="Invalid Parameter"
	token.Data=data
    return token
}

func NewLoginStatAck() *ILoginStatAck{
	data:=&LoginStatInfo{0}
	lsa:=new (ILoginStatAck)
	lsa.Data=data;
	lsa.Code=-1
	lsa.Msg=INIT_MSG
	return lsa
}

func NewQueryObjsAck(reqinfo []DataObj)*IQueryObjsAck {
	cnt:=len(reqinfo)
	data:=make([]IFDataDesc,0,cnt)
	qda:=new (IQueryObjsAck)
	qda.Code=-1
	qda.Msg=INIT_MSG
	qda.Data=data
	return qda
}

func NewSearchDataAck(result []*ShareDataNode) *ISearchDataAck{
	sda:=new (ISearchDataAck)
	sda.Code=-1
	sda.Msg=INIT_MSG
	sda.Data=result
	return sda
}

func NewSendNotifyAck()*ISendNotifyAck{
	sna:=new (ISendNotifyAck)
	sna.Code=-1
	sna.Msg=INIT_MSG
	return sna
}

func NewSearchNotifiesAck() *ISearchNotifiesAck{
	sna:=new (ISearchNotifiesAck)
	sna.Code=-1
	sna.Msg=INIT_MSG
//	sna.Data=make([]*NotifyInfo,0,50)
	return sna
}

func NewQueryNotifyAck() *IQueryNotifyAck{
	qna:=new (IQueryNotifyAck)
	qna.Code=-1
	qna.Msg=INIT_MSG
	return qna
}

func NewExProcAck() *IExProcAck{
	epa:=new (IExProcAck)
	epinfo:=new (ExportProcInfo)
	epa.Data=epinfo
	epa.Code=-1
	epa.Msg=INIT_MSG
	return epa
}


func (dinfo* EncDataInfo)PrintDataInfo(level int, keyword string,getuser func (int32)string)error{
    for i:=0;i<level;i++{
        fmt.Print("\t")
    }
    fmt.Print("-->")
    var result string
    if dinfo.FromRCId==-1{
        result=fmt.Sprintf("Local Encrypted Data(UUID: %s)  Details :",dinfo.Uuid)
    }else {
        result=fmt.Sprintf("Reprocessed Local Encrypted Data(UUID: %s)  Details :",dinfo.Uuid)
    }
    result+=fmt.Sprintf("Owner->%s(uid:%d)",getuser(dinfo.OwnerId),dinfo.OwnerId)
    if dinfo.FromRCId==-1{
        result+=fmt.Sprintf(", From Local Plain Data->%s",dinfo.OrgName)
    }else {
		result+=fmt.Sprintf(", From Encrypted/Shared Data, Original Name: %s",strings.TrimSuffix(dinfo.OrgName,".outdata"))// UUID->%s(Orginal Filename :%s)",dinfo.FromObj,strings.TrimSuffix(dinfo.OrgName,".outdata"))
    }

    result+=fmt.Sprintf(", Create at->%s\n",dinfo.CrTime)
    if keyword!=""{
        result=strings.Replace(result,keyword,"\033[7m"+keyword+"\033[0m", -1)
    }
    fmt.Print(result)
    return nil
}

func (sinfo* ShareInfoData)PrintDataInfo(level int, keyword string,getuser func(int32)string)error{
    for i:=0;i<level;i++{
        fmt.Print("\t")
    }
    fmt.Print("-->")

    result:=fmt.Sprintf("Shared Data(UUID :%s)  Details :",sinfo.Uuid)
    result+=fmt.Sprintf("Owner->%s(uid :%d)",getuser(sinfo.OwnerId),sinfo.OwnerId)
    result+=fmt.Sprintf(", Send to->%s",sinfo.Receivers)
    if sinfo.Perm==0{
        result+=fmt.Sprintf(", Perm->ReadOnly")
    }else{
        result+=fmt.Sprintf(", Perm->Resharable")
    }
/*    if sinfo.FromRCId==-1{
        result+=fmt.Sprintf(", From->Local Encrypted Data(UUID :%s)",sinfo.FromUuid)
    }else{
        result+=fmt.Sprintf(", From->User Shared Data(UUID :%s)",sinfo.FromUuid)
    }*/
    result+=fmt.Sprintf(", Create at->%s\n",sinfo.CrTime)

    if keyword!=""{
        result=strings.Replace(result,keyword,"\033[7m"+keyword+"\033[0m", -1)
    }
    fmt.Print(result)
    return nil
}
