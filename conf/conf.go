package conf

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
	"reflect"
	"strings"
)

type Profile struct {
	Name             string `mapstructure:"name"`                // 服务器名称
	Host             string `mapstructure:"host"`                // 服务器地址
	Port             int    `mapstructure:"port"`                // 服务器监听端口号
	TcpVersion       string `mapstructure:"tcp_version"`         // 服务器版本号
	MaxConn          int    `mapstructure:"max_conn"`            // 最大连接数
	MaxPacketSize    uint32 `mapstructure:"max_packet_size"`     // 一个客户端数据包的最大数据长度
	WorkerPoolSize   int    `mapstructure:"worker_pool_size"`    // Worker 数量
	MaxWorkerTaskLen int    `mapstructure:"max_worker_task_len"` // Worker 任务队列长度
	MaxMsgChanLen    int    `mapstructure:"max_msg_chan_len"`    // 连接发送队列的缓冲区长度
	LogFileName      string `mapstructure:"log_file_name"`       // 日志文件，为空则不保存
	MaxHeartbeatTime int    `mapstructure:"max_heartbeat_time"`  // 发送心跳信息的最大时间间隔
}

var GlobalProfile *Profile

func init() {
	Reload()
}

func setViperDefault() {
	viper.SetDefault("name", "DefaultName")
	viper.SetDefault("host", "127.0.0.1")
	viper.SetDefault("port", 6177)
	viper.SetDefault("tcp_version", "tcp4")
	viper.SetDefault("max_conn", 12000)
	viper.SetDefault("max_packet_size", 0)
	viper.SetDefault("worker_pool_size", 10)
	viper.SetDefault("max_worker_task_len", 1024)
	viper.SetDefault("max_msg_chan_len", 1024)
	viper.SetDefault("log_file_name", "")
	viper.SetDefault("max_heartbeat_time", 0)
}

// Reload 重新加载配置文件
func Reload() {
	// 读取配置文件
	setViperDefault()
	viper.SetConfigFile("config.yaml")
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Print(err)
		os.Exit(1)
	}

	// 加载配置文件
	if err := viper.Unmarshal(&GlobalProfile); err != nil {
		fmt.Print(err)
		os.Exit(1)
	}
}

func PrintGlobalProfile() {
	globalProfileValue := reflect.ValueOf(GlobalProfile).Elem()
	globalProfileType := reflect.TypeOf(*GlobalProfile)

	fmt.Println(`
======================================================
*                    GlobalProfile                   *
======================================================`)

	builder := strings.Builder{}
	for i := 0; i < globalProfileValue.NumField(); i++ {
		name := globalProfileType.Field(i).Name
		value := globalProfileValue.Field(i).Interface()

		builder.WriteString(fmt.Sprintf("    %v:%v\n", name, value))
	}

	fmt.Print(builder.String())
	fmt.Println("======================================================")
}
