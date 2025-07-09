package agent

import (
	"runtime/metrics"
	"time"
)

func (a Agent) readMetric(metricName string) {
	metricSample := metrics.Sample{
		Name: metricName,
	}
	samples := []metrics.Sample{metricSample}
	tsSample := timestampedSample{}

	for {
		select {
		case <-a.stopCh:
			return
		case <-time.After(a.period):
			metrics.Read(samples)
			tsSample.sample = samples[0]
			tsSample.timestamp = time.Now().Unix()
			select {
			case <-a.stopCh:
				return
			case a.samplesCh <- tsSample:
			default:
				// we will drop read sample if the buffer is full
				continue
			}
		}
	}
}
