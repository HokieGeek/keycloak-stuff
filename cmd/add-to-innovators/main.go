package main

import (
	"context"
	"encoding/csv"
	"errors"
	"flag"
	"fmt"
	"github.com/Nerzal/gocloak/v11"
	"github.com/goccy/go-yaml"
	"github.com/hashicorp/go-multierror"
	"io/ioutil"
	"os"
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
	configFilePtr := flag.String("config", "", "the yaml file with the Keycloak config")
	usersFilePtr := flag.String("users", "", "the CSV with the users to add")
	groupNamePtr := flag.String("group", "", "the name of the Keycloak group to add the users")

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

	groups, err := client.GetGroups(ctx, tokenResp.AccessToken, cfg.Keycloak.Realm, gocloak.GetGroupsParams{Search: groupNamePtr})
	if err != nil {
		panic(err)
	}
	if len(groups) > 1 {
		panic("doh")
	}
	innovatorsGroupID := *groups[0].ID

	fmt.Printf("Adding users to '%s' group:\n", *groupNamePtr)
	var addErrors error
	for _, user := range users[1:] {
		fmt.Print("  ", user[1])
		users, err := client.GetUsers(ctx, tokenResp.AccessToken, cfg.Keycloak.Realm, gocloak.GetUsersParams{Email: &user[1]})
		if err != nil {
			fmt.Printf("\terror: %v\n", err)
			addErrors = multierror.Append(addErrors, err)
			continue
		}
		if len(users) > 1 {
			fmt.Printf("\terror: %v\n", err)
			addErrors = multierror.Append(addErrors, fmt.Errorf("found too manu users for email %s", user[1]))
			continue
		}
		userGroups, err := client.GetUserGroups(ctx, tokenResp.AccessToken, cfg.Keycloak.Realm, *users[0].ID, gocloak.GetGroupsParams{Search: groupNamePtr})
		if err != nil {
			fmt.Printf("\terror: %v\n", err)
			addErrors = multierror.Append(addErrors, err)
			continue
		}
		var found bool
		for _, g := range userGroups {
			if *g.Name == *groupNamePtr {
				found = true
				fmt.Println("\talready in group")
				break
			}
		}
		if found {
			continue
		}
		if err := client.AddUserToGroup(ctx, tokenResp.AccessToken, cfg.Keycloak.Realm, *users[0].ID, innovatorsGroupID); err != nil {
			fmt.Printf("\terror: %v\n", err)
			addErrors = multierror.Append(addErrors, err)
			continue
		}
		fmt.Println("\tADDED")
	}

	if addErrors != nil {
		fmt.Println(errors.Unwrap(addErrors))
	}
}
