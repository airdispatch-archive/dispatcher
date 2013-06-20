package main

import (
	// "github.com/gorilla/sessions"
	_ "github.com/lib/pq"
	"os"
	"flag"
	"fmt"
	"log"
	"github.com/coopernurse/gorp"
	"editor/library"
	"editor/models"
	// "airdispat.ch/common"
)

var flag_port = flag.String("port", "2048", "specify the port that the server should run on")
var db_flush = flag.Bool("db_flush", false, "specifies to flush the database")

func main() {
	flag.Parse()

	temp_port := os.Getenv("PORT")
	if temp_port == "" {
		temp_port = *flag_port
	}

	theServer := &library.Server {
		Port: temp_port,
	}

	connectToDatabase()

	theServer.RunServer()
}

// START APPLICAITON-SPECIFIC CODE

func connectToDatabase() {
	// serverKey, _ := common.CreateKey()
	db, err := library.OpenDatabaseFromURL(os.Getenv("DATABASE_URL"))
	if err != nil {
		fmt.Println("Unable to Connect to DB")
		fmt.Println(err)
		return
	}

	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.PostgresDialect{}}
	dbmap.TraceOn("[gorp]", log.New(os.Stdout, "editor:", log.Lmicroseconds)) 

	dbmap.AddTable(models.Tracker{}).SetKeys(true, "Id")
	dbmap.AddTable(models.User{}).SetKeys(true, "Id")

	if *db_flush {
		dbmap.DropTables()
	}

	err = dbmap.CreateTablesIfNotExists()
	if err != nil {
		fmt.Println("Problem Creating Tables")
		fmt.Println(err)
	}
}