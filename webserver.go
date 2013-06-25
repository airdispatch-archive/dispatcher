package main

import (
	_ "github.com/lib/pq"
	"os"
	"flag"
	"fmt"
	"log"
	"github.com/coopernurse/gorp"
	"dispatcher/library"
	"dispatcher/views"
	"dispatcher/models"
	// "airdispat.ch/common"
)

var flag_port = flag.String("port", "2048", "specify the port that the server should run on")
var db_flush = flag.Bool("db_flush", false, "specifies to flush the database")
var db_create = flag.Bool("db_create", false, "specifies to create the database")

func main() {
	flag.Parse()

	temp_port := os.Getenv("PORT")
	if temp_port == "" {
		temp_port = *flag_port
	}

	theServer := &library.Server {
		Port: temp_port,
		DbMap: connectToDatabase(),
		CookieAuthKey: []byte("secret-auth"),
		CookieEncryptKey: []byte("secret-encryption-key"),
		MainSessionName: "dispatcher-session",
		Mailserver: "mailserver.airdispat.ch:2048",
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

		var username, password string

		fmt.Print("Username (no spaces): ")
		fmt.Scanln(&username)

		fmt.Print("Password: ")
		fmt.Scanln(&password)

		newUser := models.CreateUser(username, password, theServer)
		fmt.Println("New User Address", newUser.Address)
		theServer.DbMap.Insert(newUser)
	}

	theServer.ConfigServer()
	defineRoutes(theServer)

	theServer.RunServer()
}

func defineRoutes(s *library.Server) {
	s.WebServer.Get("/", views.TemplateLoginRequired(s, views.Dashboard(s)))

	s.WebServer.Get("/compose", views.TemplateLoginRequired(s, s.DisplayTemplate("compose.html")))
	s.WebServer.Post("/compose", views.TemplateLoginRequired(s, views.CreateMessage(s)))

	s.WebServer.Get("/login", s.DisplayTemplate("login.html"))
	s.WebServer.Post("/login", views.LoginView(s))

	s.WebServer.Get("/logout", views.LogoutView(s))
}

// START APPLICAITON-SPECIFIC CODE

func connectToDatabase() (*gorp.DbMap) {
	// serverKey, _ := common.CreateKey()
	db, err := library.OpenDatabaseFromURL(os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Println("Unable to Connect to DB")
		fmt.Println(err)
		return nil
	}

	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.PostgresDialect{}}
	dbmap.TraceOn("[gorp]", log.New(os.Stdout, "editor:", log.Lmicroseconds)) 

	dbmap.AddTableWithName(models.Tracker{}, "dispatch_trackers").SetKeys(true, "Id")
	dbmap.AddTableWithName(models.User{}, "dispatch_users").SetKeys(true, "Id")

	return dbmap
}