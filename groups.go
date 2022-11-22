package keycloak_stuff

import (
	"fmt"

	"github.com/Nerzal/gocloak/v11"
	"github.com/hashicorp/go-multierror"
)

func addUserToGroupByID(k *Klient, user, groupName, groupID string) (bool, error) {
	users, err := k.Client.GetUsers(k.Ctx, k.Token, k.Realm, gocloak.GetUsersParams{Email: &user})
	if err != nil {
		fmt.Printf("\terror: %v\n", err)
		return false, err
	}
	if len(users) > 1 {
		return false, err
	}
	userGroups, err := k.Client.GetUserGroups(k.Ctx, k.Token, k.Realm, *users[0].ID, gocloak.GetGroupsParams{Search: &groupName})
	if err != nil {
		return false, err
	}
	var found bool
	for _, g := range userGroups {
		if *g.Name == groupName {
			found = true
			break
		}
	}
	if found {
		return false, nil
	}
	if err := k.Client.AddUserToGroup(k.Ctx, k.Token, k.Realm, *users[0].ID, groupID); err != nil {
		return false, err
	}

	return true, nil
}

func AddUserToGroup(k *Klient, user, groupName string) (bool, error) {
	group, err := GetGroup(k, groupName)
	if err != nil {
		return false, err
	}
	return addUserToGroupByID(k, user, groupName, *group.ID)
}

func GetGroup(k *Klient, groupName string) (*gocloak.Group, error) {
	groups, err := k.Client.GetGroups(k.Ctx, k.Token, k.Realm, gocloak.GetGroupsParams{Search: &groupName})
	if err != nil {
		return nil, err
	}
	if len(groups) > 1 {
		return nil, fmt.Errorf("too many groups found for name %s", groupName)
	}
	return groups[0], nil
}

func AddUsersToGroup(k *Klient, users []string, groupName string) error {
	group, err := GetGroup(k, groupName)
	if err != nil {
		return err
	}

	var addErrors error
	for _, user := range users[1:] {
		fmt.Print("  ", user)
		added, err := addUserToGroupByID(k, user, groupName, *group.ID)
		if err != nil {
			fmt.Printf("\t[error: %v]\n", err)
			addErrors = multierror.Append(addErrors, err)
			continue
		}
		if !added {
			fmt.Println("\t[already in group]")
			continue
		}
		fmt.Println("\t[ADDED]")
	}

	return addErrors
}
