### Dispatcher - a fully featured Airdispatch Client

Dispatcher was created as a webservice on top of web.go, the gorilla toolkit, and the Airdispatch Common Frameworks.

###### Dependencies

Currently Dispatcher requires that you have a Postgres database setup to store mail in. Outside of that, `go get`ing the repository should install all go-imported dependencies. 

Dispatcher does not come with a tracking server, and assumes that you have access to one. If you do not want to download the main Airdispatch project, a sample tracking server is available at `mailserver.airdispat.ch:1024`.

##### Installing Dispatcher

Thanks to Go's package management system, you can install Dispatcher in two easy steps:

1. `go get github.com/airdispatch/dispatcher`
2. `go get github.com/airdispatch/dispatcher/mailserver`

These steps should create two executable files in your `$GOROOT/bin`: `dispatcher` and `mailserver`. Both of these programs work in tandem to provide the full functionality of Dispatcher.

##### First Run

Before running anything, Dispatcher uses several different environmental variables.

  - `DATABASE_URL` - a string with the format `postgres://user@server/db_name` to create the connection to the database.
  - `COOKIE_AUTH` - a string that represents the secret used to sign the cookies stored.
  - `COOKIE_ENCRYPTION` - a string that represents the secret used to encrypt the cookies stored.

Upon succesful installation of Dispatcher, you must run `dispatcher -db_create`  to initialize the database (BEFORE running the `mailserver` program). This will walk you through creating the first user and setting up all tracking servers. Optionally, you may specify the `-port` flag to run the program on a port other than 2048.

Next, you may run the `mailserver` command with the same environmental varaible and `-port` flag. This program will serve request for mail from the Airdispatch system.

##### Other Information

All information is stored in the Postgres database in plaintext - so you may restart the servers without warning. Additionally, everything is currently stored in plaintext and is not secure. We will be fixing this in an upcoming update.
