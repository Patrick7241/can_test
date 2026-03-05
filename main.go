package main

import (
	"fmt"
	"time"
)

/// 多节点共享总线 + 广播 + 帧结构 + 简单仲裁，先不管丢包、Watchdog、Bus Off 等高级功能，仅最简化模拟

/// 电车的电控转向优先级极高，且一般使用独立的can

// 这里用一个节点来模拟一个ecu或者设备
type Node struct {
	Name          string       // 节点名称
	Bus           *Bus         // 关联的总线 说明：实车可能包含多个can总线，这里只模拟一个总线
	SubscribedIDs map[int]bool // 订阅哪些ID 说明：can的通信方式是各个节点订阅自己需要的ID，然后通过总线接收数据
	ReceiveChan   chan Frame   // 接收数据帧
}

// 数据帧
type Frame struct {
	ID        int    // 消息ID，也表示优先级（小ID优先）说明：一般规定id为一个操作指令，优先级越小越先处理
	From      string // 发送节点
	Data      string // 消息内容
	Timestamp int64  // 发送时间，毫秒级时间戳
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
		Name:          "Node1",
		Bus:           bus,
		SubscribedIDs: map[int]bool{1: true, 2: true},
		ReceiveChan:   make(chan Frame, 100),
	}
	node2 := &Node{
		Name:          "Node2",
		Bus:           bus,
		SubscribedIDs: map[int]bool{1: true, 3: true},
		ReceiveChan:   make(chan Frame, 100),
	}
	bus.Nodes = append(bus.Nodes, node1, node2)
	go node1.Listen()
	go node2.Listen()
	node1.Send(1, "Hello World!")
	node2.Send(3, "Hi!")

	for {
		time.Sleep(time.Second)
	}

}
