package models

import (
	"airdispat.ch/client/framework"
	"airdispat.ch/common"
	"github.com/airdispatch/dispatcher/library"
	"github.com/coopernurse/gorp"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"bytes"
	"fmt"
	// "log"
	"os"
)

func ConnectToDB() (*gorp.DbMap, error) {
	// serverKey, _ := common.CreateKey()
	db, err := library.OpenDatabaseFromURL(os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, err
	}

	dbmap := &gorp.DbMap{Db: db, Dialect: gorp.PostgresDialect{}}
	// dbmap.TraceOn("[gorp]", log.New(os.Stdout, "editor:", log.Lmicroseconds)) 

	dbmap.AddTableWithName(Message{}, "dispatch_messages").SetKeys(true, "Id")
	dbmap.AddTableWithName(Alert{}, "dispatch_alerts").SetKeys(true, "Id")
	dbmap.AddTableWithName(Subscription{}, "dispatch_subscriptions").SetKeys(true, "Id")

	dbmap.AddTableWithName(Tracker{}, "dispatch_trackers").SetKeys(true, "Id")

	dbmap.AddTableWithName(User{}, "dispatch_users").SetKeys(true, "Id")

	return dbmap, nil
}

type User struct { // dispatch_userg
	Salt string
	Username string
	Password string
	FullName string
	Keypair []byte
	Id int64
	Address string
	LoadedKey *ecdsa.PrivateKey `db:"-"` // This field is transient
}

func (u *User) Populate() error {
	keys, err := common.GobDecodeKey(bytes.NewBuffer(u.Keypair))
	if err != nil {
		return err
	}

	u.LoadedKey = keys
	u.Address = common.StringAddress(&keys.PublicKey)

	return nil
}

func CreateUser(username string, password string, s *library.Server) *User {
	key, _ := common.CreateKey()
	buf := new(bytes.Buffer)
	common.GobEncodeKey(key, buf)

	newUser := &User {
		Username: username,
		Password: HashPassword(password),
		Keypair: buf.Bytes(),
	}
	newUser.Populate()

	newUser.RegisterUserWithTracker(s)

	return newUser
}

func (u  *User) RegisterUserWithTracker(s *library.Server) error {
	theTrackers, err := GetTrackerList(s.DbMap)

	c := &framework.Client {}
	c.Populate(u.LoadedKey)

	// Convert to tracker list
	success := false

	for _, v := range(theTrackers) {
		err = c.SendRegistration(v.URL, s.Mailserver)
		if err == nil {
			success = true
		} else {
			fmt.Println("Got a Tracker Error", err)
		}
	}

	if !success {
		return err
	}

	return nil
}

func GetTrackerList(dbMap *gorp.DbMap) ([]*Tracker, error) {
	var theTrackers []*Tracker
	_, err := dbMap.Select(&theTrackers, "select * from dispatch_trackers")
	if err != nil {
		fmt.Println("SQL Error", err)
		return nil, err
	}
	return theTrackers, nil
}

func GetUserWithAddress(dbMap *gorp.DbMap, address string) (*User, error) {
	var theUsers []*User
	_, err := dbMap.Select(&theUsers, "select * from dispatch_users where address='" + address + "'")
	if err != nil {
		fmt.Println("SQL Error")
		fmt.Println(err)
		return nil, err
	}

	if len(theUsers) != 1 {
		return nil, errors.New("Incorrect Number of Rows Returned")
	}
	return theUsers[0], nil
}

func (user *User) VerifyPassword(password string) bool {
	return (user.Password == HashPassword(password))
}

func HashPassword(password string) string {
	return hex.EncodeToString(common.HashSHA(nil, []byte(password)))
}

type Mailbox struct {}

type Contact struct {
	Id int64
	User int64
	Address string
	Name string
}

type Message struct {
	Id int64
	ToAddress string
	Slug string
	MessageType string
	Timestamp int64
	SendingUser int64
	Content []byte
}

type Alert struct {
	Id int64
	Content []byte
	ToAddress string
	Timestamp int64
	Folder string
	ToUser int64
}

type Stream struct {
	Id int64
	Message string
	LinkedUser string
	LinkedObject string
}

type Subscription struct {
	Id int64
	User int64
	SubscribedAddress string
	Note string
}

type Attatchment struct {}

type Tracker struct { // dispatch_tracker
	Id int64
	URL string
	Address string
}
