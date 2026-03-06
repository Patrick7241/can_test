package main

import (
	"fmt"
	"time"
)

/*
========================================
CAN总线模拟器
========================================

本项目模拟CAN总线通信机制，包括：
- 多节点共享总线
- 广播通信
- 基于ID的优先级仲裁
- 冲突检测与重发

车辆信号数据库:
- 大众ID.4 (2020+): vehicles/volkswagen_id4/
- 更多车型请查看 vehicles/ 目录

详细信号定义和使用请查看各车型目录下的README.md
*/

/// 多节点共享总线 + 广播 + 帧结构 + 简单仲裁，先不管丢包、Watchdog、Bus Off 等高级功能，仅最简化模拟

/// 电车的电控转向优先级极高，且一般使用独立的can

/// 低优先级的帧一直仲裁失败-》优先级饥饿  解决办法：错误计数，限流高频率发送高优先级的节点，优先级提升，重试队列

// 这里用一个节点来模拟一个ecu或者设备

type Node struct {
	Name          string       // 节点名称
	Bus           *Bus         // 关联的总线 说明：实车可能包含多个can总线，这里只模拟一个总线
	SubscribedIDs map[int]bool // 订阅哪些ID 说明：can的通信方式是各个节点订阅自己需要的ID，然后通过总线接收数据
	ReceiveChan   chan Frame   // 接收数据帧
}

// 数据帧
type Frame struct {
	// 消息ID，也表示优先级（小ID优先）说明：一般规定id为一个操作指令（所以id一般不会改变），优先级越小越先处理
	// 十六进制 可以进行位仲裁，效率更高，电控系统，通信效率就是生命
	ID        int
	From      string // 发送节点
	Data      string // 消息内容
	Timestamp int64  // 发送时间，毫秒级时间戳
	FailCount int    // 错误计数
}

// 总线，即一个can
type Bus struct {
	Nodes       []*Node    // 挂在总线的节点
	SendQueue   chan Frame // 节点发送的帧放这里
	BroadcastCh chan Frame // 广播给所有节点，结合send队列设置时间窗口和优先放行id最小（优先级最高）的帧
}

// 节点处理监听到的数据帧
func (n *Node) Listen() {
	for frame := range n.ReceiveChan {
		if n.SubscribedIDs[frame.ID] {
			fmt.Printf("[%s] 接收到消息: %v\n", n.Name, frame)
		}
	}
}

// 节点发送数据帧
func (n *Node) Send(id int, data string) {
	frame := Frame{
		ID:        id,
		From:      n.Name,
		Data:      data,
		Timestamp: time.Now().UnixMilli(),
	}
	n.Bus.SendQueue <- frame
}

// 总线处理数据帧
func (bus *Bus) Start() {
	go func() {
		// 缓存窗口，用于收集同时发来的帧
		var buffer []Frame

		// 仲裁的时间窗口
		ticker := time.NewTicker(5 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case frame := <-bus.SendQueue:
				buffer = append(buffer, frame)

			case <-ticker.C:
				if len(buffer) == 0 {
					continue
				}

				// 仲裁：按 ID 排序，ID 小优先
				minIdx := 0
				for i := 1; i < len(buffer); i++ {
					if buffer[i].ID < buffer[minIdx].ID {
						minIdx = i
					}
				}
				// 赢的帧
				win := buffer[minIdx]

				fmt.Printf("[总线] 仲裁结果: %v\n", win)

				// 广播
				for _, node := range bus.Nodes {
					select {
					case node.ReceiveChan <- win:
					default:
					}
				}

				// buffer 中其他帧留到下一轮，模拟 CAN 失败的节点重新发送
				var newBuffer []Frame
				for i, f := range buffer {
					if i != minIdx {
						f.FailCount++

						// 一般不使用该方式，因为id一般固定标识某一种功能，不能随便改变
						//// 失败超过三次，提升优先级
						//if f.FailCount >= 3 && f.ID > 0 {
						//	f.ID -= 1
						//}

						newBuffer = append(newBuffer, f)
					}
				}
				buffer = newBuffer
			}
		}
	}()
}

func main() {
	bus := &Bus{
		Nodes:       []*Node{},
		SendQueue:   make(chan Frame, 100),
		BroadcastCh: make(chan Frame, 100),
	}
	bus.Start()
	node1 := &Node{
		Name:          "这是高优先级指令，且频率较高", // 高优先级帧必须周期发送，但周期不能太小，周期可以设置最大周期和最小周期，动态调整
		Bus:           bus,
		SubscribedIDs: map[int]bool{1: true, 2: true},
		ReceiveChan:   make(chan Frame, 100),
	}
	node2 := &Node{
		Name:          "这是低优先级指令",
		Bus:           bus,
		SubscribedIDs: map[int]bool{1: true, 3: true},
		ReceiveChan:   make(chan Frame, 100),
	}
	bus.Nodes = append(bus.Nodes, node1, node2)
	go node1.Listen()
	go node2.Listen()
	go func() {
		for {
			node1.Send(1, "转向")
			time.Sleep(5 * time.Millisecond) // 如果高优先级帧不周期发送，低优先级帧无法进入队列
		}
	}()
	node2.Send(3, "空调温度降低")

	for {
		time.Sleep(time.Second)
	}

}
