package main

import (
	"fmt"
	"os"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

const (
	MQTT_BROKER = "tcp://localhost:1883"
	MQTT_TOPIC  = "path/to/topic"
)

func createMQTTClient(brokerAddr, clientId, username, password string) MQTT.Client {
	opts := MQTT.NewClientOptions().AddBroker(brokerAddr)
	opts.SetClientID(clientId)
	opts.SetUsername(username)
	opts.SetPassword(password)
	opts.SetWill(MQTT_TOPIC, "[Lost] clientId: "+clientId, 2, true)
	client := MQTT.NewClient(opts)
	return client
}

func subscribe(client MQTT.Client, sub chan<- MQTT.Message) {
	fmt.Println("start subscribing...")
	subToken := client.Subscribe(
		MQTT_TOPIC,
		0,
		func(client MQTT.Client, msg MQTT.Message) {
			sub <- msg
		})
	if subToken.Wait() && subToken.Error() != nil {
		fmt.Println(subToken.Error())
		os.Exit(1)
	}
}

func publish(client MQTT.Client, input string) {
	token := client.Publish(MQTT_TOPIC, 0, true, input)
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
