package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"io/ioutil"
	"os"

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

	actions := flag.Args()

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

	badstuff := make([]error, 0)
	for _, user := range users[1:] {
		fmt.Print("  ", user[2])
		if users, err := client.GetUsers(ctx, tokenResp.AccessToken, cfg.Keycloak.Realm, gocloak.GetUsersParams{Email: gocloak.StringP(user[2])}); err != nil {
			badstuff = append(badstuff, fmt.Errorf("%s: %v", user[2], err))
			fmt.Println(" [ERROR]")
		} else {
			fmt.Printf(" [%s]\n", *users[0].ID)

			users[0].RequiredActions = &actions
			if err := client.UpdateUser(ctx, tokenResp.AccessToken, cfg.Keycloak.Realm, *users[0]); err != nil {
				badstuff = append(badstuff, fmt.Errorf("%s: %v", user[2], err))
			}
		}
	}

	// TODO: add to organization

	if len(badstuff) > 0 {
		fmt.Println("The following users could not be modified:")
		for _, bad := range badstuff {
			fmt.Println("  ", bad)
		}
	}
}
