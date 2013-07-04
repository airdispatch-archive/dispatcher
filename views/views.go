package views

import (
	"github.com/hoisie/web"
	"dispatcher/models"
	"dispatcher/library"
	"airdispat.ch/common"
	"airdispat.ch/airdispatch"
	cf "airdispat.ch/client/framework"
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


		if folderName == "Sent Messages" {
			var theMessages []*models.Message
			s.DbMap.Select(&theMessages, "select * from dispatch_messages order by timestamp DESC")
			context["Messages"] = theMessages
		} else if folderName == "Inbox" {
			var theMessages []*models.Alert
			s.DbMap.Select(&theMessages, "select * from dispatch_alerts order by timestamp DESC")
			context["Messages"] = theMessages
		}


		s.WriteTemplateToContext("messages/list.html", ctx, context)
	}
}

func ShowMessage(s *library.Server) library.WildcardTemplateView {
	return func(ctx *web.Context, val string) {
		theMessage, _ := s.DbMap.Get(models.Message{}, val)

		context := make(map[string]interface{})
		context["Message"] = MessageToContext(theMessage.(*models.Message), s)

		s.WriteTemplateToContext("messages/show.html", ctx, context)
	}
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