package views

import (
	"github.com/hoisie/web"
	"dispatcher/models"
	"dispatcher/library"
	"fmt"
)

type ViewHandler func(ctx *web.Context)

const LoginSessionMapKey = "user_id"

func LoginView(s *library.Server) ViewHandler {

	return func(ctx *web.Context) {
		username := ctx.Params["username"]
		password := ctx.Params["password"]

		var theUsers []*models.User
		_, err := s.DbMap.Select(&theUsers, "select * from dispatch_users where username='" + username + "'")
		fmt.Println(theUsers, err)
		if err == nil {
			if len(theUsers) > 0 && theUsers[0] != nil {
				if theUsers[0].VerifyPassword(password) {
					LoginUser(s, theUsers[0], ctx)
					ctx.Redirect(303, "/")
					return
				}
			}
		}
		s.WriteTemplateToContext("login.html", ctx, map[string]bool{"FormError": true})
	}
}

func LoginUser(s *library.Server, u *models.User, ctx *web.Context) {
	session, err := s.GetMainSession(ctx)
	fmt.Println(err)
	session.Values[LoginSessionMapKey] = u.Id
	library.SaveSessionWithContext(session, ctx)
}

func GetLoggedInUser(s *library.Server, ctx *web.Context) (*models.User) {
	session, err := s.GetMainSession(ctx)
	fmt.Println(session, session.Values, err)

	if session.Values[LoginSessionMapKey] == nil || session.Values[LoginSessionMapKey] == "" {
		return nil
	}

	user, err := s.DbMap.Get(models.User{}, session.Values[LoginSessionMapKey])
	if err != nil {
		return nil
	}

	return user.(*models.User)
}

func TemplateLoginRequired(s *library.Server, t library.TemplateView) ViewHandler {
	return func(ctx *web.Context) {
		u := GetLoggedInUser(s, ctx)
		if (u != nil) {
			t(ctx)
		} else {
			ctx.Redirect(303, "/login")
		}
	}
}