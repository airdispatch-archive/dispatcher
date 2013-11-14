package main

import (
	"flag"
	// "fmt"
	// "github.com/airdispatch/dispatcher/models"
	// "github.com/airdispatch/dispatcher/views"
	"github.com/airdispatch/go-pressure"
	_ "github.com/lib/pq"
	"os"
	"path/filepath"
	// "airdispat.ch/common"
)

var flag_port = flag.String("port", "2048", "specify the port that the server should run on")

// var db_flush = flag.Bool("db_flush", false, "specifies to flush the database")
// var db_create = flag.Bool("db_create", false, "specifies to create the database")

// var me = flag.String("me", "", "specify the location of this running server")

func main() {
	flag.Parse()

	temp_port := os.Getenv("PORT")
	if temp_port == "" {
		temp_port = *flag_port
	}

	// dbMap, err := models.ConnectToDB()
	// if err != nil {
	// 	fmt.Println("Can't Connect to DB")
	// 	fmt.Println("err")
	// 	return
	// }

	temp_wd, _ := os.Getwd()

	// Get Relevant Paths
	template_dir := filepath.Join(temp_wd, "templates")
	static_dir := filepath.Join(temp_wd, "static")

	s := pressure.CreateServer(temp_port, true)
	t_eng := s.CreateTemplateEngine(template_dir, "base.html")

	// Register URLs
	s.RegisterURL(
		// pressure.NewURLRoute("^/project/airdispatch", &ProjectController{tEng}),
		pressure.NewURLRoute("^/$", &HomepageController{t_eng}),
		pressure.NewStaticFileRoute("^/static/", static_dir),
	)

	s.RunServer()
}

type HomepageController struct {
	templates *pressure.TemplateEngine
}

func (c *HomepageController) GetResponse(p *pressure.Request, l *pressure.Logger) (pressure.View, *pressure.HTTPError) {
	return c.templates.NewTemplateView("dashboard.html", nil), nil
}

// <<<<<<< Updated upstream
// func defineRoutes(s *library.Server) {
// 	s.WebServer.Get("/", views.HomePage(s, views.Dashboard(s)))
// =======
// // func defineRoutes(s *library.Server) {
// // 	s.WebServer.Get("/", views.TemplateLoginRequired(s, views.Dashboard(s)))
// >>>>>>> Stashed changes

// // 	s.WebServer.Get("/compose", views.TemplateLoginRequired(s, s.DisplayTemplate("compose.html")))
// // 	s.WebServer.Post("/compose", views.TemplateLoginRequired(s, views.CreateMessage(s)))

// // 	s.WebServer.Get("/subscribe", views.TemplateLoginRequired(s, views.ShowSubscriptions(s)))
// // 	s.WebServer.Post("/subscribe", views.TemplateLoginRequired(s, views.CreateSubscription(s)))

// <<<<<<< Updated upstream
// 	s.WebServer.Get("/inbox", views.TemplateLoginRequired(s, views.ShowFolder(s, "Inbox")))
// 	s.WebServer.Get("/profile", views.TemplateLoginRequired(s, views.ShowFolder(s, "Profile")))

// 	s.WebServer.Get("/alert/([0-9]*)", views.WildcardTemplateLoginRequired(s, views.ShowAlert(s)))

// 	s.WebServer.Get("/message/([0-9]*)", views.WildcardTemplateLoginRequired(s, views.ShowMessage(s)))
// =======
// // 	s.WebServer.Get("/inbox", views.TemplateLoginRequired(s, views.ShowFolder(s, "Inbox")))
// // 	s.WebServer.Get("/sent", views.TemplateLoginRequired(s, views.ShowFolder(s, "Sent Messages")))

// // 	s.WebServer.Get("/alert/([0-9]*)", views.WildcardTemplateLoginRequired(s, views.ShowAlert(s)))
// >>>>>>> Stashed changes

// 	s.WebServer.Get("/message/([0-9]*)", views.WildcardTemplateLoginRequired(s, views.ShowMessage(s)))

// 	s.WebServer.Get("/message/([0-9]*)/edit", views.WildcardTemplateLoginRequired(s, views.DisplayEditMessage(s)))
// 	s.WebServer.Post("/message/([0-9]*)/edit", views.WildcardTemplateLoginRequired(s, views.UpdateMessage(s)))

// 	s.WebServer.Get("/message/([0-9]*)/delete", views.WildcardTemplateLoginRequired(s, views.DeleteMessage(s)))

// 	s.WebServer.Get("/account/tracker_registration", views.TemplateLoginRequired(s, views.RegisterWithTrackers(s)))

// 	s.WebServer.Get("/login", s.DisplayTemplate("login.html"))
// 	s.WebServer.Post("/login", views.LoginView(s))

// 	s.WebServer.Post("/account/create", views.RegisterUser(s))

// 	s.WebServer.Get("/logout", views.TemplateLoginRequired(s, views.LogoutView(s)))
// }
