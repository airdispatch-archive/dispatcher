package views

import (
	"github.com/hoisie/web"
	"github.com/airdispatch/dispatcher/models"
	"github.com/airdispatch/dispatcher/library"
	"airdispat.ch/common"
	"airdispat.ch/airdispatch"
	cf "airdispat.ch/client/framework"
	sf "airdispat.ch/server/framework"
	"code.google.com/p/goprotobuf/proto"
	"encoding/hex"
	"time"
	"fmt"
	"regexp"
	"strconv"
)

type ViewHandler func(ctx *web.Context)
var no_encryption string = "none"
var blog_title string = "blog/title"
var blog_content string = "blog/content"
var blog_author string = "blog/author"
var blog_date string = "blog/date"

var contentRegex *regexp.Regexp = nil

func CreateMessage(s *library.Server) library.TemplateView {
	if contentRegex == nil {
		contentRegex = regexp.MustCompile(`^content\[([0-9]+)\]\[([0-9]+)\]`)
	}
	return func(ctx *web.Context) {
		defer ctx.Redirect(303, "/")

		to_address := ctx.Params["to_address"]
		sending_user := GetLoggedInUser(s, ctx)

		byteData, err := ContextToDataTypeBytes(ctx)
		if err != nil {
			fmt.Println("Cannot convert context data.")
			fmt.Println(err)
			return
		}

		newMessage := &models.Message{}
		newMessage.Content = byteData
		newMessage.Timestamp = time.Now().Unix()
		newMessage.SendingUser = sending_user.Id
		newMessage.ToAddress = to_address
		newMessage.Slug = hex.EncodeToString(common.HashSHA(nil, byteData))

		// If there _is_ a ToAddress, you should actually
		// send the message
		if to_address != "" {
			// Create Server Instance
			newServer := sf.Server{
				LocationName: s.Mailserver,
				Key: sending_user.LoadedKey,
				Delegate: sf.BasicServer{},
			}

			// Get the Tracker List
			trackerList, _ := models.GetTrackerList(s.DbMap)
			stringTrackers := make([]string, len(trackerList))
			for i, v := range(trackerList) {
				stringTrackers[i] = v.URL
			}

			// Get the Location of the Server
			serverLocation, err := common.LookupLocation(to_address, stringTrackers, sending_user.LoadedKey)
			fmt.Println("Found Location", serverLocation, "with error", err)

			// Send the Alert
			newServer.SendAlert(serverLocation, newMessage.Slug, to_address)
		}

		s.DbMap.Insert(newMessage)
	}
}

func UpdateMessage(s *library.Server) library.WildcardTemplateView {
	return func(ctx *web.Context, val string) {
		defer ctx.Redirect(303, "/")

		theMessage, _ := s.DbMap.Get(models.Message{}, val)
		byteData, err := ContextToDataTypeBytes(ctx)
		if err != nil {
			fmt.Println("Cannot convert context data.")
			fmt.Println(err)
			return
		}

		theMessage.(*models.Message).Content = byteData
		s.DbMap.Update(theMessage)
	}
}

func DisplayEditMessage(s *library.Server) library.WildcardTemplateView {
	return func(ctx *web.Context, val string) {
		theMessage, _ := s.DbMap.Get(models.Message{}, val)

		context := make(map[string]interface{})
		context["Initial"] = MessageToContext(theMessage.(*models.Message), s)

		s.WriteTemplateToContext("compose.html", ctx, context)
	}
}

func ShowSubscriptions(s *library.Server) library.TemplateView {
	return func(ctx *web.Context) {
		theUser := GetLoggedInUser(s, ctx)

		var theMessages []*models.Subscription
		s.DbMap.Select(&theMessages, "select * from dispatch_subscriptions where \"user\"=" + strconv.FormatInt(theUser.Id, 10))

		context := make(map[string]interface{})
		context["Subscriptions"] = theMessages

		context["DisplayTag"] = DisplayAirDispatchAddress(s)

		s.WriteTemplateToContext("subscribe.html", ctx, context)
	}
}

func CreateSubscription(s *library.Server) library.TemplateView {
	return func(ctx *web.Context) {
		to_address := ctx.Params["to_address"]
		sending_user := GetLoggedInUser(s, ctx)
		theSubscription := &models.Subscription {
			SubscribedAddress: to_address,
			User: sending_user.Id,
			Note: "",
		}
		s.DbMap.Insert(theSubscription)
		ctx.Redirect(303, "/")
	}
}

func ShowFolder(s *library.Server, folderName string) library.TemplateView {
	return func(ctx *web.Context) {
		context := make(map[string]interface{})
		context["FolderName"] = folderName

		context["TimeFunction"] = TimestampToString()
		context["BasePrefix"] = "alert"

		current_user := GetLoggedInUser(s, ctx)

		if folderName == "Sent Messages" {
			var theMessages []*models.Message
			s.DbMap.Select(&theMessages, "select * from dispatch_messages where sendinguser=" + strconv.FormatInt(current_user.Id, 10) + " order by timestamp DESC")
			context["Messages"] = theMessages
			context["BasePrefix"] = "message"
		} else if folderName == "Inbox" {
			var theMessages []*models.Alert
			s.DbMap.Select(&theMessages, "select * from dispatch_alerts where touser=" + strconv.FormatInt(current_user.Id, 10) + "order by timestamp DESC")
			context["Messages"] = theMessages
		}

		s.WriteTemplateToContext("messages/list.html", ctx, context)
	}
}

func ShowAlert(s *library.Server) library.WildcardTemplateView {
	return func(ctx *web.Context, val string) {
		theAlert, _ := s.DbMap.Get(models.Alert{}, val)

		castedAlert := theAlert.(*models.Alert)

		data, _, fromAddr, err := common.ReadSignedBytes(castedAlert.Content)

		unMarshalledAlert :=  &airdispatch.Alert{}
		err = proto.Unmarshal(data, unMarshalledAlert)
		if err != nil {
			ctx.WriteString("Malformed Alert")
			fmt.Println(err)
			return
		}

		// Download Message
		current_user := GetLoggedInUser(s, ctx)

		newClient := cf.Client{}
		newClient.Populate(current_user.LoadedKey)

		theMail, err := newClient.DownloadSpecificMessageFromServer(unMarshalledAlert.GetMessageId(), unMarshalledAlert.GetLocation())
		if err != nil {
			ctx.WriteString("Couldn't Download Message")
			fmt.Println(err)
			return
		}

		displayMessage(s, MailToMessage(theMail, fromAddr), ctx)
	}
}

func ShowMessage(s *library.Server) library.WildcardTemplateView {
	return func(ctx *web.Context, val string) {
		theMessage, _ := s.DbMap.Get(models.Message{}, val)

		displayMessage(s, theMessage.(*models.Message), ctx)
	}
}

func displayMessage(s *library.Server, m *models.Message, ctx *web.Context) {
	context := make(map[string]interface{})
	context["Message"] = MessageToContext(m, s)

	s.WriteTemplateToContext("messages/show.html", ctx, context)
}

func Dashboard(s *library.Server) library.TemplateView {
	return func(ctx *web.Context) {
		context := make(map[string]interface{})

		theUser := GetLoggedInUser(s, ctx)
		context["User"] = theUser

		var theMessages []*models.Subscription
		s.DbMap.Select(&theMessages, "select * from dispatch_subscriptions where \"user\"=" + strconv.FormatInt(theUser.Id, 10))

		theClient := &cf.Client {}
		theClient.Populate(theUser.LoadedKey)

		trackerList, _ := models.GetTrackerList(s.DbMap)
		stringTrackers := make([]string, len(trackerList))
		for i, v := range(trackerList) {
			stringTrackers[i] = v.URL
		}

		pastMonth := time.Now().Add(time.Duration(-7) * time.Hour * 24)

		outputMail := make([]*airdispatch.Mail, 0)

		for _, value := range(theMessages) {
			downloadedMail, err := theClient.DownloadPublicMail(stringTrackers, value.SubscribedAddress, uint64(pastMonth.Unix()))
			if err != nil {
				continue
			}

			outputMail = append(outputMail, downloadedMail...)
		}

		context["Messages"] = outputMail
		context["DisplayTag"] = DisplayMessageTag(s)

		s.WriteTemplateToContext("dashboard.html", ctx, context)
	}
}
