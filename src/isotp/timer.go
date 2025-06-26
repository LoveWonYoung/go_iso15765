package isotp

import "time"

// Timer 是一个更健壮的定时器实现。
type Timer struct {
	timeout   time.Duration
	startedAt time.Time
	active    bool
}

func NewTimer(timeoutMs int) *Timer {
	return &Timer{timeout: time.Duration(timeoutMs) * time.Millisecond}
}

func (t *Timer) IsTimedOut() bool {
	if !t.active {
		return false
	}
	return time.Since(t.startedAt) > t.timeout
}

func (t *Timer) Stop() {
	t.active = false
}

func (t *Timer) Start() {
	t.startedAt = time.Now()
	t.active = true
}

func (t *Timer) SetTimeout(ms int) {
	t.timeout = time.Duration(ms) * time.Millisecond
}
