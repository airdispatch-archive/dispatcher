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
)

type ViewHandler func(ctx *web.Context)
var no_encryption string = "none"
var blog_title string = "blog/title"
var blog_content string = "blog/content"
var blog_author string = "blog/author"
var blog_date string = "blog/date"

func CreateMessage(s *library.Server) library.TemplateView {
	return func(ctx *web.Context) {
		to_address := ctx.Params["to_address"]
		sending_user := GetLoggedInUser(s, ctx)
		fmt.Println("New Message To", to_address)
		switch ctx.Params["mes_type"] {
			case "_blog":
				fmt.Println("Blog Post")
				newMessage := &models.Message{}
				newMessage.ToAddress = ""
				newMessage.Slug = hex.EncodeToString(common.HashSHA(nil, []byte(ctx.Params["blog_title"])))
				newMessage.MessageType = "_blog"

				title := &airdispatch.MailData_DataType{
					TypeName: &blog_title,
					Payload: []byte(ctx.Params["blog_title"]),
					Encryption: &no_encryption,
				}

				content := &airdispatch.MailData_DataType{
					TypeName: &blog_content,
					Payload: []byte(ctx.Params["blog_content"]),
					Encryption: &no_encryption,
				}

				author := &airdispatch.MailData_DataType{
					TypeName: &blog_author,
					Payload: []byte(sending_user.FullName),
					Encryption: &no_encryption,
				}

				date := &airdispatch.MailData_DataType{
					TypeName: &blog_date,
					Payload: []byte(time.Now().String()),
					Encryption: &no_encryption,
				}

				theData := &airdispatch.MailData{
					Payload: []*airdispatch.MailData_DataType{title, content, author, date},
				}

				byteData, _ := proto.Marshal(theData)
				newMessage.Content = byteData
				newMessage.Timestamp = time.Now().Unix()

				s.DbMap.Insert(newMessage)
			default:
				fmt.Println("Unknown Post Type")
		}
		ctx.Redirect(303, "/")
	}
}

func ShowFolder(s *library.Server, folderName string) library.TemplateView {
	return func(ctx *web.Context) {
		context := make(map[string]interface{})
		context["FolderName"] = folderName

		context["TimeFunction"] = TimestampToString()

		var theMessages []*models.Message
		s.DbMap.Select(&theMessages, "select * from dispatch_messages")

		context["Messages"] = theMessages

		s.WriteTemplateToContext("show_messages.html", ctx, context)
	}
}

func ShowMessage(s *library.Server) library.WildcardTemplateView {
	return func(ctx *web.Context, val string) {
		s.WriteTemplateToContext("message.html", ctx, nil)
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