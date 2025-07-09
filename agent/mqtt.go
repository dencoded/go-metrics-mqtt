package agent

//go:generate protoc --go_out=. metric.proto
//go:generate protoc --descriptor_set_out=message/metric_descriptor.desc metric.proto

import (
	"crypto/tls"
	"fmt"
	"runtime/metrics"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"google.golang.org/protobuf/proto"

	"github.com/dencoded/go-metrics-mqtt/agent/message"
)

type mqttPublisher struct {
	appId  string
	opts   *mqtt.ClientOptions
	client mqtt.Client
}

func newPublisher(brokerEndpoint string, clientId string, tlsConfig *tls.Config) *mqttPublisher {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(brokerEndpoint)
	opts.SetClientID(clientId)
	opts.SetProtocolVersion(4)
	if tlsConfig != nil {
		opts.SetTLSConfig(tlsConfig)
	}
	client := mqtt.NewClient(opts)

	return &mqttPublisher{
		appId:  strings.ReplaceAll(clientId, " ", "-"),
		opts:   opts,
		client: client,
	}
}

func (p mqttPublisher) connect() error {
	// connect to MQTT broker
	if token := p.client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func (p mqttPublisher) disconnect() {
	p.client.Disconnect(1000)
}

func (p mqttPublisher) publishMetrics(ch <-chan timestampedSample) {
	for tsSample := range ch {
		name, value := tsSample.sample.Name, tsSample.sample.Value
		floatValue := float64(0)
		switch value.Kind() {
		case metrics.KindUint64:
			floatValue = float64(value.Uint64())
		case metrics.KindFloat64:
			floatValue = value.Float64()
		case metrics.KindFloat64Histogram:
			floatValue = p.medianBucket(value.Float64Histogram())
		default:
			// TODO: log this case
			continue
		}

		// serialize payload
		msg := &message.MetricData{
			Value:     floatValue,
			Timestamp: tsSample.timestamp,
		}
		payload, err := proto.Marshal(msg)
		if err != nil {
			// TODO: log this case
			continue
		}

		topic := p.appId + name

		// publish to MQTT broker
		token := p.client.Publish(topic, 0, false, payload)
		token.Wait()
		if err := token.Error(); err != nil {
			fmt.Printf("Error publishing metrics: %s\n", err)
			// TODO: log this case
		}
	}
}

// medianBucket takes median from histagram, this is as-is from Golang examples and needs to be revised
func (p mqttPublisher) medianBucket(h *metrics.Float64Histogram) float64 {
	total := uint64(0)
	for _, count := range h.Counts {
		total += count
	}
	thresh := total / 2
	total = 0
	res := 0.0
	for i, count := range h.Counts {
		total += count
		if total >= thresh {
			res = h.Buckets[i]
			break
		}
	}
	return res
}
