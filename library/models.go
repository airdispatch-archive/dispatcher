package library

import ()

type User struct {
	Username string
	Password string
	Id int64
}

type Mailbox struct {}

type Message struct {}

type Stream struct {}

type Attatchment struct {}

type Tracker struct {
	Id int64
	URL string
	Address string
}