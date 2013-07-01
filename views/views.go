package views

import (
	"code.google.com/p/goprotobuf/proto"
	"github.com/hoisie/web"
	"dispatcher/models"
	"dispatcher/library"
	"airdispat.ch/common"
	"airdispat.ch/airdispatch"
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
		to_address := ctx.Params["to_address"]
		sending_user := GetLoggedInUser(s, ctx)

		if len(ctx.Params) < 2 {
			fmt.Println("Error")
			ctx.Redirect(303, "/")
			return
		}

		content_types := make([]*airdispatch.MailData_DataType, (len(ctx.Params) - 1) / 2)

		for key, value := range(ctx.Params) {
			if key != "to_address" {
				// content[index][type]
				i := contentRegex.FindAllStringSubmatch(key, -1)[0]

				typeIndex, _ := strconv.ParseInt(i[1], 10, 0)

				theData := content_types[typeIndex]
				if theData == nil {
					theData = &airdispatch.MailData_DataType {}
				}

				if i[2] == "0" {
					theType := ctx.Params[key]
					theData.TypeName = &theType
				} else if i[2] == "1" {
					theData.Payload = []byte(value)
				}

				content_types[typeIndex] = theData
			}
		}

		theData := &airdispatch.MailData{
			Payload: content_types,
		}

		byteData, _ := proto.Marshal(theData)

		newMessage := &models.Message{}
		newMessage.Content = byteData
		newMessage.Timestamp = time.Now().Unix()
		newMessage.SendingUser = sending_user.Id
		newMessage.ToAddress = to_address
		newMessage.Slug = hex.EncodeToString(common.HashSHA(nil, byteData))

		s.DbMap.Insert(newMessage)

		ctx.Redirect(303, "/")
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
			s.DbMap.Select(&theMessages, "select * from dispatch_messages")
			context["Messages"] = theMessages
		} else if folderName == "Inbox" {
			var theMessages []*models.Alert
			s.DbMap.Select(&theMessages, "select * from dispatch_alerts")
			context["Messages"] = theMessages
		}


		s.WriteTemplateToContext("show_messages.html", ctx, context)
	}
}

func ShowMessage(s *library.Server) library.WildcardTemplateView {
	return func(ctx *web.Context, val string) {
		theMessage, _ := s.DbMap.Get(models.Message{}, val)

		context := make(map[string]interface{})
		context["Message"] = MessageToContext(theMessage.(*models.Message), s)

		s.WriteTemplateToContext("message.html", ctx, context)
	}
}

func Dashboard(s *library.Server) library.TemplateView {
	return func(ctx *web.Context) {
		s.WriteTemplateToContext("dashboard.html", ctx, GetLoggedInUser(s, ctx))
	}
}

const LoginSessionMapKey = "user_id"

func LoginView(s *library.Server) library.TemplateView {
	return func(ctx *web.Context) {
		username := ctx.Params["username"]
		password := ctx.Params["password"]

		var theUsers []*models.User
		_, err := s.DbMap.Select(&theUsers, "select * from dispatch_users where username='" + username + "'")
		fmt.Println(err, theUsers)
		if err == nil {
			if len(theUsers) > 0 && theUsers[0] != nil {
				if theUsers[0].VerifyPassword(password) {
					if !LoginUser(s, theUsers[0], ctx) {
						s.WriteTemplateToContext("login.html", ctx, map[string]bool{"DatabaseError": true})
						return
					}
					ctx.Redirect(303, "/")
					return
				}
			}
		}
		s.WriteTemplateToContext("login.html", ctx, map[string]bool{"FormError": true})
	}
}

func LoginUser(s *library.Server, u *models.User, ctx *web.Context) bool {
	session, err := s.GetMainSession(ctx)
	if err != nil {
		return false
	}
	session.Values[LoginSessionMapKey] = u.Id
	err = library.SaveSessionWithContext(session, ctx)
	if err != nil {
		return false
	}
	return true
}

func GetLoggedInUser(s *library.Server, ctx *web.Context) (*models.User) {
	session, err := s.GetMainSession(ctx)

	if session.Values[LoginSessionMapKey] == nil || session.Values[LoginSessionMapKey] == "" || session.Values[LoginSessionMapKey] == -1 {
		return nil
	}

	user, err := s.DbMap.Get(models.User{}, session.Values[LoginSessionMapKey])
	if err != nil {
		return nil
	}

	if user == nil {
		return nil
	}

	newUser := user.(*models.User)
	newUser.Populate()

	return newUser
}

func TemplateLoginRequired(s *library.Server, t library.TemplateView) library.TemplateView {
	return func(ctx *web.Context) {
		u := GetLoggedInUser(s, ctx)
		if (u != nil) {
			t(ctx)
		} else {
			ctx.Redirect(303, "/login")
		}
	}
}

func LogoutView(s *library.Server) library.TemplateView {
	return func(ctx *web.Context) {
		session, _ := s.GetMainSession(ctx)
		session.Values[LoginSessionMapKey] = -1
		library.SaveSessionWithContext(session, ctx)
		ctx.Redirect(303, "/login")
	}
}

type TemplateTag func(interface{}) interface{}

func TimestampToString() TemplateTag {
	return func(arg interface{}) interface{} {
		timestamp := arg.(int64)
		return time.Unix(timestamp, 0).Format("Jan 2, 2006 at 3:04pm")
	}
}

func DisplayAirDispatchAddress(s *library.Server) TemplateTag {
	return func(arg interface{}) interface{} {
		return arg
	}
}

func MessageToContext(m *models.Message, s *library.Server) map[string]interface{} {
	output := make(map[string]interface{})

	output["ID"] = m.Id
	output["MesType"] = m.MessageType

	if m.ToAddress == "" {
		output["TO"] = "Public"
		output["ToAddress"] = ""
		output["Encryption"] = ""
	} else {
		output["TO"] = DisplayAirDispatchAddress(s)(m.ToAddress)
		output["ToAddress"] = output["TO"]
		output["Encryption"] = "aes/256"
	}

	theUser, _ := s.DbMap.Get(&models.User{}, m.SendingUser)
	output["FROM"] = theUser.(*models.User).FullName

	output["Timestamp"] = TimestampToString()(m.Timestamp)

	theData := &airdispatch.MailData{}
	proto.Unmarshal(m.Content, theData)

	allContent := make([]map[string]interface{}, len(theData.Payload))

	for i, v := range(theData.Payload) {
		allContent[i] = map[string]interface{} {
			"TypeName": v.TypeName,
			"Payload": string(v.Payload),
		}
	}

	output["Content"] = allContent

	return output
}