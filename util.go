package main

import "fmt"

type User_ struct {
	Userid     int      `pg:"userid,pk"`
	Password   string   `pg:"password_"`
	Screenname string   `pg:"screenname"`
	Email      string   `pg:"emailaddress"`
	Firstname  string   `pg:"firstname"`
	Middlename string   `pg:"middlename"`
	Lastname   string   `pg:"lastname"`
}

type Role_ struct {
	Roleid      int      `pg:"roleid,pk"`
	Name        string   `pg:"name"`
	Description string   `pg:"description"`
	Users       []User_  `pg:"many2many:users_roles"`
}

type Usergroup struct {
	Groupid     int      `pg:"usergroupid,pk"`
	Name        string   `pg:"name"`
	Description string   `pg:"description"`
	Users       []User_  `pg:"many2many:users_usergroups"`
}

type UsersUsergroups struct {
	Userid      int      `pg:"userid,pk"`
	Usergroupid int      `pg:"usergroupid,pk"`
}

type UsersRoles struct {
	Roleid int `pg:"roleid,pk"`
	Userid int `pg:"userid,pk"`
}

func (u User_) String() string {
	return fmt.Sprintf("User_ <Screenname: %s, Email: %s, Password: %s>", u.Screenname, u.Email, "************")
}

func (r Role_) String() string {
	return fmt.Sprintf("Name <Screenname: %s>", r.Name)
}
