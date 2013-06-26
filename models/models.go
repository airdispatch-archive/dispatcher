package models

import (
	"airdispat.ch/common"
	"airdispat.ch/client/framework"
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"dispatcher/library"
	"fmt"
)

type User struct { // dispatch_userg
	Salt string
	Username string
	Password string
	FullName string
	Keypair []byte
	Id int64
	Address string `db:"-"`
	LoadedKey *ecdsa.PrivateKey `db:"-"`
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
	var theTrackers []*Tracker
	_, err := s.DbMap.Select(&theTrackers, "select * from dispatch_trackers")
	if err != nil {
		fmt.Println("SQL Error", err)
		return err
	}

	c := &framework.Client {}
	c.Populate(u.LoadedKey)

	// Convert to tracker list
	success := false

	for _, v := range(theTrackers) {
		err = c.SendRegistration(v.URL, s.Mailserver)
		fmt.Println("Tracker Registration", err)
		if err == nil {
			success = true
		}
	}

	if !success {
		return err
	}

	return nil
}

func (user *User) VerifyPassword(password string) bool {
	return (user.Password == HashPassword(password))
}

func HashPassword(password string) string {
	return hex.EncodeToString(common.HashSHA(nil, []byte(password)))
}

type Mailbox struct {}

type Message struct {
	Id int64
	ToAddress string
	Slug string
	MessageType string
	Timestamp int64
	Content []byte
}

type Stream struct {
	Id int64
	Message string
	LinkedUser string
	LinkedObject string
}

type Attatchment struct {}

type Tracker struct { // dispatch_tracker
	Id int64
	URL string
	Address string
}