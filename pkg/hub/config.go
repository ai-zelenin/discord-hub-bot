package hub

import (
	"flag"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"net"
)

type Config struct {
	*viper.Viper
}

func NewConfig() (*Config, error) {
	flag.String("c", "bot.yml", "define path to config file")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	v := viper.New()
	err := v.BindPFlags(pflag.CommandLine)
	if err != nil {
		return nil, err
	}
	v.SetEnvPrefix("BOT")
	v.AutomaticEnv()
	cfg := &Config{
		Viper: v,
	}
	cfg.Viper.SetConfigName(v.GetString("c"))
	cfg.Viper.SetConfigType("yaml")
	cfg.Viper.AddConfigPath(".")
	cfg.Viper.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("Config file changed:", e.Name)
	})
	cfg.Viper.WatchConfig()
	err = cfg.Viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("Fatal error config file: %w \n", err)
	}
	return cfg, nil
}

func (c *Config) Addr() string {
	return net.JoinHostPort(c.GetString("host"), c.GetString("port"))
}

func (c *Config) Token() string {
	return c.GetString("token")
}

func (c *Config) AppID() string {
	return c.GetString("app_id")
}

func (c *Config) GuildID() string {
	return c.GetString("guild_id")
}

func (c *Config) DB() string {
	return c.GetString("db_uri")
}

func (c *Config) Sources() map[string]*SourceConfig {
	m := make(map[string]*SourceConfig)
	for k := range c.GetStringMap("sources") {
		srcCfg := &SourceConfig{
			RequestType:     RequestType(c.GetString(fmt.Sprintf("sources.%s.request_type", k))),
			Template:        c.GetString(fmt.Sprintf("sources.%s.template", k)),
			ActionType:      ActionType(c.GetString(fmt.Sprintf("sources.%s.action_type", k))),
			FindPersonaExpr: c.GetString(fmt.Sprintf("sources.%s.find_persona_expr", k)),
			SendAsEmbed:     c.GetBool(fmt.Sprintf("sources.%s.send_as_embed", k)),
		}
		if srcCfg.RequestType == "" {
			srcCfg.RequestType = RequestTypePostJsonObj
		}
		if srcCfg.ActionType == "" {
			srcCfg.ActionType = ActionTypeProxy
		}
		m[k] = srcCfg

	}
	return m
}

type RequestType string

const (
	RequestTypeGet           = "get"
	RequestTypePostJsonObj   = "post-json-obj"
	RequestTypePostJsonArray = "post-json-array"
)

type ActionType string

const (
	ActionTypeProxy = "proxy"
)

type SourceConfig struct {
	RequestType     RequestType `json:"request_type" yaml:"request_type"`
	ActionType      ActionType  `json:"action_type" yaml:"action_type"`
	Template        string      `json:"template" yaml:"template"`
	FindPersonaExpr string      `json:"find_persona_expr" yaml:"find_persona_expr"`
	SendAsEmbed     bool        `json:"send_as_embed" yaml:"send_as_embed"`
}
