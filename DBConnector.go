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
	Schema   string
	User     string
	Pass     string
}

type dbLogger struct {
	errC chan string
}

func init() {
	orm.RegisterTable((*UsersRoles)(nil))
	orm.RegisterTable((*UsersUsergroups)(nil))
}

func (d dbLogger) BeforeQuery(c context.Context, q *pg.QueryEvent) (context.Context, error) {
	return c, nil
}

func (d dbLogger) AfterQuery(c context.Context, q *pg.QueryEvent) error {
	query, err := q.FormattedQuery()
	if err == nil {
		d.errC <- string(query)
	} else {
		d.errC <- err.Error()
	}
	return nil
}

func connectDB(state *State) {
	orm.SetTableNameInflector(func(s string) string {
		return s
	})

	state.dbsession = &dbSession{pg.Connect(&pg.Options{
		Addr:     state.Dbconf.Addr,
		Database: state.Dbconf.Database,
		User:     state.Dbconf.User,
		Password: state.Dbconf.Pass,
		PoolSize: 1,
		OnConnect: func(ctx context.Context, conn *pg.Conn) error {
			_, err := conn.Exec("set search_path=?", state.Dbconf.Schema)
			if err != nil {
				state.errC <- err.Error()
			}
			return nil
		},
	})}

	if verbose {
		state.dbsession.AddQueryHook(dbLogger{errC: state.errC})
	}
}

// Checks whether there is an active DB Connection
func checkConnection(state *State) error {
	ctx := context.Background()
	err := state.dbsession.Ping(ctx)
	if err == nil {
		state.errC <- "Connected to DB!"
	} else {
		state.errC <- "Could not connect to DB:" + err.Error()
	}
	return err
}

func getAllUsers(session *dbSession) ([]User_, error) {
	var users []User_
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

func getAllRolesWithUsers(session *dbSession) ([]Role_, error) {
	var roles []Role_
	err := session.Model(&roles).
		Relation("Users").
		Select()
	if err != nil {
		panic(err)
	}
	return roles, nil
}
