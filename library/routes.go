package library

func (s *Server)defineRoutes() {
	s.webServer.Get("/", s.displayTemplate("login.html"))
	s.webServer.Get("/login", s.displayTemplate("login.html"))
	s.webServer.Post("/login", s.displayTemplate("login.html"))
}