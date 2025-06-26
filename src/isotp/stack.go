package isotp

import "gitee.com/lovewonyoung/tp_driver/util"

// 假设

// Transport 是ISOTP协议栈的核心结构
type Transport struct {
	address              *Address
	IsFD                 bool
	MaxDataLength        int
	rxState              State
	txState              State
	rxBuffer             []byte
	txBuffer             []byte
	rxQueue              *util.Queue
	txQueue              *util.Queue
	rxFrameLen           int
	txFrameLen           int
	rxSeqNum             int
	txSeqNum             int
	rxBlockCounter       int
	txBlockCounter       int
	remoteBlocksize      int
	lastFlowControlFrame *FlowControlFrame
	pendingFlowControlTx bool
	timerRxCF            *Timer
	timerRxFC            *Timer
	timerTxSTmin         *Timer
	Blocksize            int
	StminMs              int
	WftMax               int
	wftCounter           int
}

func NewTransport(address *Address) *Transport {
	t := &Transport{
		address:       address,
		rxQueue:       util.NewQueue(),
		txQueue:       util.NewQueue(),
		IsFD:          false,
		MaxDataLength: 8,
		timerRxCF:     NewTimer(1000),
		timerRxFC:     NewTimer(1000),
		timerTxSTmin:  NewTimer(0),
		Blocksize:     0,
		StminMs:       20,
		WftMax:        5,
	}
	t.stopReceiving()
	t.stopSending()
	return t
}

func (t *Transport) SetFDMode(isFD bool) {
	t.IsFD = isFD
	if isFD {
		t.MaxDataLength = 64
	} else {
		t.MaxDataLength = 8
	}
}

func (t *Transport) Send(data []byte) {
	t.txQueue.Put(data)
}

func (t *Transport) Recv() ([]byte, bool) {
	if t.rxQueue.Available() {
		return t.rxQueue.Get(), true
	}
	return nil, false
}

func (t *Transport) Process(rx <-chan CanMessage, tx chan<- CanMessage) {
	select {
	case msg := <-rx:
		t.ProcessRx(msg) // 调用 rx_handler.go
	default:
	}

	if msg, send := t.ProcessTx(); send { // 调用 tx_handler.go
		select {
		case tx <- msg:
		default:
		}
	}
}

// 内部辅助函数
func (t *Transport) stopReceiving() {
	t.rxState = StateIdle
	t.rxBuffer = nil
	t.rxFrameLen = 0
	t.rxSeqNum = 0
	t.rxBlockCounter = 0
	t.timerRxCF.Stop()
}

func (t *Transport) stopSending() {
	t.txState = StateIdle
	t.txBuffer = nil
	t.txFrameLen = 0
	t.txSeqNum = 0
	t.txBlockCounter = 0
	t.timerRxFC.Stop()
	t.timerTxSTmin.Stop()
}

func (t *Transport) makeTxMsg(data []byte, addrType AddressType) CanMessage {
	// 1. 获取动态计算的ID
	arbitrationID := t.address.GetTxArbitrationID(addrType)

	// 2. 附加地址前缀（如果有）
	fullPayload := append(t.address.TxPayloadPrefix, data...)

	// 这里可以加入CAN-FD的填充逻辑 (padding)

	return CanMessage{
		ArbitrationID: arbitrationID,
		Data:          fullPayload,
		IsExtendedID:  t.address.Is29Bit(),
		IsFD:          t.IsFD,
	}
}
