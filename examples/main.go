package main

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"os"
	"time"

	"github.com/dencoded/go-metrics-mqtt/agent"
)

func main() {
	cert, err := tls.LoadX509KeyPair(
		"/path/to/your/thing-two.cert.pem",
		"/path/to/your/thing.private.key")
	if err != nil {
		panic(err)
	}

	caCert, err := os.ReadFile("/path/to/root-CA.crt")
	if err != nil {
		panic(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}

	metricsAgent := agent.New(
		"tls://your_AWS_IoT_Core_endpoint:8883",
		"your-thing-ID", // remember to set up policy for this thing's certificate so it can connect, pub/sub
		tlsConfig,
		1024,
		5*time.Second)
	err = metricsAgent.Start()
	if err != nil {
		log.Fatal(err)
	}
	defer metricsAgent.Stop()

	time.Sleep(23 * time.Second)
}
