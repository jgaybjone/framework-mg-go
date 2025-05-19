package main

import (
	"framework-mg/controller"
	"os"
	"path/filepath"
	"time"

	"fmt"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger() *zap.Logger {
	// 获取当前工作目录
	workDir, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("Failed to get working directory: %v", err))
	}

	// 确保日志目录存在
	logDir := filepath.Join(workDir, "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		panic(fmt.Sprintf("Failed to create log directory: %v", err))
	}

	writer, err := rotatelogs.New(
		filepath.Join(logDir, "%Y-%m", "app-%Y-%m-%d.log"),
		rotatelogs.WithLinkName(filepath.Join(logDir, "app.log")),
		rotatelogs.WithRotationTime(24*time.Hour),
		rotatelogs.WithMaxAge(365*24*time.Hour),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create rotatelogs: %v", err))
	}
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(writer),
		zapcore.InfoLevel,
	)

	return zap.New(core)
}

func main() {
	fmt.Println("Hello World")

	fx.New(
		fx.Provide(NewLogger),
		fx.WithLogger(func(logger *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: logger}
		}),
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
			go func() {
				if token := mqtt.Connect(); token.Wait() && token.Error() != nil {
					panic(token.Error())
				}
				mqtt.Publish("homeassistant/switch/building/door_lock/config", 0, false, "{\"unique_id\":\"building-door-lock-001\",\"name\":\"大楼门禁\",\"icon\":\"mdi:gesture-tap-button\",\"state_topic\":\"homeassistant/switch/building/door_lock/state\",\"command_topic\":\"homeassistant/switch/building/door_lock/set\",\"json_attributes_topic\":\"homeassistant/switch/building/door_lock/attributes\",\"device\":{\"identifiers\":\"door-lock-001\",\"manufacturer\":\"华为\",\"model\":\"LK\",\"name\":\"esp32\",\"sw_version\":\"1.0\"}}")
				mqtt.Disconnect(1)
			}()
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
