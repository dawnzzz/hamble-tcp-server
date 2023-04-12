package conf

import (
	"fmt"
	"github.com/spf13/viper"
	"os"
	"reflect"
	"strings"
	"time"
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
	MaxHeartbeatTime int    `mapstructure:"max_heartbeat_time"`  // 心跳检测的最大时间间隔
	CrtFileName      string `mapstructure:"crt_file_name"`
	KeyFileName      string `mapstructure:"key_file_name"`
	PrintBanner      bool   `mapstructure:"print_banner"`
}

func (profile *Profile) GetMaxHeartbeatTime() time.Duration {
	return time.Duration(profile.MaxHeartbeatTime) * time.Second
}

var GlobalProfile *Profile

func init() {
	GlobalProfile = &Profile{
		Name:             "DefaultName",
		Host:             "127.0.0.1",
		Port:             6177,
		TcpVersion:       "tcp4",
		MaxConn:          12000,
		MaxPacketSize:    0,
		WorkerPoolSize:   10,
		MaxWorkerTaskLen: 1024,
		MaxMsgChanLen:    1024,
		LogFileName:      "",
		MaxHeartbeatTime: 10,
		CrtFileName:      "crt.pem",
		KeyFileName:      "crt.pem",
		PrintBanner:      true,
	}
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
	viper.SetDefault("max_heartbeat_time", 10)
	viper.SetDefault("crt_file_name", "crt.pem")
	viper.SetDefault("key_file_name", "key.pem")
	viper.SetDefault("print_banner", true)
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

func BindProfile(profile *Profile) {
	if profile.Name != "" {
		GlobalProfile.Name = profile.Name
	}

	if profile.Host != "" {
		GlobalProfile.Host = profile.Host
	}

	if profile.Port != 0 {
		GlobalProfile.Port = profile.Port
	}

	if profile.TcpVersion != "" {
		GlobalProfile.TcpVersion = profile.TcpVersion
	}

	if profile.MaxConn != 0 {
		GlobalProfile.MaxConn = profile.MaxConn
	}

	if profile.MaxPacketSize != 0 {
		GlobalProfile.MaxPacketSize = profile.MaxPacketSize
	}

	if profile.WorkerPoolSize != 0 {
		GlobalProfile.WorkerPoolSize = profile.WorkerPoolSize
	}

	if profile.MaxWorkerTaskLen != 0 {
		GlobalProfile.MaxWorkerTaskLen = profile.MaxWorkerTaskLen
	}

	if profile.MaxMsgChanLen != 0 {
		GlobalProfile.MaxMsgChanLen = profile.MaxMsgChanLen
	}

	if profile.LogFileName != "" {
		GlobalProfile.LogFileName = profile.LogFileName
	}

	if profile.MaxHeartbeatTime != 0 {
		GlobalProfile.MaxHeartbeatTime = profile.MaxHeartbeatTime
	}

	if profile.CrtFileName != "" {
		GlobalProfile.CrtFileName = profile.CrtFileName
	}

	if profile.KeyFileName != "" {
		GlobalProfile.KeyFileName = profile.KeyFileName
	}

	if profile.PrintBanner != GlobalProfile.PrintBanner {
		GlobalProfile.PrintBanner = profile.PrintBanner
	}
}
