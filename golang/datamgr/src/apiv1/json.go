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

