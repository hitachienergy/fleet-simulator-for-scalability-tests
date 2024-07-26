package simulation

import (
	"fmt"
	"math"
	"os"
	"sync"
	"time"
)

type SimulationResult struct {
	StartAt  time.Time
	Duration float64
	Success  bool
}

type SimulationStats struct {
	ExecutionTime float64 `json:"Execution-Time"`
	SuccessCount  int32   `json:"Success"`
	FinishCount   int32   `json:"Finished"`
	Total         int32   `json:"Total"`
	Min           float64 `json:"Device-Min-Time"`
	Max           float64 `json:"Device-Max-Time"`
	Avg           float64 `json:"Device-Avg-Time"`
}

// dataStore is a thread-safe central storage of simulation results
type dataStore struct {
	*sync.Mutex // use the basic Mutex instead of RWMutex because of heavy write (devices) but single read (httpserver)

	startAt time.Time
	endAt   time.Time

	min          float64
	max          float64
	sum          float64
	successCount int32
	finishCount  int32
	total        int32

	details map[string]SimulationResult
}

// newDataStore creates a new instance of the data storage
func newDataStore(total int) *dataStore {
	return &dataStore{
		Mutex:   &sync.Mutex{},
		details: map[string]SimulationResult{},
		min:     math.MaxFloat64,
		max:     -1,
		total:   int32(total),
	}
}

// storeState stores a new simulation result of a device
func (s *dataStore) storeState(id string, start time.Time, elapse time.Duration, success bool) (finish bool) {
	s.Lock()
	defer s.Unlock()

	if _, ok := s.details[id]; ok {
		return s.finishCount == s.total
	}

	if s.finishCount == 0 {
		s.startAt = start
	}

	duration := float64(elapse.Nanoseconds()) / float64(time.Second)

	s.details[id] = SimulationResult{StartAt: start, Duration: duration, Success: success}
	s.endAt = start.Add(elapse)
	s.sum += duration
	s.finishCount += 1
	if success {
		s.successCount += 1
	}
	if duration > s.max {
		s.max = duration
	}
	if duration < s.min {
		s.min = duration
	}

	return s.finishCount == s.total
}

// getStatistics get the in-time statistics of the simulation
func (s *dataStore) getStatistics() (stats SimulationStats) {
	s.Lock()
	defer s.Unlock()

	execTime := 0.0
	avg := 0.0
	max := 0.0
	min := 0.0
	if s.finishCount > 0 {
		execTime = float64(time.Since(s.startAt).Nanoseconds()) / float64(time.Second)
		avg = s.sum / float64(s.finishCount)
		max = s.max
		min = s.min
	}

	return SimulationStats{
		ExecutionTime: execTime,
		Min:           min,
		Max:           max,
		Avg:           avg,
		SuccessCount:  s.successCount,
		FinishCount:   s.finishCount,
		Total:         s.total,
	}
}

// saveToDisk is called after the whole simulation finishes. It stores the analysis to the target path
func (s *dataStore) saveToDisk(opth string) error {
	s.Lock()
	defer s.Unlock()

	detailsBuffer := []byte("Device StartAt Duration(s) Success\n")
	mean := s.sum / float64(s.finishCount)
	std := 0.0
	for device, data := range s.details {
		std += math.Pow(data.Duration-mean, 2)
		detailsBuffer = append(detailsBuffer,
			[]byte(fmt.Sprintf("%s %d %.9f %t\n", device, data.StartAt.Unix(), data.Duration, data.Success))...)
	}
	std = math.Sqrt(std / float64(s.finishCount))

	buffer := []byte{}
	buffer = append(buffer,
		[]byte(fmt.Sprintf("Total: %d\n"+
			"Success: %d\n"+
			"StartAt: %d\n"+
			"Duration: %fs\n"+
			"Avg Execution Time per Device: %.9fs\n"+
			"Max Execution Time per Device: %.9fs\n"+
			"Min Execution Time per Device:: %.9fs\n"+
			"Std: %.9f\n\n",
			s.total, s.successCount,
			s.startAt.Unix(), float64(s.endAt.Sub(s.startAt).Nanoseconds())/float64(time.Second),
			mean, s.max, s.min, std))...)
	buffer = append(buffer, detailsBuffer...)

	err := os.WriteFile(opth, buffer, 0644)
	return err
}
