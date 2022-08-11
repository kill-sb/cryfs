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
	REFUSE=-1
	AGREE=iota
	WAITING
)

const (
    TEXTMSG=iota+1
    SHAREDATA
    EXPORTDATA
)

const (
	TRACE_PARENTS=-1
	TRACE_CHILDREN=1
	TRACE_BACK=-0xffff
	TRACE_FORWARD=0xffff
)

const (
	ERR_UNKNOWN=-1
	ERR_BADPARAM=iota
	ERR_LOGIN
	ERR_INVDATA
	ERR_ACCESS
	ERR_INTERNAL
)


func NewSimpleAck() *ISimpleAck{
	sa:=new (ISimpleAck)
	sa.Code=-1
	sa.Msg=INIT_MSG
	return sa
}

/*
func NewShareAck() *IShareDataAck{
//	data:=new (ShareDataAck)
	sda:=new (IShareDataAck)
	sda.Code=-1
	sda.Msg=INIT_MSG
//	sda.Data=data
	return sda
}*/

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
	ack.Data=make([]*UserInfoData,0,20)
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
/*
func NewDataAck() *IEncDataAck{
//	data:=new (EncDataAck)
	eda:=new (IEncDataAck)
	eda.Code=-1
	eda.Msg=INIT_MSG
//	eda.Data=data
	return eda
}*/

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

func NewQueryObjsAck(reqinfo []*DataObj)*IQueryObjsAck {
	cnt:=len(reqinfo)
	data:=make([]IFDataDesc,0,cnt)
	qda:=new (IQueryObjsAck)
	qda.Code=-1
	qda.Msg=INIT_MSG
	qda.Data=data
	return qda
}

func NewSearchEncAck(result []*EncDataNode) *ISearchEncAck{
	sea:=new (ISearchEncAck)
	sea.Code=-1
	sea.Msg=INIT_MSG
	sea.Data=result
	return sea
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

func NewApproveListAck() *ApproveListAck{
	ala:=new(ApproveListAck)
	ala.Code=-1
	ala.Msg=INIT_MSG
	ala.Data=nil
	return ala
}

func NewSearchExpAck() *ISearchExpAck{
	sea:=new (ISearchExpAck)
	sea.Code=-1
	sea.Msg=INIT_MSG
	sea.Data=make([]*ExportProcInfo,0,50)
	return sea
}

func NewGetContactsAck() *IGetContactsAck {
	gca:=new (IGetContactsAck)
	gca.Code=01
	gca.Msg=INIT_MSG
	gca.Data=make([]*ContactInfo,0,50)
	return gca
}

func (dinfo * EncDataInfo)GetOwnerId()int32{
	return dinfo.OwnerId
}

func (dinfo* EncDataInfo)GetUuid() string{
	return dinfo.Uuid
}

func (dinfo* EncDataInfo)GetType() int{
	return ENCDATA
}

func (dinfo* EncDataInfo)PrintDataInfo(level int, kw string,getuser func (int32)string)error{
    for i:=0;i<level;i++{
        fmt.Print("    ")
    }
   // fmt.Print("-->")
    var result string
    if dinfo.FromRCId<=0{
		result=fmt.Sprintf("Data Obj: %s (Type: Local Encrypted Data)  ",dinfo.Uuid)
        result+=fmt.Sprintf("From->Local Plain Data (%s),  ",dinfo.OrgName)
    }else {
		result=fmt.Sprintf("Data Obj: %s (Type: Local Encrypted Data),  ",dinfo.Uuid)
		result+=fmt.Sprintf("From->Encrypted/Shared Data, Original Name->%s, ",dinfo.OrgName)
    }
    result+=fmt.Sprintf("Owner->%s,  ",getuser(dinfo.OwnerId))
    //result+=fmt.Sprintf("Owner->%s(userid:%d),  ",getuser(dinfo.OwnerId),dinfo.OwnerId)

    result+=fmt.Sprintf("Create at->%s\n",dinfo.CrTime)
    if kw!=""{
        result=strings.Replace(result,kw,"\033[7m"+kw+"\033[0m", -1)
    }
    fmt.Print(result)
    return nil
}

func (sinfo* ShareInfoData)GetOwnerId()int32{
	return sinfo.OwnerId
}

func (sinfo* ShareInfoData)GetUuid() string{
    return sinfo.Uuid
}

func (sinfo* ShareInfoData)GetType() int{
    return CSDFILE
}


func (sinfo* ShareInfoData)PrintDataInfo(level int, kw string,getuser func(int32)string)error{
    for i:=0;i<level;i++{
        fmt.Print("    ")
    }
  //  fmt.Print("-->")

	result:=fmt.Sprintf("Data Obj :%s (Type: Shared Data)  ",sinfo.Uuid)
	if sinfo.Sha256==""{
		result+=fmt.Sprintf("SHA256sum :(N/A)  ")
	}else{
		result+=fmt.Sprintf("SHA256sum :%s  ",sinfo.Sha256)
	}
    result+=fmt.Sprintf("Owner->%s",getuser(sinfo.OwnerId))
    //result+=fmt.Sprintf("Owner->%s(userid :%d)",getuser(sinfo.OwnerId),sinfo.OwnerId)
    result+=fmt.Sprintf(", Send to->%s",sinfo.Receivers)
	//if sinfo.Expire!="2999-12-31 00:00:00"{
	if !strings.HasPrefix(sinfo.Expire,"2999-12-31"){
		result+=fmt.Sprintf(", Expire date:%s",sinfo.Expire)
	}
    if sinfo.Perm==0{
        result+=fmt.Sprintf(", Perm->ReadOnly")
    }
    if sinfo.MaxUse!=-1{
		result+=fmt.Sprintf(", Left/Max open times:%d/%d",sinfo.LeftUse,sinfo.MaxUse)
	}
    if sinfo.FromType==ENCDATA{
        result+=fmt.Sprintf(", From->Local Encrypted Data(UUID :%s)",sinfo.FromUuid)
    }else if sinfo.FromType==CSDFILE{
        result+=fmt.Sprintf(", From->User Shared Data(UUID :%s)",sinfo.FromUuid)
	}
	result+=fmt.Sprintf(", Original Name->%s, Create at->%s\n",sinfo.OrgName,sinfo.CrTime)

    if kw!=""{
        result=strings.Replace(result,kw,"\033[7m"+kw+"\033[0m", -1)
    }
    fmt.Print(result)
    return nil
}
