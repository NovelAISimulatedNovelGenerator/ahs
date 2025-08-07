package config

import (
	"sync"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// Config 应用配置结构
type Config struct {
	Server     ServerConfig            `mapstructure:"server"`
	RateLimit  RateLimitConfig         `mapstructure:"rate_limit"`
	Log        LogConfig               `mapstructure:"log"`
	WorkerPool WorkerPoolConfig        `mapstructure:"worker_pool"`
	LLMConfigs map[string]LLMConfig `mapstructure:"llm_configs"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host           string        `mapstructure:"host"`
	Port           int           `mapstructure:"port"`
	ReadTimeout    time.Duration `mapstructure:"read_timeout"`
	WriteTimeout   time.Duration `mapstructure:"write_timeout"`
	IdleTimeout    time.Duration `mapstructure:"idle_timeout"`
	MaxHeaderBytes int           `mapstructure:"max_header_bytes"`
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Enabled bool `mapstructure:"enabled"`
	QPS     int  `mapstructure:"qps"`
	Burst   int  `mapstructure:"burst"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level            string   `mapstructure:"level"`
	Encoding         string   `mapstructure:"encoding"`
	OutputPaths      []string `mapstructure:"output_paths"`
	ErrorOutputPaths []string `mapstructure:"error_output_paths"`
}

// WorkerPoolConfig 工作池配置
type WorkerPoolConfig struct {
	Workers   int `mapstructure:"workers"`
	QueueSize int `mapstructure:"queue_size"`
}

// LLMConfig LLM配置结构
type LLMConfig struct {
	APIBaseURL string `mapstructure:"api_base_url"`
	APIKey     string `mapstructure:"api_key"`
	Model      string `mapstructure:"model"`
}

// Load 加载配置文件
func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// 设置默认值
	setDefaults()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	// 解析配置
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// setDefaults 设置默认配置值
func setDefaults() {
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")
	viper.SetDefault("server.idle_timeout", "60s")
	viper.SetDefault("server.max_header_bytes", 1048576)

	viper.SetDefault("rate_limit.enabled", true)
	viper.SetDefault("rate_limit.qps", 50)
	viper.SetDefault("rate_limit.burst", 100)

	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.encoding", "json")
	viper.SetDefault("log.output_paths", []string{"stdout"})
	viper.SetDefault("log.error_output_paths", []string{"stderr"})

	viper.SetDefault("worker_pool.workers", 8)
	viper.SetDefault("worker_pool.queue_size", 32)
	
	// LLM配置默认为空map，用户可在配置文件中定义多个LLM提供商
	viper.SetDefault("llm_configs", map[string]interface{}{})
}

// GetAddress 获取服务器地址
func (s *ServerConfig) GetAddress() string {
	return s.Host + ":" + string(rune(s.Port))
}

// BuildZapConfig 构建Zap日志配置
func (l *LogConfig) BuildZapConfig() zap.Config {
	var level zap.AtomicLevel
	switch l.Level {
	case "debug":
		level = zap.NewAtomicLevelAt(zap.DebugLevel)
	case "info":
		level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}
	return zap.Config{
		Level:            level,
		Development:      false,
		Encoding:         l.Encoding,
		OutputPaths:      l.OutputPaths,
		ErrorOutputPaths: l.ErrorOutputPaths,
		EncoderConfig:    zap.NewProductionEncoderConfig(),
	}
}

var (
	globalConfig     *Config
	globalConfigOnce sync.Once
	globalConfigErr  error
)

// SetGlobalConfig 设置全局配置实例
func SetGlobalConfig(cfg *Config) {
	globalConfig = cfg
}

// GetGlobalConfig 获取全局配置实例
func GetGlobalConfig() *Config {
	return globalConfig
}

// LoadGlobalConfig 加载并设置全局配置
func LoadGlobalConfig(configPath string) (*Config, error) {
	globalConfigOnce.Do(func() {
		globalConfig, globalConfigErr = Load(configPath)
	})
	return globalConfig, globalConfigErr
}
