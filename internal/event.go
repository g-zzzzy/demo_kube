package demokubenet

type EventType = string

type Event struct {
	Type EventType
}

// 处理器函数类型
// 携带调度器信息，比如执行时间等
type EventHandler func(*EventBus, Event) error
