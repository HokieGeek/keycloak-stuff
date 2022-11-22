# keycloak-stuff

## installing go on mac

```shell
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```
```shell
brew install go
```
## programs
### add-to-keycloak-group

#### install
```shell
go install github.com/hokiegeek/keycloak-stuff/cmd/add-to-keycloak-group@latest
```

#### run
```shell
add-to-keycloak-group -config CONFIG_FILE.yaml -users USERS_CSV.csv -group GROUPNAME
```