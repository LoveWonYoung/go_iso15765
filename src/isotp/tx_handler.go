package isotp

import "fmt"

// ProcessTx 处理发送状态机，返回一个待发送的报文（如果需要）
func (t *Transport) ProcessTx() (CanMessage, bool) {
	// 1. 处理流控帧发送请求
	if t.pendingFlowControlTx {
		t.pendingFlowControlTx = false
		// 发送一个FC(CTS)
		payload := createFlowControlPayload(FlowStatusContinueToSend, t.Blocksize, t.StminMs)
		return t.makeTxMsg(payload, Physical), true
	}

	// 2. 处理收到的流控帧
	if t.lastFlowControlFrame != nil {
		fc := t.lastFlowControlFrame
		t.lastFlowControlFrame = nil //
		t.handleTxFlowControl(fc)
	}

	// 3. 检查FC超时
	if t.timerRxFC.IsTimedOut() {
		fmt.Println("等待流控帧超时，停止发送。")
		t.stopSending()
	}

	// 4. 根据状态处理数据发送
	switch t.txState {
	case StateIdle:
		return t.handleTxIdle()
	case StateWaitFC:
		// 等待FC中，无事可做
	case StateTransmit:
		return t.handleTxTransmit()
	}

	return CanMessage{}, false
}

func (t *Transport) handleTxFlowControl(fc *FlowControlFrame) {
	if t.txState == StateIdle {
		fmt.Println("警告：在IDLE状态下收到了流控帧。")
		return
	}

	t.timerRxFC.Stop() // 收到任何有效FC都应停止超时

	switch fc.FlowStatus {
	case FlowStatusContinueToSend:
		t.wftCounter = 0
		t.remoteBlocksize = fc.BlockSize
		t.timerTxSTmin.SetTimeout(int(fc.STmin.Milliseconds()))
		t.txState = StateTransmit
		t.txBlockCounter = 0
		t.timerTxSTmin.Start() // 立即开始计时，准备发送第一帧

	case FlowStatusWait:
		t.wftCounter++
		if t.WftMax > 0 && t.wftCounter > t.WftMax {
			fmt.Println("错误：等待帧(Wait Frame)数量超出最大限制。")
			t.stopSending()
		} else {
			t.timerRxFC.Start() // 重新开始等待FC
		}

	case FlowStatusOverflow:
		fmt.Println("错误：对方缓冲区溢出，停止发送。")
		t.stopSending()
	}
}

func (t *Transport) handleTxIdle() (CanMessage, bool) {
	if !t.txQueue.Available() {
		return CanMessage{}, false // 无事可做
	}
	payload := t.txQueue.Get()
	t.txFrameLen = len(payload)

	// 判断是单帧还是多帧
	sfPciSize := 1
	if t.txFrameLen > 7 {
		sfPciSize = 2
	}
	//if t.txFrameLen+sfPciSize <= t.MaxDataLength {
	//	// 作为单帧发送
	//	data, _ := createSingleFramePayload(payload, t.MaxDataLength)
	//	// --- 修改点: 传递地址类型 ---
	//	return t.makeTxMsg(data, Physical), true
	//} else {
	//	// ... (FF creation logic is the same) ...
	//	// --- 修改点: 传递地址类型 ---
	//	return t.makeTxMsg(data, Physical), true
	//}
	if t.txFrameLen+sfPciSize <= t.MaxDataLength {
		// 作为单帧发送
		data, _ := createSingleFramePayload(payload, t.MaxDataLength)
		return t.makeTxMsg(data, Physical), true
	} else {
		// 作为多帧发送，先发送首帧
		ffPciSize := 2
		if t.txFrameLen > 4095 {
			ffPciSize = 6
		}
		chunkSize := t.MaxDataLength - ffPciSize

		t.txBuffer = payload[chunkSize:]
		data, _ := createFirstFramePayload(payload[:chunkSize], t.txFrameLen, t.MaxDataLength)

		t.txSeqNum = 1
		t.txState = StateWaitFC
		t.timerRxFC.Start() // 开始等待FC

		return t.makeTxMsg(data, Physical), true
	}
}

func (t *Transport) handleTxTransmit() (CanMessage, bool) {
	if len(t.txBuffer) == 0 {
		fmt.Println("多帧数据发送完成。")
		t.stopSending()
		return CanMessage{}, false
	}

	if !t.timerTxSTmin.IsTimedOut() {
		return CanMessage{}, false // 还未到发送时间
	}

	chunkSize := t.MaxDataLength - 1 // CF PCI=1
	var chunk []byte
	if len(t.txBuffer) > chunkSize {
		chunk = t.txBuffer[:chunkSize]
		t.txBuffer = t.txBuffer[chunkSize:]
	} else {
		chunk = t.txBuffer
		t.txBuffer = nil
	}

	data, _ := createConsecutiveFramePayload(chunk, t.txSeqNum)
	t.txSeqNum = (t.txSeqNum + 1) % 16

	t.timerTxSTmin.Start() // 重置发送间隔定时器
	t.txBlockCounter++

	// 检查是否达到一个块的大小
	if t.remoteBlocksize > 0 && t.txBlockCounter >= t.remoteBlocksize {
		t.txState = StateWaitFC
		t.timerRxFC.Start() // 等待下一个FC
	}

	return t.makeTxMsg(data, Physical), true
}
