package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/Nerzal/gocloak/v11"
	"github.com/goccy/go-yaml"
)

type config struct {
	Keycloak keycloakConfig `yaml:"keycloak"`
}

type keycloakConfig struct {
	BaseURL      string `yaml:"baseurl"`
	Realm        string `yaml:"realm"`
	ClientName   string `yaml:"client_name"`
	ClientSecret string `yaml:"client_secret"`
}

func main() {
	configFilePtr := flag.String("config", "", "the yaml file with the keycloak config")
	usersFilePtr := flag.String("users", "", "the csv with the users to add")

	flag.Parse()

	// Load the config file
	configFile, err := os.Open(*configFilePtr)
	if err != nil {
		panic(err)
	}
	defer configFile.Close()

	buf, err := ioutil.ReadAll(configFile)
	if err != nil {
		panic(err)
	}

	var cfg config
	if err := yaml.Unmarshal(buf, &cfg); err != nil {
		panic(err)
	}

	// Load the list of users to create
	usersFile, err := os.Open(*usersFilePtr)
	if err != nil {
		panic(err)
	}
	defer usersFile.Close()

	users, err := csv.NewReader(usersFile).ReadAll()
	if err != nil {
		panic(err)
	}

	// Create the keycloak client
	client := gocloak.NewClient(cfg.Keycloak.BaseURL)
	ctx := context.Background()
	tokenResp, err := client.LoginClient(ctx, cfg.Keycloak.ClientName, cfg.Keycloak.ClientSecret, cfg.Keycloak.Realm)
	if err != nil {
		panic(fmt.Errorf("could not get token: %v", err))
	}

	// Add users to keycloak
	fmt.Println("Creating users:")
	badcreates := make([]error, 0)
	for _, user := range users[1:] {
		fmt.Println("  ", user[2])
		toCreate := gocloak.User{
			FirstName:     gocloak.StringP(user[0]),
			LastName:      gocloak.StringP(user[1]),
			Email:         gocloak.StringP(user[2]),
			Username:      gocloak.StringP(user[2]),
			Enabled:       gocloak.BoolP(true),
			EmailVerified: gocloak.BoolP(true),
			Credentials: &[]gocloak.CredentialRepresentation{{
				CreatedDate: gocloak.Int64P(time.Now().Unix()),
				Temporary:   gocloak.BoolP(true),
				Type:        gocloak.StringP("password"),
				Value:       gocloak.StringP(user[3]),
			}},
		}

		if _, err = client.CreateUser(ctx, tokenResp.AccessToken, cfg.Keycloak.Realm, toCreate); err != nil {
			badcreates = append(badcreates, fmt.Errorf("%s: %v", user[2], err))
		}
	}

	// TODO: add to organization

	if len(badcreates) > 0 {
		fmt.Println("The following users could not be created:")
		for _, bad := range badcreates {
			fmt.Println("  ", bad)
		}
	}
}
