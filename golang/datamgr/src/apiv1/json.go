package apiv1

type RetStat struct{
	Code int `json:"code"`
	Msg string `json:"message"`
}

type ISimpleAck RetStat

type AuthInfo struct{
    Name string `json:"name"`
    Passwd string `json:"passwd"`
    PriMask int32 `json:"primask"`
}

type TokenInfo struct{
    Id int32 `json:"id"`
    Token string `json:"token"`
    Key string `json:"key"`
    Timeout int32 `json:"timeout"`
}

type ITokenInfo struct{
	RetStat
	Data *TokenInfo `json:"data"`
}


type LoginStatReq struct{
	Token string `json:"token"`
}

type LoginStatInfo struct{
	Timeout int32 `json:"timeout"`
}

type ILoginStatAck struct{
	RetStat
	Data *LoginStatInfo `json:"data"`
}

type EncDataReq struct{
	Token string `json:"token"`
	Uuid string `json:"uuid"`
	Descr string `json:"descr"`
	IsDir   byte `json:"isdir"`
	FromRCId int64 `json:"fromrcid"`
	OwnerId int32 `json:"ownerid"`
	OrgName string `json:"orgname"`
}

type EncDataAck struct{
//	Uuid string `json:"uuid"`
//	LocalKey string `json:"locakkey"`
}

type IEncDataAck struct{
	RetStat
//	Data *EncDataAck `json:"data"`
}

type ShareInfoReq struct{
	Token string `json:"token"`
	Uuid string `json:"uuid"`
	NeedKey byte	`json:"needkey"`
}

type ShareInfoData struct{
	Uuid string `json:"uuid"`
	OwnerId int32 `json:"ownerid"`
	Sha256	string `json:"sha256"`
//	OwnerName string `json:"ownername"`
	Descr string `json:"descr"`
	Perm    int32 `json:"perm"`
	Receivers []string `json:"receivers"`
	RcvrIds []int32 `json:"rcvrids"`
	Expire  string `json:"expire"`
	MaxUse  int32 `json:"maxuse"`
	LeftUse int32 `json:"leftuse"`
	EncKey    string `json:"enckey"`
	FromType    int `json:"fromtype"`
	FromUuid    string `json:"fromuuid"`
	CrTime  string `json:"crtime"`
	OrgName string `json:"orgname"`
	IsDir byte `json:"isdir"`
}

type IShareInfoAck struct{
	RetStat
	Data *ShareInfoData `json:"data"`
}

type ShareDataReq struct{
	Token string `json:"token"`
	Data *ShareInfoData `json:"Data"`
}
/*
type ShareDataAck struct{
}
*/
type IShareDataAck struct{
	RetStat
//	Data *ShareDataAck `json:"data"`
}

type UserInfoData struct{
	Id int32 `json:"id"`
	Descr string `json:"descr"`
	Name string `json:"name"`
	Mobile string `json:"mobile"`
	Email string `json:"email"`
}

type GetUserReq struct{
	Token string `json:"token"`
	Id []int32 `json:"ids"`
}

type SearchUserReq struct{
	Keyword string `json:"keyword"`
} // for context search

type IUserInfoAck struct{
	RetStat
	Data []UserInfoData `json:"data"`
}

type FindUserNameReq struct{
	Token string `json:"token"`
	Name []string `json:"names"`
}

type GetDataInfoReq struct{
	Token string `json:"token"`
	Uuid string `json:"uuid"`
}

type UpdateDataInfoReq struct{
	Token string `json:"token"`
	Uuid string `json:"uuid"`
	Hash256	string `json:"sha256"`
}

/*
type IUpdateDataAck struct{
	RetStat
}*/

type SourceObj struct{
	DataType int `json:"type"`
	DataUuid string `json:"uuid"`
}

type ImportFile struct{
	RelName string `json:"relname"`
	FileDesc string `json:"desc"`
	Sha256  string `json:"sha256"`
	Size    int64 `json:"size"`
}

type RCInfo struct{
	RCId int64 `json:"rcid"`
	UserId int32 `json:"userid"`
	InputData []*SourceObj `json:"sources"`
	ImportPlain []*ImportFile `json:"imports"`
	OS string `json:"os"`
	BaseImg string `json:"baseimg"`
	IPAddr string `json:"ipaddr"`
	OutputUuid string `json:"output"`
	StartTime string `json:"start"`
	EndTime string `json:"end"`
}

type CreateRCReq struct{
	Token string `json:"token"`
	Data *RCInfo `json:"data"`
}

type UpdateRCReq struct{
	Token string `json:"token"`
	RCId int64 `json:"rcid"`
	OutputUuid string `json:"datauuid"`
	EndTime string `json:"endtime"`
}

type IRCInfoAck struct{
	RetStat
	Data *RCInfo	`json:"data"`
}

type GetRCInfoReq struct{
	Token string `json:"token"`
	RCId int64 `json:"rcid"`
}

type EncDataInfo struct{
	Uuid string `json:"uuid"`
	Descr string `json:"descr"`
	FromRCId int64	`json:"fromrcid"`
	SrcObj	[]*SourceObj `json:"srcobj"`
	OwnerId	int32	`json:"ownerid"`
	IsDir	byte	`json:"isdir"`
	OrgName	string	`json:"orgname"`
	CrTime string	`json:"crtime"`
}

type IDataInfoAck struct{
	RetStat
	Data *EncDataInfo	`json:"data"`
}

type DataObj struct{
	Obj	string	`json:"obj"`
	Type int `json:"type"`
}

type CommonTraceReq struct{
	Token string `json:"token"`
	Level int `json:"level"`
	Data *DataObj `json:"data"`
}

type TraceReq struct{
	Token string `json:"token"`
	Data *DataObj `json:"data"`
}


type ITraceAck struct{
	RetStat
	Data []*DataObj `json:"data"`
}

type QueryObjsReq struct{
	Token string `json:"token"`
	Data	[]*DataObj `json:"data"`
}

type IQueryObjsAck struct{
	RetStat
	Data []IFDataDesc `json:"data"`
}

type IFDataDesc interface{
	PrintDataInfo(int,string,func (int32)string) error
	GetOwnerId() int32
}


type SearchShareDataReq struct{
	Token string `json:"token"`
	FromId int32 `json:"fromid"`
	ToId int32 `json:"toid"`
	Start string `json:"startdate"`
	End string `json:"enddate"`
}

type ShareDataNode struct{
	Uuid string `json:"uuid"`
	LeftTimes int32 `json:"lefttimes"`
	FromId int32 `json:"fromid"`
	ToId int32 `json:"toid"`
	Crtime string `json:"crtime"`
}

type ISearchDataAck struct{
	RetStat
	Data []*ShareDataNode `json:"data"`
}

type NewExportReq struct{
	Token string `json:"token"`
	Data *DataObj `json:"data"`
	Comment string `json:"comment"`
}


type ProcDataObj struct{
	Uuid string `json:"uuid"`
	Type int `json:"type"`
	UserId int32 `json:"userid"`
}

type ExProcNode struct{
//	ExpId int64 `json:"expid"`
	ProcUid int32 `json:"procuid"`
	SrcData []*ProcDataObj `json:"srcdata"`
	Status int32	`json:"status"`
	Comment string `json:"comment"`
	ProcTime string `json:"proctime"`
}

type ExportProcInfo struct{
	ExpId int64 `json:"expid"`
	Status int32 `json:"status"`
	DstData *ProcDataObj `json:"dataobj"`
	CrTime string `json:"crtime"`
	Comment string `json:"comment"`
	ProcQueue []*ExProcNode `json:"procqueue"`
}

type IExProcAck struct{
	RetStat
	Data *ExportProcInfo `json:"data"`
}

type ExportProcReq struct{
	Token string `json:"token"`
	ExpId int64 `json:"expid"`
}  // response use IExProcAck

type SearchExpReq struct{
	Token string `json:"token"`
	FromUid int32 `json:"fromuid"`
	ToUid int32 `json:"touid"`
	Status int32 `json:"status"`
	Start string `json:"startdate"`
	End string `json:"enddate"`
}

type ISearchExpAck struct{
	RetStat
	Data []*ExportProcInfo `json:"data"`
}

type RespExpReq struct{
	Token string `json:"token"`
	ExpId int64 `json:"expid"`
	Status int32 `json:"status"`
	Comment string `json:"comment"`
}

type NotifyInfo struct{
	Id int64 `json:"id"`
	Type int32 `json:"type"`
	FromUid int32 `json:"fromuid"`
	ToUid int32 `json:"touid"`
	Content string `json:"content"`
	Comment string `json:"comment"`
	CrTime string `json:"crtime"`
	IsNew	int32 `json:"isnew"`
}

type SendNotifyReq struct{
	Token string `json:"token"`
	Data *NotifyInfo `json:"data"`
}

type ISendNotifyAck struct{
	RetStat
	Data int64 `json:"data"`
}

type SearchNotifiesReq struct{
	Token string `json:"token"`
	ToUid int32 `json:"touid"`
	FromUid int32 `json:"fromuid"`
	Type int32 `json:"type"`
	IsNew int32 `json:"isnew"`
}

type ISearchNotifiesAck struct{
	RetStat
	Data []*NotifyInfo `json:"data"`
}

type GetNotifyInfoReq struct{
	Token string `json:"token"`
	Ids []int64 `json:"ids"`
}

type SetNotifyStatReq struct{
	Token string `json:"token"`
	Ids []int64 `json:"ids"`
	Stats []int32 `json:"isnew"`
}

type DelNotifyReq struct{
	Token string `json:"token"`
	Ids []int64 `json:"ids"`
}

//type IRemoveNotifyAck-> ISimpleAck

type QueryNotifyReq struct{
	Token string `json:"token"`
	Id int64 `json:"notifyid"`
}

type IQueryNotifyAck struct{
	RetStat
	Data *NotifyInfo `json:"data"`
}

type ContactInfo struct{
	UserId int32 `json:"userid"`
	Name string `json:"name"`
}

type AddContactReq struct{
	Token string `json:"token"`
	Ids	[]int32	`json:"contactids"`
}

type DelContactReq struct{
    Token string `json:"token"`
    Ids []int32 `json:"contactids"`
}

type GetContactReq struct{
	Token string `json:"token"`
}

type FzSearchReq struct{
	Token string `json:"token"`
	Keyword string `json:"keyword"`
}

type IGetContactsAck struct{
	RetStat
	Data []*ContactInfo `json:"data"`
}
