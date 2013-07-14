package views

import (
	"github.com/airdispatch/dispatcher/models"
	library "github.com/airdispatch/go-pressure"
	"github.com/hoisie/web"
)

const LoginSessionMapKey = "user_id"

func RegisterUser(s *library.Server) library.TemplateView {
	return func(ctx *web.Context) {
		username := ctx.Params["username"]
		password := ctx.Params["password"]
		cpassword := ctx.Params["password_confirm"]
		full_name := ctx.Params["full_name"]

		if password != "" && cpassword != "" && username != "" && full_name != "" {
			if password != cpassword {
				s.WriteTemplateToContext("login.html", ctx, map[string]bool{"RegPasswordMatchError": true})
				return
			}

			n, _ := s.DbMap.Select(models.User{}, "select * from dispatch_users where username='" + username + "'")
			if len(n) != 0 {
				s.WriteTemplateToContext("login.html", ctx, map[string]bool{"RegUsernameTaken": true})
				return
			}

			theUser := models.CreateUser(username, password, s)
			theUser.FullName = full_name
			s.DbMap.Insert(theUser)

			LoginUser(s, theUser, ctx)
			
			ctx.Redirect(303, "/")
			return
		} else {
			s.WriteTemplateToContext("login.html", ctx, map[string]bool{"RegFormError": true})
			return
		}
	}
}

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
		return
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