/*
Golang implementation of a rundown protection for accessing a shared object

Source code and other details for the project are available at GitHub:

	https://github.com/RandLabs/rundown-protection

More usage please see README.md and tests.
*/

package rundown_protection

import (
	"sync/atomic"
)

//------------------------------------------------------------------------------

const (
	rundownActive uint32 = 0x80000000
)

//------------------------------------------------------------------------------

type RundownProtection struct {
	counter uint32
	done chan struct{}
}

//------------------------------------------------------------------------------

func Create() *RundownProtection {
	r := &RundownProtection{}
	r.done = make(chan struct{})
	return r
}

func (r *RundownProtection) Acquire() bool {
	for {
		val := atomic.LoadUint32(&r.counter)
		if (val & rundownActive) != 0 {
			return false
		}

		if atomic.CompareAndSwapUint32(&r.counter, val, val + 1) {
			break
		}
	}
	return true
}

func (r *RundownProtection) Release() {
	for {
		val := atomic.LoadUint32(&r.counter)
		newVal := (val & rundownActive) | ((val & (^rundownActive)) - 1)
		if atomic.CompareAndSwapUint32(&r.counter, val, newVal) {
			if newVal == rundownActive {
				r.done <- struct{}{}
			}
			break
		}
	}
	return
}

func (r *RundownProtection) Wait() {
	for {
		val := atomic.LoadUint32(&r.counter)
		if atomic.CompareAndSwapUint32(&r.counter, val, val | rundownActive) {
			break
		}
	}

	//wait
	<- r.done
	return
}