package uds_client

import (
	"context"
	"fmt"
	"gitee.com/lovewonyoung/tp_driver/driver"
	"gitee.com/lovewonyoung/tp_driver/isotp"
	"log"
	"time"
)

// UDSClient 是一个高级客户端，封装了所有初始化和通信的复杂性
type UDSClient struct {
	stack   *isotp.Transport
	adapter *driver.ToomossAdapter
	cancel  context.CancelFunc // 用于控制所有后台goroutine的生命周期
}

// NewUDSClient 是新的构造函数，负责完成所有组件的初始化和连接。
// 它接收一个CAN驱动实例和ISOTP地址配置。
func NewUDSClient(dev driver.CANDriver, addr *isotp.Address) (*UDSClient, error) {
	// 1. 初始化适配器并启动硬件驱动
	adapter, err := driver.NewToomossAdapter(dev)
	if err != nil {
		return nil, fmt.Errorf("无法创建Toomoss适配器: %w", err)
	}

	// 2. 初始化ISOTP协议栈
	stack := isotp.NewTransport(addr)

	// 3. 创建用于goroutine生命周期管理的context
	ctx, cancel := context.WithCancel(context.Background())

	// 4. 创建内部通信channels，作为协议栈和适配器之间的桥梁
	rxFromAdapter := make(chan isotp.CanMessage, 100)
	txToAdapter := make(chan isotp.CanMessage, 100)

	// 5. 启动所有必要的后台goroutines ("粘合"逻辑)
	// a. 从适配器接收数据，送入协议栈
	go func() {
		for {
			select {
			case <-ctx.Done():
				return // 接收到退出信号
			default:
				if msg, ok := adapter.RxFunc(); ok {
					rxFromAdapter <- msg
				} else {
					time.Sleep(1 * time.Millisecond) // 避免CPU空转
				}
			}
		}
	}()

	// b. 从协议栈获取待发送数据，通过适配器发送
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-txToAdapter:
				adapter.TxFunc(msg)
			}
		}
	}()

	// c. 驱动协议栈核心状态机
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				stack.Process(rxFromAdapter, txToAdapter)
				time.Sleep(1 * time.Millisecond)
			}
		}
	}()

	log.Println("UDS客户端已成功初始化并启动。")
	client := &UDSClient{
		stack:   stack,
		adapter: adapter,
		cancel:  cancel,
	}
	return client, nil
}

// SendAndRecv 发送一个请求并阻塞等待响应，内置超时处理。
// 这是您最常用的应用层函数。
func (c *UDSClient) SendAndRecv(payload []byte, timeout time.Duration) ([]byte, error) {
	// 发送前清空可能存在的旧响应
	for {
		if _, ok := c.stack.Recv(); !ok {
			break
		}
	}

	c.stack.Send(payload) // 将数据包放入发送队列

	deadline := time.NewTimer(timeout)
	defer deadline.Stop()

	for {
		select {
		case <-deadline.C:
			return nil, fmt.Errorf("等待响应超时 (%v)", timeout)
		default:
			if data, ok := c.stack.Recv(); ok {
				// 5. 检查是否为 "Response Pending" (0x7F, ServiceID, 0x78)
				if len(data) >= 3 && data[0] == 0x7F && data[1] == payload[0] && data[2] == 0x78 {
					log.Println("收到NRC 0x78 (响应等待中)，超时时间延长5秒...")

					// 【关键修正】: 重置定时器，给予ECU额外的5秒处理时间
					// 首先要安全地停止旧的定时器
					if !deadline.Stop() {
						// 如果定时器已经触发，它的通道中可能还有值，需要排空
						// 这是time.Timer的标准安全操作
						<-deadline.C
					}
					// 然后重置为一个新的5秒倒计时
					deadline.Reset(5000 * time.Millisecond)

					// 使用continue进入下一次循环，等待最终的响应
					continue
				}

				// 6. 如果不是Pending响应，那就是最终的响应（肯定或否定），直接返回
				return data, nil
			}
			time.Sleep(2 * time.Millisecond) // 短暂等待，避免抢占CPU
		}
	}
}

// SetFDMode 允许动态切换CAN FD模式。
func (c *UDSClient) SetFDMode(isFD bool) {
	c.stack.SetFDMode(isFD)
}

// Close 优雅地关闭客户端，释放所有资源。
func (c *UDSClient) Close() {
	log.Println("正在关闭UDS客户端...")
	c.cancel()        // 发送信号，停止所有后台goroutines
	c.adapter.Close() // 调用适配器的方法，关闭硬件驱动
}
