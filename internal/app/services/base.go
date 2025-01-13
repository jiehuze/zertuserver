package services

import (
	"sync"

	"zertuserver/internal/app/models"
	"zertuserver/internal/pkg/code"
)

var initOnce sync.Once
var (
	
)

func Init() {
	initOnce.Do(func() {
		
	})
}

var successInfo = models.RespInfo{
	Code: code.Success,
	Msg:  code.MsgSuccess,
}
