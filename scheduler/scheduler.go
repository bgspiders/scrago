package scheduler

import (
	"container/heap"
	"scrago/request"
	"sync"
	"sync/atomic"
)

// Scheduler 调度器接口
type Scheduler interface {
	Enqueue(req *request.Request)
	Dequeue() *request.Request
	Empty() bool
	Size() int
}

// FIFOScheduler FIFO调度器
type FIFOScheduler struct {
	queue []*request.Request
	mutex sync.RWMutex
}

// NewFIFOScheduler 创建FIFO调度器
func NewFIFOScheduler() *FIFOScheduler {
	return &FIFOScheduler{
		queue: make([]*request.Request, 0),
	}
}

// Enqueue 入队
func (s *FIFOScheduler) Enqueue(req *request.Request) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.queue = append(s.queue, req)
}

// Dequeue 出队
func (s *FIFOScheduler) Dequeue() *request.Request {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if len(s.queue) == 0 {
		return nil
	}
	
	req := s.queue[0]
	s.queue = s.queue[1:]
	return req
}

// Empty 检查是否为空
func (s *FIFOScheduler) Empty() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.queue) == 0
}

// Size 获取队列大小
func (s *FIFOScheduler) Size() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.queue)
}

// PriorityScheduler 优先级调度器
type PriorityScheduler struct {
	queue PriorityQueue
	mutex sync.RWMutex
}

// PriorityQueue 优先级队列
type PriorityQueue []*PriorityItem

// PriorityItem 优先级项目
type PriorityItem struct {
	Request  *request.Request
	Priority int
	Index    int
}

// NewPriorityScheduler 创建优先级调度器
func NewPriorityScheduler() *PriorityScheduler {
	pq := make(PriorityQueue, 0)
	heap.Init(&pq)
	
	return &PriorityScheduler{
		queue: pq,
	}
}

// Enqueue 入队
func (s *PriorityScheduler) Enqueue(req *request.Request) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	item := &PriorityItem{
		Request:  req,
		Priority: req.Priority,
	}
	
	heap.Push(&s.queue, item)
}

// Dequeue 出队
func (s *PriorityScheduler) Dequeue() *request.Request {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if s.queue.Len() == 0 {
		return nil
	}
	
	item := heap.Pop(&s.queue).(*PriorityItem)
	return item.Request
}

// Empty 检查是否为空
func (s *PriorityScheduler) Empty() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.queue.Len() == 0
}

// Size 获取队列大小
func (s *PriorityScheduler) Size() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.queue.Len()
}

// 实现heap.Interface接口

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// 优先级越高，越先执行（数值越大优先级越高）
	return pq[i].Priority > pq[j].Priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = i
	pq[j].Index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*PriorityItem)
	item.Index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.Index = -1
	*pq = old[0 : n-1]
	return item
}

// LIFOScheduler LIFO调度器（栈）
type LIFOScheduler struct {
	stack []*request.Request
	mutex sync.RWMutex
}

// NewLIFOScheduler 创建LIFO调度器
func NewLIFOScheduler() *LIFOScheduler {
	return &LIFOScheduler{
		stack: make([]*request.Request, 0),
	}
}

// Enqueue 入栈
func (s *LIFOScheduler) Enqueue(req *request.Request) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.stack = append(s.stack, req)
}

// Dequeue 出栈
func (s *LIFOScheduler) Dequeue() *request.Request {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if len(s.stack) == 0 {
		return nil
	}
	
	req := s.stack[len(s.stack)-1]
	s.stack = s.stack[:len(s.stack)-1]
	return req
}

// Empty 检查是否为空
func (s *LIFOScheduler) Empty() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.stack) == 0
}

// Size 获取栈大小
func (s *LIFOScheduler) Size() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return len(s.stack)
}

// ChannelScheduler 高性能channel调度器
type ChannelScheduler struct {
	requestChan chan *request.Request
	size        int64
}

// NewChannelScheduler 创建channel调度器
func NewChannelScheduler(bufferSize int) *ChannelScheduler {
	return &ChannelScheduler{
		requestChan: make(chan *request.Request, bufferSize),
		size:        0,
	}
}

// Enqueue 入队（非阻塞）
func (s *ChannelScheduler) Enqueue(req *request.Request) {
	select {
	case s.requestChan <- req:
		atomic.AddInt64(&s.size, 1)
	default:
		// 队列满时，扩容处理
		go func() {
			s.requestChan <- req
			atomic.AddInt64(&s.size, 1)
		}()
	}
}

// Dequeue 出队
func (s *ChannelScheduler) Dequeue() *request.Request {
	select {
	case req := <-s.requestChan:
		atomic.AddInt64(&s.size, -1)
		return req
	default:
		return nil
	}
}

// Empty 检查是否为空
func (s *ChannelScheduler) Empty() bool {
	return atomic.LoadInt64(&s.size) == 0
}

// Size 获取队列大小
func (s *ChannelScheduler) Size() int {
	return int(atomic.LoadInt64(&s.size))
}