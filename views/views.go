package views

import (
	"github.com/hoisie/web"
	"dispatcher/models"
	"dispatcher/library"
	"fmt"
)

type ViewHandler func(ctx *web.Context)

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
					ctx.Redirect(303, "/")
					return
				}
			}
		}
		s.WriteTemplateToContext("login.html", ctx, map[string]bool{"FormError": true})
	}
}