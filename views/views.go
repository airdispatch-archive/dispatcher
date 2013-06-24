package views

import (
	"github.com/hoisie/web"
	"dispatcher/models"
	"dispatcher/library"
	"airdispat.ch/common"
	"bytes"
)

type ViewHandler func(ctx *web.Context)

func CreateMessage(s *library.Server) library.TemplateView {
	return func(ctx *web.Context) {
		ctx.Redirect(303, "/")
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
	
	keys, _ := common.GobDecodeKey(bytes.NewBuffer(newUser.Keypair))
	newUser.LoadedKey = keys
	newUser.Address = common.StringAddress(&keys.PublicKey)

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