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

type State struct {
	// Exported to JSON when state is exported
	Dbconf   dbConfig
	Ldapconf ldapConfig

	// runtime-specific
	dbsession   *dbSession
	ldapsession *ldapSession
	logFile     *os.File
}

func main() {
	headlessMode := flag.Bool("headless", false, "Starts the client in Headless Mode")
	configFileName := flag.String("config", "", "")
	flag.Parse()

	logFile := openLogFile()
	if *headlessMode {
		if *configFileName != "" {
			startHeadless(&State{logFile: logFile}, configFileName)
		} else {
			fileLog(logFile, "Could not start in headless mode: Missing config File!")
			return
		}
	} else {
		startGUI(&State{logFile: logFile})
	}
}
func startHeadless(state *State, confFileName *string) {
	importConfig(state, *confFileName)
	connectDB(state)
	err = connectLDAP(state)
	if err != nil {
		fileLog(state.logFile, "Could not connect to LDAP: "+err.Error())
		return
	}
	transformDBToLDAP(state)
}

func transformDBToLDAP(state *State) {
	users, err := getAllUsers(state.dbsession)
	if err != nil {
		fileLog(state.logFile, "Could not get all users: "+err.Error())
		return
	}
	groups, err := getAllUsersGroupsWithUsers(state.dbsession)
	if err != nil {
		fileLog(state.logFile, "Could not get all usergroups: "+err.Error())
		return
	}
	roles, err := getAllRolesWithUsers(state.dbsession)
	if err != nil {
		fileLog(state.logFile, "Could not get all roles: "+err.Error())
		return
	}

	err = upsertLDAPOrganizationalUnit(state)
	if err != nil {
		fileLog(state.logFile, "Could not upsert organizationalUnit: "+err.Error())
	}
	for _, u := range users {
		err = upsertLDAPUser(state, u)
		if err != nil {
			fileLog(state.logFile, fmt.Sprintf("Failed to upsert user <%s> with error: %s", u.Screenname, err.Error()))
		}
	}
	for _, r := range roles {
		err = upsertLDAPGroupOfNames(state, r.Name, r.Description, "roles", users)
		if err != nil {
			fileLog(state.logFile, fmt.Sprintf("Failed to upsert role <%s> with error: %s", r.Name, err.Error()))
		}
	}
	for _, g := range groups {
		err = upsertLDAPGroupOfNames(state, g.Name, g.Description, "usergroups", users)
		if err != nil {
			fileLog(state.logFile, fmt.Sprintf("Failed to upsert usergroup <%s> with error: %s", g.Name, err.Error()))
		}
	}
}

func importConfig(state *State, confFileName string) {
	config, err := ioutil.ReadFile(confFileName)
	if err != nil {
		fileLog(state.logFile, "Could not read configFile: "+err.Error())
		return
	}
	err = json.Unmarshal(config, state)
	if err != nil {
		fileLog(state.logFile, "Could not parse configFile: "+err.Error())
		return
	}
}

func exportConfig(state *State) {
	configOut, err := os.OpenFile("config.conf", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		fileLog(state.logFile, "Could not open logfile: "+err.Error())
	}
	jsonConfig, err := json.MarshalIndent(state, "", "    ")
	if err != nil {
		fileLog(state.logFile, err.Error())
	}
	_, err = configOut.Write(jsonConfig)
	if err != nil {
		//TODO: exit nicely as we export from GUI
		fileLog(state.logFile, err.Error())
	}
}
