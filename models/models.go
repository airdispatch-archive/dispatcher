package models

import (
	"airdispat.ch/common"
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
)

type User struct {
	Salt string
	Username string
	Password string
	Keypair []byte
	Id int64
	Address string `db:"-"`
	LoadedKey *ecdsa.PrivateKey `db:"-"`
}

func CreateUser(username string, password string) *User {
	key, _ := common.CreateKey()
	buf := new(bytes.Buffer)
	common.GobEncodeKey(key, buf)

	newUser := &User {
		Username: username,
		Password: HashPassword(password),
		Keypair: buf.Bytes(),
	}
	return newUser
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
	Content []byte
}

type Stream struct {}

type Attatchment struct {}

type Tracker struct {
	Id int64
	URL string
	Address string
}