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
