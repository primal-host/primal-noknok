package server

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/primal-host/noknok/internal/session"
)

// handleHealth returns 200 if the server is running.
func (s *Server) handleHealth(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

// handleAuth is the Traefik forwardAuth endpoint.
// Valid session → 200 with X-User-DID and X-User-Handle headers.
// Authorization header present → 200 (let backend validate the token).
// No/invalid session → 302 redirect to login page.
func (s *Server) handleAuth(c echo.Context) error {
	host := c.Request().Header.Get("X-Forwarded-Host")

	// Check if the service is disabled — deny all access regardless of session.
	if host != "" {
		svc, _ := s.db.GetServiceByHost(c.Request().Context(), host)
		if svc != nil && !svc.Enabled {
			accept := c.Request().Header.Get("X-Forwarded-Accept")
			if accept == "" {
				accept = c.Request().Header.Get("Accept")
			}
			if strings.Contains(accept, "text/html") {
				return c.Redirect(http.StatusFound, s.cfg.PublicURL+"/disabled?service="+url.QueryEscape(svc.Name))
			}
			return c.NoContent(http.StatusServiceUnavailable)
		}
	}

	cookie, err := c.Cookie(session.CookieName())
	if err == nil && cookie.Value != "" {
		sess, err := s.sess.Validate(c.Request().Context(), cookie.Value)
		if err == nil {
			// Check if user is owner/admin (full access) or has a grant for this service.
			if host != "" {
				role, roleErr := s.db.GetUserServiceRole(c.Request().Context(), sess.DID, host)
				if roleErr != nil || role == "" {
					// User has no grant for this service — deny access.
					// Redirect browser to portal so they see what they can access.
					accept := c.Request().Header.Get("X-Forwarded-Accept")
					if accept == "" {
						accept = c.Request().Header.Get("Accept")
					}
					if strings.Contains(accept, "text/html") {
						return c.Redirect(http.StatusFound, s.cfg.PublicURL+"/")
					}
					return c.NoContent(http.StatusForbidden)
				}
				c.Response().Header().Set("X-User-Role", role)
			}

			c.Response().Header().Set("X-User-DID", sess.DID)
			c.Response().Header().Set("X-User-Handle", sess.Handle)
			if sess.Username != "" {
				c.Response().Header().Set("X-WEBAUTH-USER", sess.Username)
			}

			return c.NoContent(http.StatusOK)
		}
	}

	// Pass through requests with an Authorization header (e.g. PATs, API tokens)
	// so the backend service can validate them itself.
	if c.Request().Header.Get("X-Forwarded-Authorization") != "" ||
		c.Request().Header.Get("Authorization") != "" {
		return c.NoContent(http.StatusOK)
	}

	// Non-browser clients (git, curl, API) get 401 so they can retry with
	// credentials. The backend (e.g. Gitea) will issue its own WWW-Authenticate
	// challenge once it receives the request.
	accept := c.Request().Header.Get("X-Forwarded-Accept")
	if accept == "" {
		accept = c.Request().Header.Get("Accept")
	}
	if !strings.Contains(accept, "text/html") {
		return c.NoContent(http.StatusUnauthorized)
	}

	// Build redirect URL from forwarded headers.
	scheme := c.Request().Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		scheme = "https"
	}
	if host == "" {
		host = c.Request().Header.Get("X-Forwarded-Host")
	}
	uri := c.Request().Header.Get("X-Forwarded-Uri")

	redirectTarget := ""
	if host != "" {
		redirectTarget = fmt.Sprintf("%s://%s%s", scheme, host, uri)
	}

	loginURL := fmt.Sprintf("%s/login", s.cfg.PublicURL)
	if redirectTarget != "" {
		loginURL += "?redirect=" + url.QueryEscape(redirectTarget)
	}

	return c.Redirect(http.StatusFound, loginURL)
}

// handleDisabled renders a status page for disabled services.
func (s *Server) handleDisabled(c echo.Context) error {
	name := c.QueryParam("service")
	if name == "" {
		name = "This service"
	}
	return c.HTML(http.StatusOK, disabledHTML(name))
}

func disabledHTML(serviceName string) string {
	initial := "?"
	if len(serviceName) > 0 {
		initial = string([]rune(serviceName)[0])
	}
	return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Service Disabled</title>
<style>
  *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
  body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
    background: #0f172a;
    color: #e2e8f0;
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .card {
    display: flex;
    align-items: center;
    gap: 1rem;
    background: #1e293b;
    border-radius: 12px;
    padding: 1.25rem;
    position: relative;
    min-width: 280px;
    cursor: pointer;
    transition: background 0.15s;
  }
  .card:hover { background: #334155; }
  .icon {
    width: 48px;
    height: 48px;
    background: #3b82f6;
    border-radius: 10px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 1.25rem;
    font-weight: 700;
    color: #fff;
    flex-shrink: 0;
  }
  .info h3 {
    font-size: 1rem;
    font-weight: 600;
    color: #f8fafc;
    margin-bottom: 0.25rem;
  }
  .info p {
    font-size: 0.8125rem;
    color: #94a3b8;
  }
  .disabled-dot {
    position: absolute;
    top: 0.5rem;
    right: 2.5rem;
    width: 1rem;
    height: 1rem;
    border-radius: 4px;
    background: #ef4444;
  }
</style>
</head>
<body>
<div class="card" onclick="goBack()">
  <div class="icon">` + initial + `</div>
  <div class="info">
    <h3>` + serviceName + `</h3>
    <p>Disabled by administrator</p>
  </div>
  <div class="disabled-dot"></div>
</div>
<script>
function goBack() {
  if (typeof BroadcastChannel !== 'undefined') {
    var ch = new BroadcastChannel('noknok_portal');
    ch.postMessage({ type: 'focus' });
  }
  if (window.opener) {
    try { window.close(); return; } catch(e) {}
  }
  window.location.href = '/';
}
</script>
</body>
</html>`
}

// handleLogout destroys the entire session group and redirects to login.
func (s *Server) handleLogout(c echo.Context) error {
	cookie, err := c.Cookie(session.CookieName())
	if err == nil && cookie.Value != "" {
		sess, err := s.sess.Validate(c.Request().Context(), cookie.Value)
		if err == nil && sess.GroupID != "" {
			_ = s.sess.DestroyGroup(c.Request().Context(), sess.GroupID)
		} else {
			_ = s.sess.Destroy(c.Request().Context(), cookie.Value)
		}
	}
	c.SetCookie(s.sess.ClearCookie())
	return c.Redirect(http.StatusFound, s.cfg.PublicURL+"/login")
}
