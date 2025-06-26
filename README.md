# go_iso15765

🚗 **go_iso15765** 是一个用 Go 语言实现的 ISO 15765‑2（ISO‑TP）传输层协议库，可在 CAN 总线上进行 UDS 诊断通信。

该项目灵感来自 [python-can-isotp (v2.x)](https://github.com/pylessard/python-can-isotp/tree/v2.x/isotp)，并基于 [go-uds](https://github.com/andrewarrow/go-uds) 构建。

---

## ✨ 功能特性

- 完整实现 ISO 15765‑2 传输层  
- 支持标准帧 (11‑bit) 与扩展帧 (29‑bit) CAN ID  
- Single Frame、First + Consecutive Frames、Flow Control 全流程支持  
- 多帧分段发送与重组  
- 可配置 Block Size、STmin、帧填充 (Padding)  
- 与 UDS (ISO 14229) 协议兼容，对接上层诊断逻辑  

---

## 📦 安装

```bash
go get github.com/yourusername/go_iso15765
```

> **提示**：推荐使用 Go Modules（Go 1.18+）。如仍在使用 GOPATH，请确认已正确设置环境。


---

## 🛠 依赖

- [go-uds](https://github.com/andrewarrow/go-uds) — UDS 诊断层
- 能发送/接收 CAN 帧的接口（如 SocketCAN、PCAN、Canalyst 等）

---

## 📚 参考资料

- ISO 15765‑2: Road vehicles — Diagnostic communication over Controller Area Network (DoCAN) — Part 2: Transport protocol  
- ISO 14229 (UDS)  
- [python-can-isotp](https://github.com/pylessard/python-can-isotp)  
- [go-uds](https://github.com/andrewarrow/go-uds)

---

## 🤝 贡献

欢迎提出 Issue 或 PR！在提交重大变更前，请先开启讨论。

---

## 📜 License

MIT
