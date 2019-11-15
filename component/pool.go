package component

import (
	"sync"
)

// 协程池
type GoroutinePool struct {
	Queue          chan interface{} // 队列池
	Number         int              // 协程数
	Total          int              // 数据总数
	Worker         func(obj ...interface{}) bool
	finishCallback func() // 执行完成回调方法
	wait           sync.WaitGroup
}

// 创建协程池
func NewPool(number int, worker func(obj ...interface{}) bool) *GoroutinePool {
	return &GoroutinePool{
		Number: number,
		Worker: worker,
		wait:   sync.WaitGroup{},
	}
}

// 开始执行
func (g *GoroutinePool) Start() {
	for i := 0; i < g.Number; i++ {
		g.wait.Add(1)
		// 将i作为变量传入go程闭包，防止变量共享问题
		go func(index int) {
			isDone := true
			for isDone {
				select {
				case task := <-g.Queue:
					g.Worker(task, i)
				default:
					isDone = false
				}
			}
			g.wait.Done()
		}(i)
	}
	g.wait.Wait()
	if g.finishCallback != nil {
		g.finishCallback()
	}
	g.Stop()
}

// 关闭
func (g *GoroutinePool) Stop() {
	close(g.Queue)
}

// 添加task任务到队列池
func (g *GoroutinePool) AddTaskInterface(task []interface{}) {
	total := len(task)
	g.Total = total
	g.Queue = make(chan interface{}, total)
	for _, v := range task {
		g.Queue <- v
	}
}

// 设置完成时候的回调方法
func (g *GoroutinePool) FinishCallBack(callback func()) {
	g.finishCallback = callback
}
