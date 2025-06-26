/*
 * @Author: LoveWonYoung leeseoimnida@gmail.com
 * @Date: 2025-05-19 16:01:40
 * @LastEditors: LoveWonYoung leeseoimnida@gmail.com
 * @LastEditTime: 2025-06-26 11:09:14
 * @FilePath: \src\main.go
 * @Description: 这是默认设置,请设置`customMade`, 打开koroFileHeader查看配置 进行设置: https://github.com/OBKoro1/koro1FileHeader/wiki/%E9%85%8D%E7%BD%AE
 */
package main

import (
	"gitee.com/lovewonyoung/tp_driver/driver"
	"gitee.com/lovewonyoung/tp_driver/isotp"
	"gitee.com/lovewonyoung/tp_driver/uds_client"
	"log"
	"time"
)

const (
	UDS_REQUEST_ID  = 0x701
	UDS_RESPONSE_ID = 0x702
)

func init() {
	//log_recorder.MakeDir()
	//log_recorder.RecorderAsNameInit("new tp test" + log_recorder.NowString())
}

func main() {
	log.Println("--- 启动UDS诊断应用程序 (集成客户端版) ---")
	var fdMode = false
	var canDevice driver.CANDriver
	if fdMode {
		// canDevice = driver.yourdevices()
		log.Println("当前模式: CAN FD")
	} else {
		// canDevice = driver.yourdevices()
		log.Println("当前模式: CAN")
	}
	// --- 步骤 1: 选择驱动并定义地址 ---
	// 这里我们演示CAN FD模式，只需替换驱动和ID即可

	addr, _ := isotp.NewAddress(
		isotp.Normal11Bit,               // 寻址模式
		isotp.WithTxID(UDS_REQUEST_ID),  // 使用您CAN FD日志中的发送ID
		isotp.WithRxID(UDS_RESPONSE_ID), // 使用您CAN FD日志中的响应ID
	)

	client, _ := uds_client.NewUDSClient(canDevice, addr)

	defer client.Close() // 保证程序退出时资源能被释放
	client.SetFDMode(fdMode)

	// --- 步骤 3: 执行诊断任务 ---
	udsRequest := []byte{0x22, 0xF0, 0xFA}
	client.SendAndRecv(udsRequest, 2*time.Second)
	client.SendAndRecv([]byte{0x22, 0xf1, 0x89}, 2*time.Second)
	client.SendAndRecv([]byte{0x22, 0xf0, 0x89}, 2*time.Second)
	client.SendAndRecv([]byte{0x22, 0xf1, 0x93}, 2*time.Second)
	client.SendAndRecv([]byte{0x22, 0xf1, 0x95}, 2*time.Second)
	client.SendAndRecv([]byte{0x10, 0x03}, 2*time.Second)
	client.SendAndRecv([]byte{0x10, 0x83}, 0)
	client.SendAndRecv([]byte{0x85, 0x82}, 0)
	client.SendAndRecv([]byte{0x28, 0x83, 0x03}, 0)
	client.SendAndRecv([]byte{0x31, 0x01, 0x2, 0x03}, 2000*time.Millisecond)
	client.SendAndRecv([]byte{0x10, 2}, 2000*time.Millisecond)
}
