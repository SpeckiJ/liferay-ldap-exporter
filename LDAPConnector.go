package main

import (
	"fmt"
	"gopkg.in/ldap.v2"
)

type ldapSession struct {
	*ldap.Conn
}

type ldapConfig struct {
	Addr           string
	PasswordPrefix string
	AdminDN        string
	AdminPass      string
	DN             string
}

// pLDAPConnectAdmin binds to LDAP with editing permissions
func connectLDAP(state *State) error {
	conn, err := ldap.Dial("tcp", state.Ldapconf.Addr)
	if err != nil {
		return err
	}
	// Bind with Admin credentials
	err = conn.Bind(state.Ldapconf.AdminDN, state.Ldapconf.AdminPass)
	state.ldapsession = &ldapSession{conn}
	fileLog(state.logFile, "Connected to LDAP")
	return nil
}

func upsertLDAPOrganizationalUnit(state *State) error {
	err = insertIfNotExistsOrganizationalUnit(state, "usergroups")
	if err != nil {
		return err
	}
	err = insertIfNotExistsOrganizationalUnit(state, "users")
	if err != nil {
		return err
	}
	err = insertIfNotExistsOrganizationalUnit(state, "roles")
	if err != nil {
		return err
	}
	return nil
}

func insertIfNotExistsOrganizationalUnit(state *State, name string) error {
	searchRequest := ldap.NewSearchRequest(
		state.Ldapconf.DN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectClass=organizationalUnit)(ou=%s))", name),
		[]string{"dn"},
		nil,
	)
	sr, err := state.ldapsession.Search(searchRequest)
	if err != nil {
		return err
	}
	// only e
	if len(sr.Entries) == 0 {
		ar := ldap.NewAddRequest("ou=" + name + "," + state.Ldapconf.DN)
		ar.Attribute("objectClass", []string{"top", "organizationalUnit"})
		ar.Attribute("ou", []string{name})
		err = state.ldapsession.Add(ar)
		if err != nil {
			return err
		}
	}
	return nil
}

// LDAPAddUser adds User with given dn to LDAP
func upsertLDAPUser(state *State, user User) error {
	searchRequest := ldap.NewSearchRequest(
		state.Ldapconf.DN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectClass=organizationalPerson)(cn=%s))", user.Screenname),
		[]string{"dn"},
		nil,
	)
	sr, err := state.ldapsession.Search(searchRequest)
	if err != nil {
		return err
	}

	dn := "cn=" + user.Screenname + ",ou=users," + state.Ldapconf.DN
	if len(sr.Entries) > 0 {
		// Update existing User
		ar := ldap.NewModifyRequest(dn)
		ar.Replace("sn", []string{user.Lastname})
		ar.Replace("givenName", []string{user.Firstname})
		ar.Replace("mail", []string{user.Email})
		ar.Replace("userPassword", []string{user.Password})
		err = state.ldapsession.Modify(ar)
		return nil
	} else {
		// Create new User
		ar := ldap.NewAddRequest(dn)
		ar.Attribute("objectclass", []string{"inetOrgPerson", "person", "top", "organizationalPerson"})
		ar.Attribute("cn", []string{user.Screenname})
		ar.Attribute("sn", []string{user.Lastname})
		ar.Attribute("givenName", []string{user.Firstname})
		ar.Attribute("mail", []string{user.Email})
		ar.Attribute("userPassword", []string{user.Password})
		err = state.ldapsession.Add(ar)
		return err
	}
}

func upsertLDAPGroupOfNames(state *State, name, description, ou string, users []User) error {
	searchRequest := ldap.NewSearchRequest(
		state.Ldapconf.DN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectClass=groupOfNames)(cn=%s))", name),
		[]string{"dn"},
		nil,
	)
	sr, err := state.ldapsession.Search(searchRequest)
	if err != nil {
		return err
	}

	dn := "cn=" + name + ",ou=" + ou + "," + state.Ldapconf.DN
	if len(sr.Entries) > 0 {
		// Update existing
		ar := ldap.NewModifyRequest(dn)
		if description != "" {
			ar.Replace("description", []string{description})
		}
		err = state.ldapsession.Modify(ar)
		return err
	} else {
		// Create new
		ar := ldap.NewAddRequest(dn)
		ar.Attribute("objectclass", []string{"groupOfNames", "top"})
		ar.Attribute("cn", []string{name})
		if description != "" {
			ar.Attribute("description", []string{description})
		}
		// As Member is required attribute we add the admin to all groups
		members := []string{state.Ldapconf.AdminDN}
		for _, u := range users {
			members = append(members, "cn="+u.Screenname+",ou=users,"+state.Ldapconf.DN)
		}
		ar.Attribute("member", members)
		err = state.ldapsession.Add(ar)
		return err
	}
}
