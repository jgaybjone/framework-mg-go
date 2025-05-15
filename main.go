package main

import (
	"framework-mg/controller"

	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/fx"
)

func main() {
	fmt.Println("Hello World")
	fx.New(
		fx.Provide(
			AsRouter(controller.NewIndexController),
			AsRouter(controller.NewFrameworkController),
			NewMqttClient,
			fx.Annotate(NewHttpServer, fx.ParamTags(`group:"routers"`)),
		),
		fx.Invoke(func(server *gin.Engine) {
			server.SetTrustedProxies([]string{"127.0.0.1"})
			go server.Run(":8080")
		}),
		fx.Invoke(func(mqtt mqtt.Client) {
			// go func() {
			// 	if token := mqtt.Connect(); token.Wait() && token.Error() != nil {
			// 		panic(token.Error())
			// 	}
			// 	mqtt.Publish("homeassistant/switch/building/door_lock/config", 0, false, "{\"unique_id\":\"building-door-lock-001\",\"name\":\"大楼门禁\",\"icon\":\"mdi:gesture-tap-button\",\"state_topic\":\"homeassistant/switch/building/door_lock/state\",\"command_topic\":\"homeassistant/switch/building/door_lock/set\",\"json_attributes_topic\":\"homeassistant/switch/building/door_lock/attributes\",\"device\":{\"identifiers\":\"door-lock-001\",\"manufacturer\":\"华为\",\"model\":\"LK\",\"name\":\"esp32\",\"sw_version\":\"1.0\"}}")
			// 	mqtt.Disconnect(1)
			// }()
		}),
	).Run()
}

func AsRouter(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(controller.Router)),
		fx.ResultTags(`group:"routers"`),
	)
}

func NewHttpServer(rs []controller.Router) *gin.Engine {

	server := gin.Default()
	for _, v := range rs {
		v.Handler(server)
	}
	return server
}

func NewMqttClient() mqtt.Client {
	opts := mqtt.NewClientOptions().SetUsername("YzyMqttClient").SetPassword("YzyMqttClient").AddBroker("tcp://192.168.5.8:1883").SetClientID(uuid.NewString())
	client := mqtt.NewClient(opts)
	return client
}
