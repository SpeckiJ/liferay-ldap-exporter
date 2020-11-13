package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

const (
	ERROR = "ERROR"
	INFO  = "INFO"
)

var err error
var verbose bool

type State struct {
	// Exported to JSON when state is exported
	Dbconf   dbConfig
	Ldapconf ldapConfig

	// runtime-specific
	dbsession   *dbSession
	ldapsession *ldapSession
	errC        chan string
}

func main() {
	headlessMode := flag.Bool("headless", false, "Starts the client in Headless Mode")
	configFileName := flag.String("config", "", "Config file to be loaded")
	flag.BoolVar(&verbose, "verbose", false, "Toggle verbose logging to console")
	flag.Parse()

	logFile := openLogFile()
	errC := make(chan string, 100)
	go fileLog(logFile, errC)

	if *headlessMode {
		if *configFileName != "" {
			startHeadless(&State{errC: errC}, configFileName)
		} else {
			errC <- "Could not start in headless mode: Missing config File!"
			return
		}
	} else {
		startGUI(&State{errC: errC})
	}
}
func startHeadless(state *State, confFileName *string) {
	importConfig(state, *confFileName)
	connectDB(state)
	err = connectLDAP(state)
	if err != nil {
		state.errC <- "Could not connect to LDAP: " + err.Error()
		return
	}
	transformDBToLDAP(state)
}

func transformDBToLDAP(state *State) {
	users, err := getAllUsers(state.dbsession)
	if err != nil {
		state.errC <- "Could not get all users: " + err.Error()
		return
	}
	groups, err := getAllUsersGroupsWithUsers(state.dbsession)
	if err != nil {
		state.errC <- "Could not get all usergroups: " + err.Error()
		return
	}
	roles, err := getAllRolesWithUsers(state.dbsession)
	if err != nil {
		state.errC <- "Could not get all roles: " + err.Error()
		return
	}

	err = upsertLDAPOrganizationalUnit(state)
	if err != nil {
		state.errC <- "Could not upsert organizationalUnit: " + err.Error()
	}
	for _, u := range users {
		err = upsertLDAPUser(state, u)
		if err != nil {
			state.errC <- fmt.Sprintf("Failed to upsert user <%s> with error: %s", u.Screenname, err.Error())
		}
	}
	for _, r := range roles {
		err = upsertLDAPGroupOfNames(state, r.Name, r.Description, "roles", users)
		if err != nil {
			state.errC <- fmt.Sprintf("Failed to upsert role <%s> with error: %s", r.Name, err.Error())
		}
	}
	for _, g := range groups {
		err = upsertLDAPGroupOfNames(state, g.Name, g.Description, "usergroups", users)
		if err != nil {
			state.errC <- fmt.Sprintf("Failed to upsert usergroup <%s> with error: %s", g.Name, err.Error())
		}
	}
}

func importConfig(state *State, confFileName string) {
	config, err := ioutil.ReadFile(confFileName)
	if err != nil {
		state.errC <- "Could not read configFile: " + err.Error()
		return
	}
	err = json.Unmarshal(config, state)
	if err != nil {
		state.errC <- "Could not parse configFile: " + err.Error()
		return
	}
}

func exportConfig(state *State) {
	configOut, err := os.OpenFile("config.conf", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		state.errC <- "Could not open logfile: " + err.Error()
	}
	jsonConfig, err := json.MarshalIndent(state, "", "    ")
	if err != nil {
		state.errC <- err.Error()
	}
	_, err = configOut.Write(jsonConfig)
	if err != nil {
		//TODO: exit nicely as we export from GUI
		state.errC <- err.Error()
	}
}
