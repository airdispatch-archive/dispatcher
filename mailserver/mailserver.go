package main

import (
	"code.google.com/p/goprotobuf/proto"
	"github.com/coopernurse/gorp"
	// "github.com/airdispatch/dispatcher/library"
	"github.com/airdispatch/dispatcher/models"
	"airdispat.ch/server/framework"
	"flag"
	"time"
	"airdispat.ch/common"
	"airdispat.ch/airdispatch"
	"crypto/ecdsa"
	"encoding/hex"
	"os"
	"fmt"
	"strings"
	"strconv"
	"bytes"
)

// Configuration Varables
var port = flag.String("port", "2048", "select the port on which to run the mail server")
var me = flag.String("me", getHostname(), "the location of the server that it should broadcast to the world")
var key_file = flag.String("key", "", "the file to store keys")

var dbMap *gorp.DbMap

var noEncryption string = "none"

func getHostname() string {
	s, _ := os.Hostname()
	return s
}

// Variables that store information about the server
var serverLocation string
var serverKey *ecdsa.PrivateKey

func main() {
	// Parse the configuration Command Line Falgs
	flag.Parse()

	// Create a Signing Key for the Server
	loadedKey, err := common.LoadKeyFromFile(*key_file)

	if err != nil {

		loadedKey, err = common.CreateKey()
		if err != nil {
			fmt.Println("Unable to Create Tracker Key")
			return
		}

		if *key_file != "" {

			err = common.SaveKeyToFile(*key_file, loadedKey)
			if err != nil {
				fmt.Println("Unable to Save Tracker Key")
				return
			}
		}

	}
	fmt.Println("Loaded Address", common.StringAddress(&loadedKey.PublicKey))

	dbMap, err = models.ConnectToDB()
	if err != nil {
		fmt.Println("Couldn't connect to DB")
		fmt.Println(err)
		return
	}

	theTrackers, err := models.GetTrackerList(dbMap)
	if err != nil {
		fmt.Println("Couldn't get Trackers from DB")
		fmt.Println(err)
		return
	}

	savedTrackers := make([]string, len(theTrackers))
	for i, v := range(theTrackers) {
		savedTrackers[i] = v.URL
	}

	// Find the location of this server
	serverLocation = *me
	handler := &myServer{}
	theServer := framework.Server{
		LocationName: *me,
		Key: loadedKey,
		TrackerList: savedTrackers,
		Delegate: handler,
	}
	serverErr := theServer.StartServer(*port)
	if serverErr != nil {
		fmt.Println("Unable to Start Server")
		fmt.Println(err)
	}

}

type myServer struct{
	framework.BasicServer
}

// Function that Handles an Alert of a Message
func (myServer) SaveIncomingAlert(alert *airdispatch.Alert, alertData []byte, fromAddr string) {
	// Get the recipient address of the message
	toAddr := *alert.ToAddress
	theUser, _ := models.GetUserWithAddress(dbMap, toAddr)

	theSavedAlert := &models.Alert {
		Content: alertData,
		ToAddress: fromAddr,
		Timestamp: time.Now().Unix(),
		ToUser: theUser.Id,
	}

	dbMap.Insert(theSavedAlert)
}

func (myServer) SavePublicMail(theMail []byte, fromAddr string) {
	// theUser, _ := models.GetUserWithAddress(dbMap, fromAddr)

	// theSavedMail := &models.Message {
	// 	ToAddress: "",
	// 	Slug: GetMessageId(theMail),
	// 	MessageType: "_unknown",
	// }

	// dbMap.Insert(theSavedMail)
}

func (myServer) SavePrivateMail(theMail []byte, toAddress []string) (id string) {
	// // Get a hash of the Message
	// hash := GetMessageId(theMail)

	// // Create a Record to Store the Message in the Outgoing Mail Box
	// storedData := Mail {
	// 	approved: toAddress,
	// 	data: theMail,
	// 	receivedTime: time.Now(),
	// }

	// // Store the Message in the Database
	// storedMessages[hash] = storedData

	// return hash
	return ""
}

func GetMessageId(theMail []byte) string {
	return hex.EncodeToString(common.HashSHA(theMail, nil))
}

func (myServer) RetrieveMessage(id string) ([]byte, []string) {
	type queryResult struct {
		Content []byte
		ToAddress string
		Keypair []byte
		Address string
	}

	query := "select m.content, m.toaddress, u.keypair, u.address " 
	query += "from dispatch_messages m, dispatch_users u "
	query += "where m.slug = '" + id + "' and m.sendinguser = u.id"
	query += "limit 1 "

	var results []*queryResult
	dbMap.Select(&results, query)

	if len(results) != 1 {
		fmt.Println("Incorrect Number of Messages Returned")
		return nil, nil
	}

	keys, err := common.GobDecodeKey(bytes.NewBuffer(results[0].Keypair))
	if err != nil {
		fmt.Println("Error Getting Keys")
		return nil, nil
	}

	newMail := &airdispatch.Mail {
		FromAddress: &results[0].Address,
		Data: results[0].Content,
		Encryption: &noEncryption,
	}
	data, _ := proto.Marshal(newMail)

	toSend, _ := common.CreateAirdispatchMessage(data, keys, common.MAIL_MESSAGE)

	return toSend, strings.Split(results[0].ToAddress, ",")
}

func (m myServer) RetrieveInbox(addr string, since uint64) [][]byte {
	type queryResult struct {
		Content []byte
	}

	query := "select m.content " 
	query += "from dispatch_alerts m, dispatch_users u "
	query += "where m.touser = u.id and toaddress='' and timestamp>" + strconv.FormatUint(since, 10) + " "
	query += "and u.address='" + addr + "' "
	query += "order by m.timestamp desc "

	var results []*queryResult
	dbMap.Select(&results, query)

	output := make([][]byte, len(results))

	for i, v := range(results) {
		output[i] = v.Content
	}

	return output
}

func (m myServer) RetrievePublic(fromAddr string, since uint64) [][]byte {
	type queryResult struct {
		Content []byte
		Keypair []byte
	}

	query := "select m.content, u.keypair " 
	query += "from dispatch_messages m, dispatch_users u "
	query += "where m.sendinguser = u.id and toaddress='' and timestamp > " + strconv.FormatUint(since, 10) + " "
	query += "and u.address = '" + fromAddr + "' "
	query += "order by m.timestamp desc"

	var results []*queryResult
	dbMap.Select(&results, query)

	output := make([][]byte, len(results))

	var keys *ecdsa.PrivateKey = nil

	for i, v := range(results) {
		if keys == nil {
			var err error
			keys, err = common.GobDecodeKey(bytes.NewBuffer(v.Keypair))
			if err != nil {
				fmt.Println("Error Getting Keys")
				return nil
			}
		}

		newMail := &airdispatch.Mail {
			FromAddress: &fromAddr,
			Data: v.Content,
			Encryption: &noEncryption,
		}
		data, _ := proto.Marshal(newMail)

		toSend, _ := common.CreateAirdispatchMessage(data, keys, common.MAIL_MESSAGE)

		// Remove the Prefix
		output[i] = toSend[6:]
	}

	return output
}
