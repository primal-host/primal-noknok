package server

func (s *Server) registerRoutes() {
	s.echo.GET("/health", s.handleHealth)
	s.echo.GET("/auth", s.handleAuth)
	s.echo.GET("/login", s.handleLoginPage)
	s.echo.POST("/login", s.handleLogin)
	s.echo.POST("/logout", s.handleLogout)
	s.echo.GET("/", s.handlePortal)

	// OAuth endpoints.
	s.echo.GET("/oauth/callback", s.handleOAuthCallback)
	s.echo.GET("/.well-known/oauth-client-metadata", s.handleClientMetadata)
	s.echo.GET("/oauth/jwks.json", s.handleJWKS)
}
