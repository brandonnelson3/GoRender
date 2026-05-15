package benchmark

import (
	"fmt"
	"sync"
	"time"

	"github.com/brandonnelson3/GoRender/messagebus"
	"github.com/go-gl/glfw/v3.1/glfw"
)

var (
	metricsMu     sync.Mutex
	phaseStarts   = make(map[string]float64)
	phaseAverages = make(map[string]*rollingAverage)

	// RecordMode enables per-frame raw data collection.
	RecordMode      bool
	records         []map[string]float64
	currentFrameRaw = make(map[string]float64)
)

type rollingAverage struct {
	samples [100]float64
	index   int
	count   int
}

func (ra *rollingAverage) add(val float64) {
	ra.samples[ra.index] = val
	ra.index = (ra.index + 1) % 100
	if ra.count < 100 {
		ra.count++
	}
}

func (ra *rollingAverage) average() float64 {
	if ra.count == 0 {
		return 0
	}
	sum := 0.0
	for i := 0; i < ra.count; i++ {
		sum += ra.samples[i]
	}
	return sum / float64(ra.count)
}

// Start marks the beginning of a benchmarked phase.
func Start(name string) {
	metricsMu.Lock()
	defer metricsMu.Unlock()
	phaseStarts[name] = glfw.GetTime()
}

// End marks the end of a benchmarked phase and records the duration.
func End(name string) {
	now := glfw.GetTime()
	metricsMu.Lock()
	defer metricsMu.Unlock()
	start, ok := phaseStarts[name]
	if !ok {
		return
	}
	duration := now - start
	ra, ok := phaseAverages[name]
	if !ok {
		ra = &rollingAverage{}
		phaseAverages[name] = ra
	}
	ra.add(duration)

	if RecordMode {
		currentFrameRaw[name] = duration
	}
}

// RecordFrame captures all phase durations for the current frame and resets the frame-local counters.
func RecordFrame() {
	if !RecordMode {
		return
	}
	metricsMu.Lock()
	defer metricsMu.Unlock()

	newRecord := make(map[string]float64)
	for k, v := range currentFrameRaw {
		newRecord[k] = v
	}
	records = append(records, newRecord)
	currentFrameRaw = make(map[string]float64)
}

// WriteSummary prints a statistical summary of all recorded frames and saves the raw data to a CSV.
func WriteSummary() {
	metricsMu.Lock()
	defer metricsMu.Unlock()

	if len(records) == 0 {
		fmt.Println("No benchmark records to summarize.")
		return
	}

	fmt.Printf("\n--- Benchmark Results (%d frames) ---\n", len(records))
	fmt.Printf("%-25s | %-10s | %-10s | %-10s\n", "Phase", "Avg (ms)", "Min (ms)", "Max (ms)")
	fmt.Println("----------------------------------------------------------------------")

	phases := make(map[string][]float64)
	for _, rec := range records {
		for k, v := range rec {
			phases[k] = append(phases[k], v*1000)
		}
	}

	for name, times := range phases {
		var sum, min, max float64
		min = 1e9
		for _, t := range times {
			sum += t
			if t < min {
				min = t
			}
			if t > max {
				max = t
			}
		}
		avg := sum / float64(len(times))
		fmt.Printf("%-25s | %-10.4f | %-10.4f | %-10.4f\n", name, avg, min, max)
	}
	fmt.Println("----------------------------------------------------------------------")
}

func init() {
	go func() {
		for range time.Tick(time.Millisecond * 500) {
			if RecordMode {
				continue
			}
			metricsMu.Lock()
			for name, ra := range phaseAverages {
				avg := ra.average()
				// Send as milliseconds for better readability in UI
				messagebus.SendAsync(&messagebus.Message{
					Type:  "console",
					Data1: "benchmark_" + name,
					Data2: fmt.Sprintf("%.4f", avg*1000),
				})
			}
			metricsMu.Unlock()
		}
	}()
}
