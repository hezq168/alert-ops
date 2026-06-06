package config

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// Conf 全局变量，用来保存程序的所有配置信息
var Conf = new(AppConfig)

// AppConfig 应用总配置结构体
type AppConfig struct {
	*AppInfo     `mapstructure:"app"`
	*LogConfig   `mapstructure:"log"`
	*MysqlConfig `mapstructure:"mysql"`
	//*RedisConfig `mapstructure:"redis"`
	*JwtConfig `mapstructure:"jwt"`
	*AIConfig  `mapstructure:"ai"`
}

// AIConfig AI调用配置
type AIConfig struct {
	MaxConcurrent  int `mapstructure:"max_concurrent"`  // 最大并发数，默认1
	DegradeTimeout int `mapstructure:"degrade_timeout"` // 降级冷却时间（秒），默认60
	QueueSize      int `mapstructure:"queue_size"`      // 队列缓冲区大小，默认100
}

// AppInfo 应用基本信息配置
type AppInfo struct {
	Name      string `mapstructure:"name"`
	Mode      string `mapstructure:"mode"`
	Version   string `mapstructure:"version"`
	StartTime string `mapstructure:"start_time"`
	MachineID int64  `mapstructure:"machine_id"`
	Port      int    `mapstructure:"port"`
}

// LogConfig 日志配置结构体
type LogConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
}

// MysqlConfig 日志配置结构体
type MysqlConfig struct {
	Host         string `mapstructure:"host"`
	Port         int    `mapstructure:"port"`
	Username     string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
	Dbname       string `mapstructure:"dbname"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxLifetime  string `mapstructure:"max_lifetime"`
}

// RedisConfig 日志配置结构体
//type RedisConfig struct {
//	Host     string `mapstructure:"host"`
//	Port     int    `mapstructure:"port"`
//	Password string `mapstructure:"password"`
//	Db       int    `mapstructure:"db"`
//	PoolSize int    `mapstructure:"pool_size"`
//}

type JwtConfig struct {
	Secret string `mapstructure:"secret"`
	Expire int64  `mapstructure:"expire"`
}

// Init 初始化配置，读取并监听配置文件变化
// filename: 配置文件路径
// 返回错误如果读取或解析失败
func Init(filename string) (err error) {
	// 1、直接获取配置文件
	// viper.SetConfigFile("../conf/config.yaml")
	// 2、指定配置文件名和配置文件的位置，viper自行查找可用的配置文件
	// viper.SetConfigName("config")  // 指定配置文件名称(不需要带后缀)
	// viper.SetConfigType("yaml")    // 指定配置文件类型(专用于从远程获取配置信息,基本上是配合远程配置中心使用的，告诉viper当前的数据使用什么格式解析)
	// viper.AddConfigPath("../conf") // 指定查找配置文件的路径 (这里使用相对路径)
	// 设置配置文件路径
	viper.SetConfigFile(filename)

	// 开启环境变量支持
	viper.AutomaticEnv()

	// 读取配置信息
	err = viper.ReadInConfig()
	if err != nil {
		// 读取配置信息失败
		fmt.Printf("viper.ReadInConfig() failed with %s\n", err)
		return err
	}

	// 配置环境变量名映射
	// 环境变量 MYSQL_HOST 自动覆盖 mysql.host
	_ = viper.BindEnv("mysql.host", "MYSQL_HOST")

	// 把读取到的配置信息反序列化到Conf变量中
	if err := viper.Unmarshal(Conf); err != nil {
		fmt.Printf("viper.Unmarshal() failed with %s\n", err)
		return err
	}
	// 监听配置文件变化
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("配置文件修改了:", e.Name)
		if err := viper.Unmarshal(Conf); err != nil {
			fmt.Printf("viper.Unmarshal() failed with %s\n", err)
		}
	})
	return nil
}
