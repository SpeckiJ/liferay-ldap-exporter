package main

import "fmt"

type User struct {
	tableName  struct{} `sensorweb2:"user_"`
	Userid     int      `sensorweb2:"userid,pk"`
	Password   string   `sensorweb2:"password_"`
	Screenname string   `sensorweb2:"screenname"`
	Email      string   `sensorweb2:"emailaddress"`
	Firstname  string   `sensorweb2:"firstname"`
	Middlename string   `sensorweb2:"middlename"`
	Lastname   string   `sensorweb2:"lastname"`
}

type Role struct {
	tableName   struct{} `sensorweb2:"role_"`
	Roleid      int      `sensorweb2:"roleid,pk"`
	Name        string   `sensorweb2:"name"`
	Description string   `sensorweb2:"description"`
	Users       []User   `sensorweb2:"many2many:users_roles"`
}

type Usergroup struct {
	tableName   struct{} `sensorweb2:"usergroup"`
	Groupid     int      `sensorweb2:"usergroupid,pk"`
	Name        string   `sensorweb2:"name"`
	Description string   `sensorweb2:"description"`
	Users       []User   `sensorweb2:"many2many:users_usergroups"`
}

type UsersUsergroups struct {
	tableName   struct{} `sensorweb2:"users_usergroups"`
	Userid      int      `sensorweb2:"userid,pk"`
	Usergroupid int      `sensorweb2:"usergroupid,pk"`
}

type UsersRoles struct {
	// tableName struct{} `sensorweb2:"users_roles"`
	Roleid int `sensorweb2:"roleid,pk"`
	Userid int `sensorweb2:"userid,pk"`
}

func (u User) String() string {
	return fmt.Sprintf("User <Screenname: %s, Email: %s, Password: %s>", u.Screenname, u.Email, "************")
}

func (r Role) String() string {
	return fmt.Sprintf("Name <Screenname: %s>", r.Name)
}
