package main

import "fmt"

type User struct {
	tableName  struct{} `pg:"user_"`
	Userid     int      `pg:"userid,pk"`
	Password   string   `pg:"password_"`
	Screenname string   `pg:"screenname"`
	Email      string   `pg:"emailaddress"`
	Firstname  string   `pg:"firstname"`
	Middlename string   `pg:"middlename"`
	Lastname   string   `pg:"lastname"`
}

type Role struct {
	tableName   struct{} `pg:"role_"`
	Roleid      int      `pg:"roleid,pk"`
	Name        string   `pg:"name"`
	Description string   `pg:"description"`
	Users       []User   `pg:"many2many:users_roles"`
}

type Usergroup struct {
	tableName   struct{} `pg:"usergroup"`
	Groupid     int      `pg:"usergroupid,pk"`
	Name        string   `pg:"name"`
	Description string   `pg:"description"`
	Users       []User   `pg:"many2many:users_usergroups"`
}

type UsersUsergroups struct {
	tableName   struct{} `pg:"users_usergroups"`
	Userid      int      `pg:"userid,pk"`
	Usergroupid int      `pg:"usergroupid,pk"`
}

type UsersRoles struct {
	// tableName struct{} `pg:"users_roles"`
	Roleid int `pg:"roleid,pk"`
	Userid int `pg:"userid,pk"`
}

func (u User) String() string {
	return fmt.Sprintf("User <Screenname: %s, Email: %s, Password: %s>", u.Screenname, u.Email, "************")
}

func (r Role) String() string {
	return fmt.Sprintf("Name <Screenname: %s>", r.Name)
}
