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
}

type ITokenInfo struct{
	RetStat
	Data *TokenInfo `json:"data"`
}

func NewToken()*ITokenInfo{
    data:=&TokenInfo{Id:-1,Token:"nil",Key:"nil"}
	token:=new (ITokenInfo)
	token.Code=-1
	token.Msg="Error Parameter"
	token.Data=data
    return token
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
	EncKey string `json:"enckey"`
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

func NewDataAck() *IEncDataAck{
	data:=new (EncDataAck)
	eda:=new (IEncDataAck)
	eda.Code=-1
	eda.Msg="Invalid parameter"
	eda.Data=data
	return eda
}

type ShareInfoReq struct{
	Token string `json:"token"`
	Uuid string `json:"uuid"`
}

type ShareInfoData struct{
	Uuid string `json:"uuid"`
	OwnerId int32 `json:"ownerid"`
	OwnerName string `json:"ownername"`
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
	FileUri string `json:"fileuri"`
	OrgName string `json:"orgname"`
}

type IShareInfoAck struct{
	RetStat
	Data *ShareInfoData `json:"data"`
}

func NewShareInfoAck()*IShareInfoAck{
    data:=new (ShareInfoData)
	data.RcvrIds=[]int32{-1}
	data.Receivers=[]string{""}
	ack:=new (IShareInfoAck)
	ack.Msg="Error Parameter"
	ack.Data=data
    return ack
}

type ShareDataReq struct{
	Token string `json:"token"`
	Data *ShareInfoData `json:"Data"`
}

type ShareDataAck struct{
}

type IShareDataAck struct{
	RetStat
	Data *ShareDataAck `json:"data"`
}

func NewShareAck() *IShareDataAck{
	data:=new (ShareDataAck)
	sda:=new (IShareDataAck)
	sda.Code=-1
	sda.Msg="Invalid parameter"
	sda.Data=data
	return sda
}
