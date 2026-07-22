package configs

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

const DefaultConfigPath = "configs/config.yaml"

type (
	Config struct {
		HTTP    HTTPConfig    `mapstructure:"http"`
		Store   StoreConfig   `mapstructure:"store"`
		MongoDB MongoDBConfig `mapstructure:"mongodb"`
		Admin   AdminConfig   `mapstructure:"admin"`
		Ejudge  EjudgeConfig  `mapstructure:"ejudge"`
		JPlag   JPlagConfig   `mapstructure:"jplag"`
		Analyze AnalyzeConfig `mapstructure:"analyze"`
		AIUsage AIUsageConfig `mapstructure:"aiusage"`
		OpenAI  OpenAIConfig  `mapstructure:"openai"`
	}

	HTTPConfig struct {
		Addr            string        `mapstructure:"addr"`
		ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	}

	StoreConfig struct {
		Dir string `mapstructure:"dir"`
	}

	MongoDBConfig struct {
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
		Host     string `mapstructure:"host"`
		Database string `mapstructure:"database"`
	}

	AdminConfig struct {
		Username string        `mapstructure:"username"`
		Password string        `mapstructure:"password"`
		JWTSecret string       `mapstructure:"jwt_secret"`
		JWTTTL   time.Duration `mapstructure:"jwt_ttl"`
	}

	EjudgeConfig struct {
		BaseURL string        `mapstructure:"base_url"`
		APIKey  string        `mapstructure:"api_key"`
		Timeout time.Duration `mapstructure:"timeout"`
	}

	JPlagConfig struct {
		JarPath string `mapstructure:"jar_path"`
	}

	AnalyzeConfig struct {
		JobsMaxConcurrent  int `mapstructure:"jobs_max_concurrent"`
		AnalyzeConcurrency int `mapstructure:"analyze_concurrency"`
	}

	AIUsageConfig struct {
		Enabled bool `mapstructure:"enabled"`
	}

	OpenAIConfig struct {
		BaseURL  string        `mapstructure:"base_url"`
		Username string        `mapstructure:"username"`
		Password string        `mapstructure:"password"`
		APIKey   string        `mapstructure:"api_key"`
		Timeout  time.Duration `mapstructure:"timeout"`
		Model    string        `mapstructure:"model"`
	}
)

func (c AdminConfig) Enabled() bool {
	return strings.TrimSpace(c.Username) != "" &&
		strings.TrimSpace(c.Password) != "" &&
		strings.TrimSpace(c.JWTSecret) != ""
}

func (c EjudgeConfig) Enabled() bool {
	return strings.TrimSpace(c.BaseURL) != "" && strings.TrimSpace(c.APIKey) != ""
}

func (c MongoDBConfig) URI() (string, error) {
	user := strings.TrimSpace(c.Username)
	host := strings.TrimSpace(c.Host)
	if user == "" || c.Password == "" || host == "" {
		return "", fmt.Errorf("нужны mongodb.username, mongodb.password и mongodb.host")
	}
	if !strings.Contains(host, ":") {
		host += ":27017"
	}
	u := &url.URL{
		Scheme:   "mongodb",
		User:     url.UserPassword(user, c.Password),
		Host:     host,
		Path:     "/",
		RawQuery: "authSource=admin",
	}
	return u.String(), nil
}

func LoadConfig(path string) (*Config, error) {
	_ = godotenv.Load()

	if path == "" {
		path = os.Getenv("CONFIG_PATH")
	}
	if path == "" {
		path = DefaultConfigPath
	}

	v := viper.New()
	v.SetConfigFile(path)
	bindEnv(v)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("config: unmarshal: %w", err)
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) validate() error {
	if !c.Admin.Enabled() {
		return fmt.Errorf("нужны admin.username, admin.password и admin.jwt_secret (или ADMIN_*/JWT_SECRET)")
	}
	if c.Admin.JWTTTL <= 0 {
		return fmt.Errorf("admin.jwt_ttl должен быть > 0")
	}
	if _, err := c.MongoDB.URI(); err != nil {
		return err
	}
	if strings.TrimSpace(c.MongoDB.Database) == "" {
		return fmt.Errorf("mongodb.database обязателен")
	}
	if strings.TrimSpace(c.Store.Dir) == "" {
		return fmt.Errorf("store.dir обязателен")
	}
	if c.Analyze.JobsMaxConcurrent <= 0 {
		return fmt.Errorf("analyze.jobs_max_concurrent должен быть > 0")
	}
	if c.Analyze.AnalyzeConcurrency <= 0 {
		return fmt.Errorf("analyze.analyze_concurrency должен быть > 0")
	}
	if strings.TrimSpace(c.JPlag.JarPath) == "" {
		return fmt.Errorf("jplag.jar_path обязателен")
	}
	if strings.TrimSpace(c.HTTP.Addr) == "" {
		return fmt.Errorf("http.addr обязателен")
	}
	return nil
}

func bindEnv(v *viper.Viper) {
	_ = v.BindEnv("http.addr", "HTTP_ADDR")
	_ = v.BindEnv("store.dir", "STORE_DIR")
	_ = v.BindEnv("mongodb.username", "MONGO_INITDB_ROOT_USERNAME")
	_ = v.BindEnv("mongodb.password", "MONGO_INITDB_ROOT_PASSWORD")
	_ = v.BindEnv("mongodb.host", "MONGODB_HOST")
	_ = v.BindEnv("mongodb.database", "MONGODB_DATABASE")
	_ = v.BindEnv("admin.username", "ADMIN_USERNAME")
	_ = v.BindEnv("admin.password", "ADMIN_PASSWORD")
	_ = v.BindEnv("admin.jwt_secret", "JWT_SECRET")
	_ = v.BindEnv("admin.jwt_ttl", "JWT_TTL")
	_ = v.BindEnv("ejudge.base_url", "EJUDGE_BASE_URL")
	_ = v.BindEnv("ejudge.api_key", "EJUDGE_API_KEY")
	_ = v.BindEnv("ejudge.timeout", "EJUDGE_TIMEOUT")
	_ = v.BindEnv("jplag.jar_path", "JPLAG_JAR_PATH")
	_ = v.BindEnv("analyze.jobs_max_concurrent", "JOBS_MAX_CONCURRENT")
	_ = v.BindEnv("analyze.analyze_concurrency", "ANALYZE_CONCURRENCY")
	_ = v.BindEnv("aiusage.enabled", "AIUSAGE_ENABLED")
	_ = v.BindEnv("openai.base_url", "OPENAI_BASE_URL")
	_ = v.BindEnv("openai.username", "OPENAI_USERNAME")
	_ = v.BindEnv("openai.password", "OPENAI_PASSWORD")
	_ = v.BindEnv("openai.api_key", "OPENAI_API_KEY")
	_ = v.BindEnv("openai.timeout", "OPENAI_TIMEOUT")
	_ = v.BindEnv("openai.model", "OPENAI_MODEL")
}
