package can

import (
	"container/heap"
	"fmt"
	"sync"
	"time"
)

/*
========================================
优先级队列丢帧策略
========================================

当缓冲区满时，丢弃低优先级帧而不是最新帧
*/

// PriorityFrame 带优先级的帧
type PriorityFrame struct {
	frame    CANFrame
	priority int // 优先级数值越小越重要
	index    int // 堆中的索引
}

// PriorityQueue 优先级队列
type PriorityQueue []*PriorityFrame

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// 优先级高的（数值小的）排前面
	return pq[i].priority < pq[j].priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*PriorityFrame)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*pq = old[0 : n-1]
	return item
}

// PriorityCANBus 基于优先级的CAN总线
type PriorityCANBus struct {
	name          string
	priorityQueue PriorityQueue
	maxSize       int
	listeners     map[uint32][]CANListener
	running       bool
	stats         CANBusStats
	mu            sync.RWMutex
}

// NewPriorityCANBus 创建优先级总线
func NewPriorityCANBus(name string, maxSize int) *PriorityCANBus {
	pq := make(PriorityQueue, 0, maxSize)
	heap.Init(&pq)

	return &PriorityCANBus{
		name:          name,
		priorityQueue: pq,
		maxSize:       maxSize,
		listeners:     make(map[uint32][]CANListener),
		stats: CANBusStats{
			FramesByID: make(map[uint32]uint64),
		},
	}
}

// SendFrameWithPriority 带优先级发送帧
func (pbus *PriorityCANBus) SendFrameWithPriority(frame CANFrame, priority int) error {
	pbus.mu.Lock()
	defer pbus.mu.Unlock()

	frame.Timestamp = time.Now()

	// 如果队列满了
	if pbus.priorityQueue.Len() >= pbus.maxSize {
		// 查看队列中优先级最低的帧
		lowestPriorityFrame := pbus.priorityQueue[pbus.priorityQueue.Len()-1]

		if priority < lowestPriorityFrame.priority {
			// 新帧优先级更高，丢弃旧的低优先级帧
			_ = heap.Pop(&pbus.priorityQueue)

			heap.Push(&pbus.priorityQueue, &PriorityFrame{
				frame:    frame,
				priority: priority,
			})

			pbus.stats.DroppedFrames++ // 丢弃低优先级帧
			fmt.Printf("🔄 丢弃低优先级帧 ID:0x%03X (优先级:%d), 接受高优先级帧 ID:0x%03X (优先级:%d)\n",
				lowestPriorityFrame.frame.ID, lowestPriorityFrame.priority,
				frame.ID, priority)
		} else {
			// 新帧优先级更低，直接丢弃
			pbus.stats.DroppedFrames++
			return fmt.Errorf("新帧优先级过低，已丢弃")
		}
	} else {
		// 队列未满，直接加入
		heap.Push(&pbus.priorityQueue, &PriorityFrame{
			frame:    frame,
			priority: priority,
		})
	}

	pbus.stats.TotalFrames++
	pbus.stats.FramesByID[frame.ID]++

	return nil
}

// ProcessFrames 处理帧（按优先级）
func (pbus *PriorityCANBus) ProcessFrames() {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for pbus.running {
		<-ticker.C

		pbus.mu.Lock()
		if pbus.priorityQueue.Len() > 0 {
			// 取出优先级最高的帧
			pf := heap.Pop(&pbus.priorityQueue).(*PriorityFrame)
			frame := pf.frame

			// 分发给监听器
			listeners, exists := pbus.listeners[frame.ID]
			if exists {
				for _, listener := range listeners {
					go listener.OnCANFrame(frame)
				}
			}
		}
		pbus.mu.Unlock()
	}
}

// 定义CAN ID的优先级映射
var CANIDPriority = map[uint32]int{
	0xFD:  1, // 车速 - 高优先级
	0xB5:  2, // 档位 - 高优先级
	0x3EB: 3, // 刹车油门 - 中优先级
	0x3DA: 4, // 转向 - 中优先级
	0x101: 5, // 加速度 - 低优先级
	0x5E1: 6, // 温度 - 最低优先级（不影响驾驶）
}

// GetPriority 获取CAN ID的优先级
func GetPriority(canID uint32) int {
	if priority, exists := CANIDPriority[canID]; exists {
		return priority
	}
	return 99 // 未知ID的默认优先级（最低）
}
