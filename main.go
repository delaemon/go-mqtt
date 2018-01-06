package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

const (
	MQTT_BROKER   = "tcp://localhost:1883"
	MQTT_TOPIC    = "path/to/topic"
	MQTT_RETAIN   = false
	MQTT_QOS_ZERO = 0
	MQTT_QOS_TWO  = 2
)

func NewTLSConfig() *tls.Config {
	certpool := x509.NewCertPool()
	pemCerts, err := ioutil.ReadFile("samplecerts/CAfile.pem")
	if err == nil {
		certpool.AppendCertsFromPEM(pemCerts)
	}

	cert, err := tls.LoadX509KeyPair("samplecerts/client-crt.pem", "samplecerts/client-key.pem")
	if err != nil {
		panic(err)
	}

	cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		panic(err)
	}
	fmt.Println(cert.Leaf)

	return &tls.Config{
		RootCAs:            certpool,
		ClientAuth:         tls.NoClientCert,
		ClientCAs:          nil,
		InsecureSkipVerify: true,
		Certificates:       []tls.Certificate{cert},
	}
}

func createMQTTClient(brokerAddr, clientId, username, password string) MQTT.Client {
	tlsconfig := NewTLSConfig()
	opts := MQTT.NewClientOptions().AddBroker(brokerAddr)
	opts.SetClientID(clientId).SetTLSConfig(tlsconfig)
	opts.SetUsername(username)
	opts.SetPassword(password)
	opts.SetWill(MQTT_TOPIC, "[Lost] clientId: "+clientId, MQTT_QOS_TWO, MQTT_RETAIN)
	client := MQTT.NewClient(opts)
	return client
}

func subscribe(client MQTT.Client, sub chan<- MQTT.Message) {
	fmt.Println("start subscribing...")
	subToken := client.Subscribe(
		MQTT_TOPIC,
		MQTT_QOS_ZERO,
		func(client MQTT.Client, msg MQTT.Message) {
			sub <- msg
		})
	if subToken.Wait() && subToken.Error() != nil {
		fmt.Println(subToken.Error())
		os.Exit(1)
	}
}

func publish(client MQTT.Client, input string) {
	token := client.Publish(MQTT_TOPIC, MQTT_QOS_ZERO, MQTT_RETAIN, input)
	if token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
}

func input(pub chan<- string) {
	for {
		var input string
		fmt.Scanln(&input)
		pub <- input
	}
}

func main() {
	fmt.Print("your id: ")
	var id string
	fmt.Scanln(&id)
	client := createMQTTClient(MQTT_BROKER, id, "YourID", "YouPassword")
	defer client.Disconnect(250)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	sub := make(chan MQTT.Message)
	go subscribe(client, sub)
	pub := make(chan string)
	go input(pub)
	for {
		select {
		case s := <-sub:
			msg := string(s.Payload())
			fmt.Printf("\nmsg: %s\n", msg)
		case p := <-pub:
			publish(client, p)
		}
	}
}
