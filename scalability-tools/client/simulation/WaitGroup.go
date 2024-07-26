package simulation

import "sync"

// waitGroup is a thread-safe counter to calculates the success/failure counts of multiple obejcts
// It helps the caller to wait until all objects to finish (either success or failure)
type waitGroup struct {
	*sync.Mutex
	readyChan chan int

	successCount int
	failureCount int
	total        int
}

// newWaitGroup creates a new instance of WaitGroup
func newWaitGroup(total int) *waitGroup {
	return &waitGroup{
		Mutex:     &sync.Mutex{},
		readyChan: make(chan int, 1),
		total:     total,
	}
}

// add notifies the WaitGroup instance that a go routine is finishes its job
func (w *waitGroup) add(success bool) {
	w.Lock()
	defer w.Unlock()

	if success {
		w.successCount += 1
	} else {
		w.failureCount += 1
	}

	if w.successCount+w.failureCount == w.total {
		w.readyChan <- w.failureCount
	}
}
