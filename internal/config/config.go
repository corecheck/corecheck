package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

var Config ConfigT

type ConfigT struct {
	Dev     bool `yaml:"dev" env:"DEV" env-default:"false"`
	Autorun bool `yaml:"autorun" env:"AUTORUN" env-default:"false"`
	API     struct {
		Port          string `yaml:"port" env:"API_PORT" env-default:"8000"`
		SessionSecret string `yaml:"session_secret" env:"API_SESSION_SECRET"`
		CookieSecure  bool   `yaml:"cookie_secure" env:"API_COOKIE_SECURE" env-default:"false"`
	} `yaml:"api"`
	DB struct {
		Host     string `yaml:"host" env:"DB_HOST" env-default:"localhost"`
		Port     string `yaml:"port" env:"DB_PORT" env-default:"5432"`
		User     string `yaml:"user" env:"DB_USER"`
		Password string `yaml:"password" env:"DB_PASSWORD"`
		Name     string `yaml:"name" env:"DB_NAME" env-default:"corecheck"`
	} `yaml:"db"`

	Github struct {
		AccessToken string `yaml:"access_token" env:"GH_ACCESS_TOKEN"`

		OauthClientID    string `yaml:"oauth_client_id" env:"GH_OAUTH_CLIENT_ID"`
		OauthSecret      string `yaml:"oauth_secret" env:"GH_OAUTH_SECRET"`
		OauthRedirectURL string `yaml:"oauth_redirect_url" env:"GH_OAUTH_REDIRECT_URL"`
	} `yaml:"github"`
}

func Load(filename string) error {
	if filename == "" {
		filename = "config.yml"
	}
	err := cleanenv.ReadConfig(filename, &Config)
	if err != nil {
		fmt.Println(err)
	}

	return cleanenv.ReadEnv(&Config)
}
