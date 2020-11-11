package main

import (
	"context"
	"github.com/go-pg/pg/v10"
	"github.com/go-pg/pg/v10/orm"
)

type dbSession struct {
	*pg.DB
}

type dbConfig struct {
	Addr     string
	Database string
	User     string
	Pass     string
}

func init() {
	orm.RegisterTable((*UsersRoles)(nil))
	orm.RegisterTable((*UsersUsergroups)(nil))
}

func connectDB(state *State) {
	state.dbsession = &dbSession{pg.Connect(&pg.Options{
		Addr:     state.Dbconf.Addr,
		Database: state.Dbconf.Database,
		User:     state.Dbconf.User,
		Password: state.Dbconf.Pass,
	})}
	fileLog(state.logFile, "Connected to DB")
}

// Checks whether there is an active DB Connection
func checkConnection(session *dbSession) error {
	ctx := context.Background()
	err := session.Ping(ctx)
	return err
}

func getAllUsers(session *dbSession) ([]User, error) {
	var users []User
	err := session.Model(&users).Select()
	if err != nil {
		panic(err)
	}
	return users, nil
}

func getAllUsersGroupsWithUsers(session *dbSession) ([]Usergroup, error) {
	var groups []Usergroup
	err := session.Model(&groups).
		Relation("Users").
		Select()
	if err != nil {
		panic(err)
	}
	return groups, nil
}

func getAllRolesWithUsers(session *dbSession) ([]Role, error) {
	var roles []Role
	err := session.Model(&roles).
		Relation("Users").
		Select()
	if err != nil {
		panic(err)
	}
	return roles, nil
}