package memberlist

import (
	"sync"
	"time"

	"github.com/armon/go-metrics"
)

// awareness manages a simple metric for tracking the estimated health of the
// local node. Health is primary the node's ability to respond in the soft
// real-time manner required for correct health checking of other nodes in the
// cluster.
// 对本机的状态评估
type awareness struct {
	sync.RWMutex

	// max is the upper threshold for the timeout scale (the score will be
	// constrained to be from 0 <= score < max).
	max int // 健康指数最大值

	// score is the current awareness score. Lower values are healthier and
	// zero is the minimum value.
	score int
}

// newAwareness returns a new awareness object.
func newAwareness(max int) *awareness { // 直接返回0
	return &awareness{
		max:   max,
		score: 0,
	}
}

// ApplyDelta takes the given delta and applies it to the score in a thread-safe
// manner. It also enforces a floor of zero and a max of max, so deltas may not
// change the overall score if it's railed at one of the extremes.
func (a *awareness) ApplyDelta(delta int) { // 更新健康指数，加锁确保线程安全
	a.Lock()
	initial := a.score
	a.score += delta
	if a.score < 0 {
		a.score = 0
	} else if a.score > (a.max - 1) {
		a.score = (a.max - 1)
	}
	final := a.score
	a.Unlock()

	if initial != final {
		metrics.SetGauge([]string{"memberlist", "health", "score"}, float32(final))
	}
}

// GetHealthScore returns the raw health score.
func (a *awareness) GetHealthScore() int { // 返回健康指数
	a.RLock()
	score := a.score
	a.RUnlock()
	return score
}

// ScaleTimeout takes the given duration and scales it based on the current
// score. Less healthyness will lead to longer timeouts.
func (a *awareness) ScaleTimeout(timeout time.Duration) time.Duration {
	a.RLock()
	score := a.score
	a.RUnlock()
	return timeout * (time.Duration(score) + 1) // 自身健康状况不佳的情况下会等待更长时间
}
