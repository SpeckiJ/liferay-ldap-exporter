package main

import (
	"errors"
	"fmt"
	"fyne.io/fyne"
	"fyne.io/fyne/app"
	"fyne.io/fyne/container"
	"fyne.io/fyne/dialog"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
	"log"
)

func startGUI(state *State) {
	a := app.New()
	a.Settings().SetTheme(theme.LightTheme())
	w := a.NewWindow("Liferay LDAP Exporter")

	var hasValidDBConf, hasValidLDAPConf bool

	dbPass := widget.NewPasswordEntry()
	dbUser := widget.NewEntry()
	dbSchema := widget.NewEntry()
	dbAddr := widget.NewEntry()
	dbDatabase := widget.NewEntry()

	ldapPass := widget.NewPasswordEntry()
	ldapUser := widget.NewEntry()
	ldapAddr := widget.NewEntry()
	ldapDN := widget.NewEntry()
	ldapPasswordPrefix := widget.NewEntry()

	// DEBUG
	dbPass.SetText("sa")
	dbUser.SetText("sa")
	dbSchema.SetText("public")
	dbAddr.SetText("localhost:6432")
	dbDatabase.SetText("liferay")

	ldapPass.SetText("admin")
	ldapUser.SetText("cn=admin,dc=52North,dc=org")
	ldapAddr.SetText("localhost:389")
	ldapDN.SetText("dc=52North,dc=org")
	ldapPasswordPrefix.SetText("{MD5}")

	consoleOut := widget.NewLabel("")
	consoleOutScroller := container.NewScroll(consoleOut)
	transferProgress := widget.NewProgressBar()

	var users []User_

	insertUsersToLDAP := widget.NewButton("Insert LDAP users", func() {

		groups, err := getAllUsersGroupsWithUsers(state.dbsession)
		if err != nil {
			log.Fatal(err)
		}
		for _, group := range groups {
			fmt.Print(group)
			//err = upsertLDAPGroupOfNames(state, group)
		}
		transferProgress.Max = float64(len(users) + len(groups))
		for i, user := range users {
			err = upsertLDAPUser(state, user)
			if err != nil {
				state.errC <- err.Error()
				return
			}
			transferProgress.SetValue(float64(i + 1))
		}
	})
	insertUsersToLDAP.Disable()

	loadUsers := widget.NewButton("Load Users", func() {
		//users, err = getAllUsersWithRoles(state.dbsession)
		if err != nil {
			state.errC <- "error getting users"
		}
		for _, user := range users {
			state.errC <- "Read DB User_:" + user.String()
			insertUsersToLDAP.Enable()
		}
	})
	loadUsers.Disable()

	configPanes := container.NewAdaptiveGrid(2,
		widget.NewCard("Database Config",
			"",
			container.NewVBox(
				widget.NewLabel("Database URI"), dbAddr,
				widget.NewLabel("Database Name"), dbDatabase,
				widget.NewLabel("Database Schema Name"), dbSchema,
				widget.NewLabel("DB Admin Username"), dbUser,
				widget.NewLabel("DB Admin Password"), dbPass,
				widget.NewButton("Verify DB connection", func() {
					state.Dbconf = dbConfig{
						Addr:     dbAddr.Text,
						Database: dbDatabase.Text,
						User:     dbUser.Text,
						Pass:     dbPass.Text,
						Schema:   dbSchema.Text,
					}
					connectDB(state)
					err = checkConnection(state)
					if err != nil {
						state.errC <- "DB Connection Failed: " + err.Error()
					} else {
						hasValidDBConf = true
						state.errC <- "DB Connection established"
					}
				}),
			),
		),
		widget.NewCard("LDAP Config",
			"",
			container.NewVBox(
				widget.NewLabel("LDAP URI"), ldapAddr,
				widget.NewLabel("LDAP Password Prefix"), ldapPasswordPrefix,
				widget.NewLabel("LDAP DN"), ldapDN,
				widget.NewLabel("LDAP Admin Username"), ldapUser,
				widget.NewLabel("LDAP Admin Password"), ldapPass,
				widget.NewButton("Verify LDAP connection", func() {
					state.Ldapconf = ldapConfig{
						Addr:           ldapAddr.Text,
						PasswordPrefix: ldapPasswordPrefix.Text,
						AdminDN:        ldapUser.Text,
						AdminPass:      ldapPass.Text,
						DN:             ldapDN.Text,
					}
					err = connectLDAP(state)
					if err != nil {
						state.errC <- "LDAP Connection failed:" + err.Error()
					} else {
						hasValidLDAPConf = true
						insertUsersToLDAP.Enable()
					}
				}),
			),
		),
	)
	content := container.NewVBox(
		configPanes,
		widget.NewButton("Transform DB Users to LDAP Users", func() {
			transformDBToLDAP(state)
		}),
		consoleOutScroller,
		transferProgress,
	)

	footer := widget.NewToolbar(
		widget.NewToolbarAction(theme.ErrorIcon(), func() {}),
		widget.NewToolbarSeparator(),
		widget.NewToolbarAction(theme.ColorPaletteIcon(), func() {}),
	)

	exportConfigMenu := fyne.NewMenuItem("Export Config", func() {
		if hasValidLDAPConf {
			if hasValidDBConf {
				exportConfig(state)
			} else {
				dialog.NewError(errors.New("Cannot export: DB connection is not verified! Please validate first"), w)
			}
		} else {
			dialog.NewError(errors.New("Cannot export: LDAP connection is not verified! Please validate first"), w)
		}
	})

	w.SetMainMenu(fyne.NewMainMenu(fyne.NewMenu("File", exportConfigMenu)))
	w.SetContent(
		fyne.NewContainerWithLayout(layout.NewBorderLayout(content, footer, nil, nil),
			content,
			footer,
		),
	)

	consoleOutScroller.SetMinSize(configPanes.Size().Subtract(fyne.NewSize(0, 100)))
	w.ShowAndRun()
}
