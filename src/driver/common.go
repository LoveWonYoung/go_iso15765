package driver

import (
	"context"
)

// UnifiedCANMessage 是一个通用的CAN/CAN-FD消息结构体，用于在channel中传递。
// 它屏蔽了底层 CAN_MSG 和 CANFD_MSG 的差异。
type UnifiedCANMessage struct {
	ID   uint32
	DLC  byte
	Data [64]byte // 使用64字节以兼容CAN-FD
	IsFD bool     // 标志位，用于区分是CAN还是CAN-FD消息
}


type CANDriver interface {
	Init() error
	Start()
	Stop()
	Write(id int32, data []byte) error
	RxChan() <-chan UnifiedCANMessage
	Context() context.Context
}

type ReadFuncFd func(id int32) ([]byte, error)

type WriteFuncFd func(id int32, data []byte) error

type Devices interface {
	Init() bool
	Read(id int32) ([8]byte, error)
	Write(id int32, data [8]byte) error
}
