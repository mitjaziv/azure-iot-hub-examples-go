package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/SentimensRG/ctx/sigctx"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/namsral/flag"
)

func init() {
	mqtt.DEBUG = log.New(os.Stderr, "DEBUG    ", log.Ltime)
	mqtt.WARN = log.New(os.Stderr, "WARNING  ", log.Ltime)
	mqtt.CRITICAL = log.New(os.Stderr, "CRITICAL ", log.Ltime)
	mqtt.ERROR = log.New(os.Stderr, "ERROR    ", log.Ltime)
}

func main() {
	var hub, id, certFile, keyFile string

	flag.StringVar(&hub, "hub-name", "", "Azure IoT Hub name")
	flag.StringVar(&id, "device-id", "", "IoT device IoT")
	flag.StringVar(&certFile, "cert-file", "", "Certificate file")
	flag.StringVar(&keyFile, "key-file", "", "Key file")
	flag.Parse()

	if hub == "" || id == "" || certFile == "" || keyFile == "" {
		log.Fatal("missing parameters")
	}

	var deadline = sigctx.New()

	// Connect to Azure Iot Hub.
	client := connect(hub, id, certFile, keyFile)
	defer client.Disconnect(250)

	// Define topic to publish.
	topic := "messages/events/"

	// Publish message periodically every two seconds.
	go func() {
		for x := range time.Tick(2 * time.Second) {
			// Generate payload for event.
			payload := "{device_time:" + x.String() + "}"

			// Publish payload to topic.
			err := publish(client, id, topic, payload)
			if err != nil {
				log.Fatal(err)
			}
		}
	}()

	<-deadline.Done()
}

func connect(hub string, deviceId string, certFile string, keyFile string) mqtt.Client {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatal(err)
	}

	// Decode and add root certificates.
	certs, err := readPemFromFile()
	if err != nil {
		log.Fatal(err)
	}

	roots := x509.NewCertPool()
	for _, cert := range certs.Certificate {
		x509Cert, err := x509.ParseCertificate(cert)
		if err != nil {
			log.Fatal(err)
		}
		roots.AddCert(x509Cert)
	}

	// Create tls config.
	cfg := tls.Config{
		RootCAs:                  roots,
		InsecureSkipVerify:       false,
		MinVersion:               tls.VersionTLS12,
		MaxVersion:               tls.VersionTLS12,
		PreferServerCipherSuites: true,
		ClientAuth:               tls.RequireAnyClientCert,
		Renegotiation:            tls.RenegotiateFreelyAsClient,
		Certificates:             []tls.Certificate{cert},
	}
	cfg.BuildNameToCertificate()

	// Create MQTT client options.
	opts := &mqtt.ClientOptions{
		ClientID:             deviceId,
		CleanSession:         true,
		AutoReconnect:        true,
		MaxReconnectInterval: 1 * time.Second,
		KeepAlive:            int64(30 * time.Second),
		TLSConfig:            &cfg,
		ProtocolVersion:      4,
	}

	// Add MQTT broker, notice the `tcps` protocol.
	opts.AddBroker("tcps://" + hub + ".azure-devices.net:8883")

	// Set username for Azure IoT Hub.
	opts.Username = hub + ".azure-devices.net/" + deviceId + "/?api-version=2018-06-30"

	// Create client and connect.
	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}
	return client
}

func publish(client mqtt.Client, deviceId string, topic string, payload string) error {
	// Fix topic string for Azure IoT Hub.
	topic = "devices/" + deviceId + "/" + topic

	// Publish message.
	token := client.Publish(topic, 1, false, payload)
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func readPemFromFile() (tls.Certificate, error) {
	// Read cert from file.
	raw, err := ioutil.ReadFile("certs/IoTHubRootCA_Baltimore.pem")
	if err != nil {
		return tls.Certificate{}, err
	}

	// Decode PEM certificate.
	var cert tls.Certificate
	certPEMBlock := []byte(raw)
	var certDERBlock *pem.Block
	for {
		certDERBlock, certPEMBlock = pem.Decode(certPEMBlock)
		if certDERBlock == nil {
			break
		}
		if certDERBlock.Type == "CERTIFICATE" {
			cert.Certificate = append(cert.Certificate, certDERBlock.Bytes)
		}
	}
	return cert, nil
}
