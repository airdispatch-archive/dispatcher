package views

import (
	"dispatcher/models"
	"dispatcher/library"
	"github.com/hoisie/web"
	"fmt"
)

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

func RegisterWithTrackers(s *library.Server) library.TemplateView {
	return func(ctx *web.Context) {
		defer ctx.Redirect(303, "/")

		u := GetLoggedInUser(s, ctx)
		u.RegisterUserWithTracker(s)
	}
}