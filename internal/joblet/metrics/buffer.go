package metrics

import (
	"sync"
	"time"

	"github.com/ehsaniara/joblet/internal/joblet/metrics/domain"
)

// MetricsBuffer is a circular buffer that stores recent metrics samples per job
// This buffer prevents gaps when transitioning from historical (persist) to live streaming
// Similar to logs buffer, it retains recent samples in memory
type MetricsBuffer struct {
	// jobBuffers maps jobID to a circular buffer of samples
	jobBuffers     map[string]*circularBuffer
	capacityPerJob int
	mutex          sync.RWMutex
}

// circularBuffer stores metrics samples in a circular buffer
type circularBuffer struct {
	samples  []*domain.JobMetricsSample
	capacity int
	head     int // Index of the oldest sample
	size     int // Current number of samples
	mutex    sync.RWMutex
}

// NewMetricsBuffer creates a new metrics buffer with specified capacity per job
func NewMetricsBuffer(capacityPerJob int) *MetricsBuffer {
	if capacityPerJob <= 0 {
		capacityPerJob = 100 // Default: store last 100 samples per job (~8 minutes at 5s interval)
	}

	return &MetricsBuffer{
		jobBuffers:     make(map[string]*circularBuffer),
		capacityPerJob: capacityPerJob,
	}
}

// Add adds a metrics sample to the buffer for a specific job
func (mb *MetricsBuffer) Add(sample *domain.JobMetricsSample) {
	if sample == nil || sample.JobID == "" {
		return
	}

	mb.mutex.Lock()
	defer mb.mutex.Unlock()

	// Get or create buffer for this job
	buffer, exists := mb.jobBuffers[sample.JobID]
	if !exists {
		buffer = &circularBuffer{
			samples:  make([]*domain.JobMetricsSample, 100), // Default capacity
			capacity: 100,
			head:     0,
			size:     0,
		}
		mb.jobBuffers[sample.JobID] = buffer
	}

	// Add to circular buffer
	buffer.add(sample)
}

// GetRecent returns recent samples for a job (up to limit)
// Samples are returned in chronological order (oldest first)
func (mb *MetricsBuffer) GetRecent(jobID string, limit int) []*domain.JobMetricsSample {
	mb.mutex.RLock()
	defer mb.mutex.RUnlock()

	buffer, exists := mb.jobBuffers[jobID]
	if !exists {
		return nil
	}

	return buffer.getRecent(limit)
}

// GetSince returns samples for a job since the given timestamp (inclusive)
// Samples are returned in chronological order (oldest first)
func (mb *MetricsBuffer) GetSince(jobID string, sinceTimestamp time.Time) []*domain.JobMetricsSample {
	mb.mutex.RLock()
	defer mb.mutex.RUnlock()

	buffer, exists := mb.jobBuffers[jobID]
	if !exists {
		return nil
	}

	return buffer.getSince(sinceTimestamp)
}

// Clear removes all samples for a specific job
// This should be called when a job is cleaned up
func (mb *MetricsBuffer) Clear(jobID string) {
	mb.mutex.Lock()
	defer mb.mutex.Unlock()

	delete(mb.jobBuffers, jobID)
}

// add adds a sample to the circular buffer
func (cb *circularBuffer) add(sample *domain.JobMetricsSample) {
	cb.mutex.Lock()
	defer cb.mutex.Unlock()

	if cb.size < cb.capacity {
		// Buffer not full yet, just append
		cb.samples[cb.size] = sample
		cb.size++
	} else {
		// Buffer full, overwrite oldest
		cb.samples[cb.head] = sample
		cb.head = (cb.head + 1) % cb.capacity
	}
}

// getRecent returns the most recent samples (up to limit)
// Returns in chronological order (oldest first)
func (cb *circularBuffer) getRecent(limit int) []*domain.JobMetricsSample {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	if cb.size == 0 {
		return nil
	}

	// Determine how many samples to return
	count := cb.size
	if limit > 0 && limit < count {
		count = limit
	}

	result := make([]*domain.JobMetricsSample, count)

	// If buffer not full, samples are contiguous from index 0
	if cb.size < cb.capacity {
		startIndex := cb.size - count
		for i := 0; i < count; i++ {
			// Return a copy to prevent external modifications
			result[i] = cb.samples[startIndex+i]
		}
	} else {
		// Buffer is full, need to handle circular wrap-around
		// Start from the oldest sample (head) and go forward
		startIndex := cb.head + (cb.size - count)
		if startIndex >= cb.capacity {
			startIndex -= cb.capacity
		}

		for i := 0; i < count; i++ {
			idx := (startIndex + i) % cb.capacity
			result[i] = cb.samples[idx]
		}
	}

	return result
}

// getSince returns samples since the given timestamp (inclusive)
// Returns in chronological order (oldest first)
func (cb *circularBuffer) getSince(sinceTimestamp time.Time) []*domain.JobMetricsSample {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()

	if cb.size == 0 {
		return nil
	}

	var result []*domain.JobMetricsSample

	// Iterate through buffer in chronological order
	for i := 0; i < cb.size; i++ {
		var sample *domain.JobMetricsSample

		if cb.size < cb.capacity {
			// Buffer not full, samples are contiguous from index 0
			sample = cb.samples[i]
		} else {
			// Buffer full, start from head
			idx := (cb.head + i) % cb.capacity
			sample = cb.samples[idx]
		}

		// Include samples at or after the timestamp
		if sample != nil && !sample.Timestamp.Before(sinceTimestamp) {
			result = append(result, sample)
		}
	}

	return result
}
