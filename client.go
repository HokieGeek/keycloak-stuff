package keycloak_stuff

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/Nerzal/gocloak/v11"
	"github.com/goccy/go-yaml"
)

type Config struct {
	Keycloak keycloakConfig `yaml:"keycloak"`
}

type keycloakConfig struct {
	BaseURL      string `yaml:"baseurl"`
	Realm        string `yaml:"realm"`
	ClientName   string `yaml:"client_name"`
	ClientSecret string `yaml:"client_secret"`
}

func LoadConfigFromFile(name string) (Config, error) {
	configFile, err := os.Open(name)
	if err != nil {
		return Config{}, nil
	}
	defer configFile.Close()

	buf, err := io.ReadAll(configFile)
	if err != nil {
		return Config{}, nil
	}

	var cfg Config
	if err := yaml.Unmarshal(buf, &cfg); err != nil {
		return Config{}, nil
	}

	return cfg, nil
}

type Klient struct {
	Client       gocloak.GoCloak
	Ctx          context.Context
	Realm, Token string
}

func New(cfg Config) (*Klient, error) {

	k := new(Klient)
	k.Client = gocloak.NewClient(cfg.Keycloak.BaseURL)
	k.Ctx = context.Background()
	k.Realm = cfg.Keycloak.Realm
	tokenResp, err := k.Client.LoginClient(k.Ctx, cfg.Keycloak.ClientName, cfg.Keycloak.ClientSecret, cfg.Keycloak.Realm)
	if err != nil {
		return nil, fmt.Errorf("could not get token: %v", err)
	}
	k.Token = tokenResp.AccessToken

	return k, nil
}
