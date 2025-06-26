package isotp

import "fmt"

// ProcessRx 处理接收到的单个CAN报文
func (t *Transport) ProcessRx(msg CanMessage) {
	// 0. 检查ID是否匹配 (这里简化)
	// if msg.ArbitrationID != t.address.RxID { return }
	if !t.address.IsForMe(&msg) {
		return // 这不是给我们的报文
	}
	// 1. 检查CF超时
	if t.timerRxCF.IsTimedOut() {
		fmt.Println("接收连续帧超时，重置接收状态。")
		t.stopReceiving()
	}

	// 2. 解析报文
	frame, err := ParseFrame(&msg, t.address.RxPrefixSize)
	if err != nil {
		fmt.Println("报文解析失败:", err)
		return
	}

	// 3. 根据帧类型处理状态机
	switch f := frame.(type) {
	case *FlowControlFrame:
		t.lastFlowControlFrame = f
		if t.rxState == StateWaitCF { // Python代码中的 t.rx_state == WAIT
			if f.FlowStatus == FlowStatusWait || f.FlowStatus == FlowStatusContinueToSend {
				t.timerRxCF.Start()
			}
		}
		return // 流控帧不改变接收状态机，只影响发送状态机

	case *SingleFrame:
		t.handleRxSingleFrame(f)

	case *FirstFrame:
		t.handleRxFirstFrame(f)

	case *ConsecutiveFrame:
		t.handleRxConsecutiveFrame(f)
	}
}

func (t *Transport) handleRxSingleFrame(f *SingleFrame) {
	if t.rxState != StateIdle {
		fmt.Println("警告：在多帧接收过程中被一个新单帧打断。")
	}
	t.stopReceiving()
	t.rxQueue.Put(f.Data)
	//fmt.Println("成功接收一个单帧。")
}

func (t *Transport) handleRxFirstFrame(f *FirstFrame) {
	if t.rxState != StateIdle {
		fmt.Println("警告：在多帧接收过程中被一个新首帧打断。")
	}
	t.stopReceiving()

	// 缓冲区检查等逻辑可以加在这里
	t.rxFrameLen = f.TotalSize
	t.rxBuffer = append(t.rxBuffer, f.Data...)

	if len(t.rxBuffer) >= t.rxFrameLen {
		// FF包含了所有数据 (实际上不可能，除非FF的payload非常小)
		t.rxQueue.Put(t.rxBuffer)
		t.stopReceiving()
	} else {
		// 准备接收连续帧，并请求发送流控
		t.rxState = StateWaitCF
		t.rxSeqNum = 1
		t.pendingFlowControlTx = true // 请求发送一个FC(CTS)
	}
}

func (t *Transport) handleRxConsecutiveFrame(f *ConsecutiveFrame) {
	if t.rxState != StateWaitCF {
		fmt.Println("错误：在非等待CF状态下收到了连续帧。")
		t.stopReceiving()
		return
	}

	if f.SequenceNumber != t.rxSeqNum {
		fmt.Printf("错误：序列号不匹配。期望: %d,收到: %d\n", t.rxSeqNum, f.SequenceNumber)
		t.stopReceiving()
		return
	}

	t.timerRxCF.Start() // 重置超时定时器
	t.rxSeqNum = (t.rxSeqNum + 1) % 16

	bytesToReceive := t.rxFrameLen - len(t.rxBuffer)
	if len(f.Data) > bytesToReceive {
		t.rxBuffer = append(t.rxBuffer, f.Data[:bytesToReceive]...)
	} else {
		t.rxBuffer = append(t.rxBuffer, f.Data...)
	}

	if len(t.rxBuffer) >= t.rxFrameLen {
		// 接收完成
		t.rxQueue.Put(append([]byte{}, t.rxBuffer...)) // 复制一份数据放入队列
		fmt.Println("多帧数据接收完成。")
		t.stopReceiving()
	} else {
		// 继续接收
		t.rxBlockCounter++
		if t.Blocksize > 0 && t.rxBlockCounter >= t.Blocksize {
			t.rxBlockCounter = 0
			t.pendingFlowControlTx = true // 一个块接收完成，请求发送FC
			t.timerRxCF.Stop()            // 等待FC发送后再重新启动
		}
	}
}
