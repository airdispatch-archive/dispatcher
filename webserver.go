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
	}
	theServer.ConfigServer()
	defineRoutes(theServer)

	theServer.RunServer()
}

func defineRoutes(s *library.Server) {
	s.WebServer.Get("/", views.TemplateLoginRequired(s, s.DisplayTemplate("dashboard.html")))
	s.WebServer.Get("/login", s.DisplayTemplate("login.html"))
	s.WebServer.Post("/login", views.LoginView(s))
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

	if *db_flush {
		dbmap.DropTables()
	}

	if *db_create {
		err = dbmap.CreateTablesIfNotExists()
		if err != nil {
			fmt.Println("Problem Creating Tables")
			fmt.Println(err)
			return nil
		}

		fmt.Println("Let's create the first user.")

		var username, password string

		fmt.Print("Username (no spaces): ")
		fmt.Scanln(&username)

		fmt.Print("Password: ")
		fmt.Scanln(&password)

		newUser := models.CreateUser(username, password)
		dbmap.Insert(newUser)
	}

	return dbmap
}