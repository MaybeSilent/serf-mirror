package serf

import (
	"sync/atomic"
)

// lamport clock 分布式时钟
// LamportClock is a thread safe implementation of a lamport clock. It
// uses efficient atomic operations for all of its functions, falling back
// to a heavy lock only if there are enough CAS failures. // 过多的CAS失败会进行降级
type LamportClock struct {
	counter uint64 // lamport逻辑时钟
}

// LamportTime is the value of a LamportClock.
type LamportTime uint64

// Time is used to return the current value of the lamport clock
func (l *LamportClock) Time() LamportTime {
	return LamportTime(atomic.LoadUint64(&l.counter))
}

// Increment is used to increment and return the value of the lamport clock
func (l *LamportClock) Increment() LamportTime {
	return LamportTime(atomic.AddUint64(&l.counter, 1))
}

// Witness is called to update our local clock if necessary after
// witnessing a clock value received from another process
func (l *LamportClock) Witness(v LamportTime) {
WITNESS:
	// If the other value is old, we do not need to do anything
	cur := atomic.LoadUint64(&l.counter)
	other := uint64(v)
	if other < cur {
		return // 逻辑时钟更新，如果本地更新则不更新
	}

	// Ensure that our local clock is at least one ahead.
	if !atomic.CompareAndSwapUint64(&l.counter, cur, other+1) {
		// The CAS failed, so we just retry. Eventually our CAS should
		// succeed or a future witness will pass us by and our witness
		// will end.
		goto WITNESS
	}
}
