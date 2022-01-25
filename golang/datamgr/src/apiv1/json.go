package apiv1

type AuthInfo struct{
    Name string `json:"name"`
    Passwd string `json:"passwd"`
    PriMask int32 `json:"primask"`
}

type TokenInfo struct{
    Id int32 `json:"id"`
    Token string `json:"token"`
    Key string `json:"key"`
    Status int32 `json:"retval"`
    ErrInfo string `json:"errinfo"`
}

func NewToken()*TokenInfo{
    token:=&TokenInfo{Id:-1,Token:"nil",Key:"nil",Status:-1,ErrInfo:"Error Parameter"}
    return token
}

type EncDataReq struct{
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
	Status int32 `json:"status"`
	ErrInfo string `json:"errinfo"`
	Uuid string `json:"uuid"`
	LocalKey string `json:"locakkey"`
}

func NewDataAck() *EncDataAck{
	eda:=&EncDataAck{
		Status:-1,
		ErrInfo:"Invalid parameter",
		Uuid:"nil",
		LocalKey:"nil"	}
	return eda
}
