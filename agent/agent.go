package agent

import (
	"crypto/tls"
	"runtime/metrics"
	"time"
)

type Agent struct {
	samplesCh chan timestampedSample
	stopCh    chan struct{}
	period    time.Duration
	publisher *mqttPublisher
}

func New(mqttBroker string, mqttClientId string, tlsConfig *tls.Config, bufSize uint64, period time.Duration) *Agent {
	if bufSize == 0 {
		bufSize = defaultBufferSize
	}
	return &Agent{
		samplesCh: make(chan timestampedSample, bufSize),
		stopCh:    make(chan struct{}),
		period:    period,
		publisher: newPublisher(mqttBroker, mqttClientId, tlsConfig),
	}
}

func (a Agent) Start() error {
	// connect to MQTT broker
	if err := a.publisher.connect(); err != nil {
		return err
	}

	// start reading metrics Go-routines
	allMetrics := metrics.All()
	for _, m := range allMetrics {
		go a.readMetric(m.Name)
	}

	go a.publisher.publishMetrics(a.samplesCh)

	return nil
}

func (a Agent) Stop() {
	// stop reading metrics Go-routines
	close(a.stopCh)
	time.Sleep(5 * time.Second)
	// stop publisher Go-routine and disconnect from broker
	close(a.samplesCh)
	a.publisher.disconnect()
}
