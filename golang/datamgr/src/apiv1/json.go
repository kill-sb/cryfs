package apiv1

type RetStat struct{
	Code int `json:"code"`
	Msg string `json:"message"`
}

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
	FromType int `json:"fromtype"`
	FromObj string `json:"fromobj"`
	OwnerId int32 `json:"ownerid"`
	Hash256 string `json:"hash256"`
//	EncKey string `json:"enckey"`
	OrgName string `json:"orgname"`
}

type EncDataAck struct{
//	Uuid string `json:"uuid"`
//	LocalKey string `json:"locakkey"`
}

type IEncDataAck struct{
	RetStat
	Data *EncDataAck `json:"data"`
}

type ShareInfoReq struct{
	Token string `json:"token"`
	Uuid string `json:"uuid"`
	NeedKey byte	`json:"needkey"`
}

type ShareInfoData struct{
	Uuid string `json:"uuid"`
	OwnerId int32 `json:"ownerid"`
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
//	FileUri string `json:"fileuri"`
	OrgName string `json:"orgname"`
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

type IUpdateDataAck struct{
	RetStat
}

type EncDataInfo struct{
	Uuid string `json:"uuid"`
	Descr string `json:"descr"`
	FromType int	`json:"fromtype"`
	FromObj	string	`json:"fromobj"`
	OwnerId	int32	`json:"ownerid"`
	Hash256	string	`json:"sha256"`
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

type TraceBackReq struct{
	Token string `json:"token"`
	Data	[]DataObj `json:"data"`
}

type ITraceBackAck struct{
	RetStat
	Data [][]*DataObj `json:"data"`
}

type QueryObjsReq struct{
	Token string `json:"token"`
	Data	[]DataObj `json:"data"`
}

type IQueryObjsAck struct{
	RetStat
	Data []IFDataDesc `json:"dataobj"`
}

type IFDataDesc interface{
	PrintDataInfo(int,string,func (int32)string) error
}
