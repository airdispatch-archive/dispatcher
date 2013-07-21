package main

import (
	"flag"
	"fmt"
	"github.com/airdispatch/dispatcher/models"
	"github.com/airdispatch/dispatcher/views"
	library "github.com/airdispatch/go-pressure"
	_ "github.com/lib/pq"
	"os"
	// "airdispat.ch/common"
)

var flag_port = flag.String("port", "2048", "specify the port that the server should run on")
var db_flush = flag.Bool("db_flush", false, "specifies to flush the database")
var db_create = flag.Bool("db_create", false, "specifies to create the database")
var me = flag.String("me", "", "specify the location of this running server")

func main() {
	flag.Parse()

	temp_port := os.Getenv("PORT")
	if temp_port == "" {
		temp_port = *flag_port
	}

	dbMap, err := models.ConnectToDB()
	if err != nil {
		fmt.Println("Can't Connect to DB")
		fmt.Println("err")
		return
	}

	theServer := &library.Server{
		Port:            temp_port,
		DbMap:           dbMap,
		MainSessionName: "dispatcher-session",
		Mailserver:      *me,
	}

	// Flush the Database
	if *db_flush {
		theServer.DbMap.DropTables()
	}

	// Create the Database
	if *db_create {
		// Create Tables
		err := theServer.DbMap.CreateTablesIfNotExists()
		if err != nil {
			fmt.Println("Problem Creating Tables")
			fmt.Println(err)
			return
		}

		// The Tracker
		fmt.Println("Time to setup the default Trackers")

		for {
			var t_url, t_address string
			fmt.Print("Tracker URL (or 'done' to stop): ")
			fmt.Scanln(&t_url)

			if t_url == "done" {
				break
			}

			fmt.Print("Tracker Address: ")
			fmt.Scanln(&t_address)

			t_full := &models.Tracker{URL: t_url, Address: t_address}
			theServer.DbMap.Insert(t_full)
		}
		// Create the User
		fmt.Println("Let's create the first user.")

		var username, password, first, last string

		fmt.Print("Username (no spaces): ")
		fmt.Scanln(&username)

		fmt.Print("Password: ")
		fmt.Scanln(&password)

		fmt.Print("First Name: ")
		fmt.Scanln(&first)
		fmt.Print("Last Name: ")
		fmt.Scanln(&last)

		newUser := models.CreateUser(username, password, theServer)
		newUser.FullName = (first + " " + last)

		fmt.Println("New User Address", newUser.Address)
		theServer.DbMap.Insert(newUser)
	}

	theServer.ConfigServer()
	defineRoutes(theServer)

	theServer.RunServer()
}

func defineRoutes(s *library.Server) {
	s.WebServer.Get("/", views.HomePage(s, views.Dashboard(s)))

	s.WebServer.Get("/compose", views.TemplateLoginRequired(s, s.DisplayTemplate("compose.html")))
	s.WebServer.Post("/compose", views.TemplateLoginRequired(s, views.CreateMessage(s)))

	s.WebServer.Get("/subscribe", views.TemplateLoginRequired(s, views.ShowSubscriptions(s)))
	s.WebServer.Post("/subscribe", views.TemplateLoginRequired(s, views.CreateSubscription(s)))

	s.WebServer.Get("/inbox", views.TemplateLoginRequired(s, views.ShowFolder(s, "Inbox")))
	s.WebServer.Get("/profile", views.TemplateLoginRequired(s, s.DisplayTemplate("account/profile.html")))

	s.WebServer.Get("/alert/([0-9]*)", views.WildcardTemplateLoginRequired(s, views.ShowAlert(s)))

	s.WebServer.Get("/message/([0-9]*)", views.WildcardTemplateLoginRequired(s, views.ShowMessage(s)))

	s.WebServer.Get("/message/([0-9]*)/edit", views.WildcardTemplateLoginRequired(s, views.DisplayEditMessage(s)))
	s.WebServer.Post("/message/([0-9]*)/edit", views.WildcardTemplateLoginRequired(s, views.UpdateMessage(s)))

	s.WebServer.Get("/message/([0-9]*)/delete", views.WildcardTemplateLoginRequired(s, views.DeleteMessage(s)))

	s.WebServer.Get("/account/tracker_registration", views.TemplateLoginRequired(s, views.RegisterWithTrackers(s)))

	s.WebServer.Get("/login", s.DisplayTemplate("login.html"))
	s.WebServer.Post("/login", views.LoginView(s))

	s.WebServer.Post("/account/create", views.RegisterUser(s))

	s.WebServer.Get("/logout", views.TemplateLoginRequired(s, views.LogoutView(s)))
}
