package models

type User struct {
	Salt string
	Username string
	Password string
	Id int64
}

func CreateUser(username string, password string) *User {
	newUser := &User {
		Username: username,
		Password: password,
	}
	return newUser
}

func (user *User) VerifyPassword(password string) bool {
	return (user.Password == password)
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