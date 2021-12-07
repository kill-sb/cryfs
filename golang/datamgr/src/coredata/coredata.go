package coredata

import (
	"net"
)

type EncryptedData struct{
    Uuid string
    Descr string
    FromType int
    FromObj string
    OwnerId int
    EncryptedKey []byte
}
const (
    INVALID=iota
    ENCODE
    DISTRIBUTE
    MOUNT
)

const (
    RAWDATA=iota
    TAG
)

type LoginInfo struct{
    Conn net.Conn
    Name string
    Id int
    Keylocalkey []byte
}

func (info *LoginInfo) Logout() error{
    return  nil
}

