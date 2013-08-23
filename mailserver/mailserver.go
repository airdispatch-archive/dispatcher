package main

import (
	"code.google.com/p/goprotobuf/proto"
	"github.com/coopernurse/gorp"
	// "github.com/airdispatch/dispatcher/library"
	"airdispat.ch/airdispatch"
	"airdispat.ch/common"
	"airdispat.ch/server/framework"
	"bytes"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/airdispatch/dispatcher/models"
	"os"
	"strconv"
	"strings"
	"time"
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

func main() {
	// Parse the configuration Command Line Falgs
	flag.Parse()

	// Create a Signing Key for the Server
	loadedKey, err := common.LoadKeyFromFile(*key_file)

	if err != nil {

		loadedKey, err = common.CreateADKey()
		if err != nil {
			fmt.Println("Unable to Create Tracker Key")
			return
		}

		if *key_file != "" {

			err = loadedKey.SaveKeyToFile(*key_file)
			if err != nil {
				fmt.Println("Unable to Save Tracker Key")
				return
			}
		}

	}
	fmt.Println("Loaded Address", loadedKey.HexEncode())

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
	for i, v := range theTrackers {
		savedTrackers[i] = v.URL
	}

	// Find the location of this server
	handler := &myServer{}
	theServer := framework.Server{
		LocationName: *me,
		Key:          loadedKey,
		TrackerList:  savedTrackers,
		Delegate:     handler,
	}
	serverErr := theServer.StartServer(*port)
	if serverErr != nil {
		fmt.Println("Unable to Start Server")
		fmt.Println(err)
	}

}

type myServer struct {
	framework.BasicServer
}

func (myServer) AllowSendConnection(user string) bool {
	return false
}

// Function that Handles an Alert of a Message
func (myServer) SaveIncomingAlert(alert *airdispatch.Alert, alertData []byte, fromAddr string) {
	// Get the recipient address of the message
	toAddr := *alert.ToAddress
	theUser, err := models.GetUserWithAddress(dbMap, toAddr)
	if err != nil {
		fmt.Println("Getting User Error", err)
	}

	theSavedAlert := &models.Alert{
		Content:   alertData,
		ToAddress: fromAddr,
		Timestamp: time.Now().Unix(),
		ToUser:    theUser.Id,
	}

	dbMap.Insert(theSavedAlert)
}

func (myServer) SavePublicMail(theMail []byte, fromAddr string)                 {}
func (myServer) SavePrivateMail(theMail []byte, toAddress []string) (id string) { return "" }

func GetMessageId(theMail []byte) string {
	return hex.EncodeToString(common.HashSHA(theMail))
}

func (myServer) RetrieveMessageForUser(id string, addr string) []byte {
	var newMail *airdispatch.Mail
	var keys *common.ADKey

	// Respond to the Profile Message Seperately
	if strings.HasSuffix(id, "profile") {
		idComponents := strings.Split(id, "::")

		query := "select * from dispatch_users where address='" + idComponents[0] + "'"

		var results []*models.User
		dbMap.Select(&results, query)

		if len(results) != 1 {
			fmt.Println("No use for that address.")
			return nil
		}

		scopedKey, err := common.GobDecodeKey(bytes.NewBuffer(results[0].Keypair))
		if err != nil {
			fmt.Println("Error Getting Keys")
			return nil
		}
		keys = scopedKey

		currentTime := uint64(time.Now().Unix())
		blank := ""
		ad_profile_name := "airdispat.ch/profiles/name"
		ad_profile_photo := "airdispat.ch/profiles/photo"

		nameData := &airdispatch.MailData_DataType{
			TypeName: &ad_profile_name,
			Payload:  []byte(results[0].FullName),
		}

		photoData := &airdispatch.MailData_DataType{
			TypeName: &ad_profile_photo,
			Payload:  []byte("https://fbcdn-sphotos-f-a.akamaihd.net/hphotos-ak-ash4/302066_10151616869615972_963571219_n.jpg"),
		}

		allData := &airdispatch.MailData{
			Payload: []*airdispatch.MailData_DataType{nameData, photoData},
		}

		data, err := proto.Marshal(allData)
		if err != nil {
			fmt.Println("Unable to Marshal the Data")
			return nil
		}

		newMail = &airdispatch.Mail{
			FromAddress: &results[0].Address,
			Data:        data,
			Encryption:  &noEncryption,
			Timestamp:   &currentTime,
			ToAddress:   &blank,
		}
	} else {
		type queryResult struct {
			Content   []byte
			ToAddress string
			Keypair   []byte
			Address   string
			Timestamp int64
		}

		query := "select m.content, m.toaddress, u.keypair, u.address, m.timestamp "
		query += "from dispatch_messages m, dispatch_users u "
		query += "where m.slug = '" + id + "' and m.sendinguser = u.id and m.toaddress = '" + addr + "' "
		query += "limit 1 "

		var results []*queryResult
		dbMap.Select(&results, query)

		if len(results) != 1 {
			fmt.Println("Incorrect Number of Messages Returned")
			return nil
		}

		scopedKey, err := common.GobDecodeKey(bytes.NewBuffer(results[0].Keypair))
		if err != nil {
			fmt.Println("Error Getting Keys")
			return nil
		}
		keys = scopedKey

		currentTime := uint64(results[0].Timestamp)

		newMail = &airdispatch.Mail{
			FromAddress: &results[0].Address,
			Data:        results[0].Content,
			Encryption:  &noEncryption,
			Timestamp:   &currentTime,
			ToAddress:   &results[0].ToAddress,
		}
	}

	data, err := proto.Marshal(newMail)
	if err != nil {
		fmt.Println("Erorr marshalling", err)
	}

	newMessage := &common.ADMessage{data, common.MAIL_MESSAGE, ""}

	toSend, err := keys.CreateADMessage(newMessage)
	if err != nil {
		fmt.Println("Error making message", err)
	}

	return toSend[6:]
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

	for i, v := range results {
		output[i] = v.Content
	}

	return output
}

func (m myServer) RetrievePublic(fromAddr string, since uint64) [][]byte {
	type queryResult struct {
		Content   []byte
		Keypair   []byte
		Timestamp int64
	}

	query := "select m.content, u.keypair, m.timestamp "
	query += "from dispatch_messages m, dispatch_users u "
	query += "where m.sendinguser = u.id and toaddress='' and timestamp > " + strconv.FormatUint(since, 10) + " "
	query += "and u.address = '" + fromAddr + "' "
	query += "order by m.timestamp desc"

	var results []*queryResult
	dbMap.Select(&results, query)

	output := make([][]byte, len(results))

	var keys *common.ADKey = nil
	toAll := ""

	for i, v := range results {
		if keys == nil {
			var err error
			keys, err = common.GobDecodeKey(bytes.NewBuffer(v.Keypair))
			if err != nil {
				fmt.Println("Error Getting Keys")
				return nil
			}
		}

		currentTime := uint64(v.Timestamp)

		newMail := &airdispatch.Mail{
			FromAddress: &fromAddr,
			Data:        v.Content,
			Encryption:  &noEncryption,
			Timestamp:   &currentTime,
			ToAddress:   &toAll,
		}
		data, _ := proto.Marshal(newMail)

		newMessage := &common.ADMessage{data, common.MAIL_MESSAGE, ""}

		toSend, err := keys.CreateADMessage(newMessage)
		if err != nil {
			fmt.Println("Error Creating Message")
			fmt.Println(err)
			continue
		}

		// Remove the Prefix
		output[i] = toSend[6:]
	}

	return output
}
