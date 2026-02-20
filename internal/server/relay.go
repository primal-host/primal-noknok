package server

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// handleRelay sets a session cookie on the current domain and redirects.
// Used to relay an authenticated session from the primary domain (where OAuth
// happens) to an external domain (e.g. ker.ai).
//
// GET /__noknok_set?t=SESSION_TOKEN&r=/path
func (s *Server) handleRelay(c echo.Context) error {
	token := c.QueryParam("t")
	redirect := c.QueryParam("r")

	if token == "" {
		return c.NoContent(http.StatusBadRequest)
	}

	// Validate the session token.
	sess, err := s.sess.Validate(c.Request().Context(), token)
	if err != nil {
		return c.Redirect(http.StatusFound, s.cfg.PublicURL+"/login")
	}

	// Determine the cookie domain from the request host.
	host := c.Request().Host
	domain := s.cfg.DomainForHost(host)

	// Set the session cookie for this domain.
	c.SetCookie(s.sess.MakeCookieForDomain(token, sess.ExpiresAt, domain))

	// Redirect must be a relative path to prevent open redirect.
	if redirect == "" || !strings.HasPrefix(redirect, "/") {
		redirect = "/"
	}

	return c.Redirect(http.StatusFound, redirect)
}
