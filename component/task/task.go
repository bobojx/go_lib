package task

import (
	"go_lib/component"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
)

// 定义一个枚举类型
type timeType int

// 枚举的值，默认从0开始
const (
	Second timeType = iota
	Minute
	Hour
	Day
	Week
	Month
)

// 任务规则
type Rule struct {
	Row   string   // 原始值
	Type  timeType // 时间类型
	Loop  bool     // 是否循环
	Value int      // 循环值
}

// 任务选项
type Item struct {
	RuleList     []*Rule               // 规则列表
	ExecFunc     func(item *Item) bool // 执行方法
	CallBackFunc func(item *Item)      // 回调函数
	LastExecTime time.Time             // 最后执行时间
	Args         []interface{}         // 任务参数
}

// 任务管理
type Manage struct {
	stop bool       // 是否停止
	list []*Item    // 任务列表
	lock sync.Mutex // 任务锁
}

// 实例化定时任务管理
func NewManage() *Manage {
	return &Manage{
		stop: false,
		list: nil,
		lock: sync.Mutex{},
	}
}

// 添加任务
func (m *Manage) Add(item *Item) {
	m.lock.Lock()
	m.list = append(m.list, item)
	m.lock.Unlock()
}

// 移除任务
func (m *Manage) Remove(item *Item) {
	m.lock.Lock()
	for i, v := range m.list {
		if v == item {
			m.list = append(m.list[:i], m.list[i+1:]...)
			break
		}
	}
	m.lock.Unlock()
}

// 清除任务列表
func (m *Manage) Clear() {
	m.lock.Lock()
	m.list = nil
	m.lock.Unlock()
}

// 添加任务
// timeStr 定时任务规则：
// 采用time.Sleep方式的规则为：* * * * * * (秒 分 时 日 周 月)
// 每小时执行：0 0 */1 * * *
// 每月10号执行：0 0 0 10 * *
// 每周三执行： 0 0 0 0 */2 *，其中0~6表示星期一到星期日
// 每年7月份执行：0 0 0 0 0 */7，其中1~12表示一月到十二月
//
// 采用time.Tick方式的规则为：* * * * * * (秒 分 时 日 月 年)
// 每小时执行：0 0 */1 * * *
// 每月10号执行：0 0 0 10 * *
func (m *Manage) AddTask(timeStr string, exec func(item *Item) bool, callback func(item *Item), args ...interface{}) {
	timeList := strings.Split(timeStr, " ")
	if len(timeList) < 6 {
		log.Fatalln("timeStr format error，should be:* * * * * *")
		return
	}
	var rules []*Rule
	// 将任务规则拆分为具体的时间类型
	rules = append(rules, m.explainRule(timeList[0], Second))
	rules = append(rules, m.explainRule(timeList[1], Minute))
	rules = append(rules, m.explainRule(timeList[2], Hour))
	rules = append(rules, m.explainRule(timeList[3], Day))
	rules = append(rules, m.explainRule(timeList[4], Week))
	rules = append(rules, m.explainRule(timeList[5], Month))

	item := &Item{
		RuleList:     rules,
		ExecFunc:     exec,
		CallBackFunc: callback,
		Args:         args,
	}
	m.Add(item)
}

// 解析定时任务规则
func (m *Manage) explainRule(str string, tt timeType) *Rule {
	if str == "*" {
		return nil
	}
	rule := &Rule{
		Row:  str,
		Type: tt,
	}

	// 是否循环执行任务，例如：0 0 */1 * * * （每小时执行）
	val := strings.Split(str, "/")
	var err error
	tmpVal := val[0]
	if len(val) > 1 {
		rule.Loop = true
		tmpVal = val[1]
	}
	rule.Value, err = strconv.Atoi(tmpVal)
	if err != nil {
		log.Fatalln(err)
		return nil
	}
	return rule
}

// 采用time.Sleep的方式执行定时任务
func (m *Manage) run2Sleep() {
	for {
		if m.stop {
			break
		}
		go m.execute(time.Unix(time.Now().Unix(), 0))
		time.Sleep(time.Second - time.Duration(time.Now().Nanosecond()))
	}
}

// 采用time.Tick的方式执行任务，区别在于使用channel阻塞协程完成定时任务比较灵活，
// 可以结合select设置超时时间以及默认执行方法，而且可以设置timer的主动关闭，
// 以及不需要每次都生成一个timer(这方面节省系统内存，垃圾收回也需要时间)。
// 如果对于不确定时长的定时和任务，则需要每次执行完成之后重新计算下次执行时间，也就是要重新生成一个timer
func (m *Manage) run2Tick() {
	nowTime := time.Unix(time.Now().Unix(), 0)
	for _, item := range m.list {
		// 计算任务时长
		timeVal, isLoop := m.calculationTime(item.RuleList, nowTime)
		if timeVal < 0 {
			log.Println("timeStr format error，cannot be set to the past time")
			return
		}
		go m.doTick(timeVal, isLoop, item)
	}
	time.Sleep(1000 * time.Second)
}

func (m *Manage) doTick(timeVal time.Duration, isLoop bool, item *Item) {
	t := time.NewTicker(timeVal)
	for {
		select {
		case <-t.C:
			// 执行操作
			go func() {
				if item.ExecFunc(item) && item.CallBackFunc != nil {
					item.CallBackFunc(item)
				}
			}()
		}
		if !isLoop {
			t.Stop()
			break
		}
	}
	nowTime := time.Unix(time.Now().Unix(), 0)
	timeVal, _ = m.calculationTime(item.RuleList, nowTime)
	go m.doTick(timeVal, isLoop, item)
}

// 是否启用多线程执行任务
func (m *Manage) execute(timer time.Time) {
	if len(m.list) > 1 {
		thread := 5
		if len(m.list) > 20 {
			thread = 10
		}
		poll := component.NewPool(thread, func(obj ...interface{}) bool {
			return m.runTask(obj[0].(*Item), timer)
		})
		var taskList []interface{}
		for _, v := range m.list {
			taskList = append(taskList, v)
		}
		poll.AddTaskInterface(taskList)
		poll.Start()
	} else {
		m.runTask(m.list[0], timer)
	}

}

// 执行定时任务
func (m *Manage) runTask(item *Item, nowTime time.Time) bool {
	if item.ExecFunc == nil {
		return false
	}
	// 验证规则
	for _, rule := range item.RuleList {
		if !m.checkRule(rule, nowTime, item.LastExecTime) {
			return false
		}
	}
	item.LastExecTime = nowTime
	// 回调方法
	if item.ExecFunc(item) && item.CallBackFunc != nil {
		item.CallBackFunc(item)
	}
	return true
}

// 检查定时任务时间格式
func (m *Manage) checkRule(rule *Rule, nowTime time.Time, lastTime time.Time) bool {
	if rule == nil {
		return true
	}
	if rule.Loop && lastTime.IsZero() {
		return true
	}
	var timeVal int
	switch rule.Type {
	case Second:
		if rule.Loop {
			timeVal = int(nowTime.Sub(lastTime).Seconds())
		} else {
			timeVal = nowTime.Second()
		}
	case Minute:
		if rule.Loop {
			timeVal = int(nowTime.Sub(lastTime).Minutes())
		} else {
			timeVal = nowTime.Minute()
		}
	case Hour:
		if rule.Loop {
			timeVal = int(nowTime.Sub(lastTime).Hours())
		} else {
			timeVal = nowTime.Hour()
		}
	case Day:
		if rule.Loop {
			timeVal = int(nowTime.Sub(lastTime).Hours() / 24)
		} else {
			timeVal = nowTime.Day()
		}
	case Week:
		if rule.Loop {
			timeVal = int(nowTime.Weekday() - lastTime.Weekday())
		} else {
			timeVal = int(nowTime.Weekday())
		}
	case Month:
		if rule.Loop {
			timeVal = int(nowTime.Month() - lastTime.Month())
		} else {
			timeVal = int(nowTime.Month())
		}
	default:
		return false
	}
	if rule.Loop {
		return timeVal >= rule.Value
	}
	return timeVal == rule.Value
}

// 计算定时任务时长
// 返回值：距离下次任务还有多少duration，是否循环执行定时任务
// 逻辑：获取到当前日期，然后依次去比较定时规则，如果规则是 */1这种格式，那么就表示为循环执行任务，具备周期性。
// 如果是数字，那么表示指定时间执行，如：0 0 0 12 * *，表示每天12点整执行任务，时间不具备周期性，
// 所以这种情况就需要重新去计算下次执行任务时间。
func (m *Manage) calculationTime(ruleList []*Rule, nowTime time.Time) (time.Duration, bool) {
	year := nowTime.Year()
	month := nowTime.Month()
	day := nowTime.Day()
	hour := nowTime.Hour()
	minute := nowTime.Minute()
	second := nowTime.Second()
	// 是否循环执行任务
	isLoop := false
	for _, rule := range ruleList {
		if rule == nil {
			continue
		}
		value := rule.Value
		loop := rule.Loop
		if loop {
			isLoop = true
		}
		switch rule.Type {
		case Second:
			if loop {
				second += value
			} else {
				if second >= value {
					second = value
					minute += 1
				} else {
					second += value - second
				}
			}
		case Minute:
			if loop {
				minute += value
			} else {
				if minute >= value {
					minute = value
					hour += 1
				} else {
					minute += value - minute
				}
			}
		case Hour:
			if loop {
				hour += value
			} else {
				if hour > value {
					hour = value
					day += 1
				} else {
					hour += value - hour
				}
			}
		case Day:
			if loop {
				day += value
			} else {
				if day > value {
					day = value
					month += 1
				} else {
					day += value - day
				}
			}
		case Week:
			val := time.Month(value)
			if loop {
				month += val
			} else {
				if month > val {
					month = val
					year += 1
				} else {
					month += val - month
				}
			}
		case Month:
			if loop {
				year += value
			} else {
				year = value
			}
		default:
			return 0, isLoop
		}
	}

	nextTime := time.Date(year, month, day, hour, minute, second, 0, nowTime.Location())
	return nextTime.Sub(nowTime), isLoop
}
