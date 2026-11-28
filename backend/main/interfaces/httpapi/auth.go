package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/domain/account"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/service/auth_service"
	"github.com/gin-gonic/gin"
)

const (
	sessionCookie = "magistral_session"

	maxAuthBody = 4096
)

type AuthQueries interface {
	Register(ctx context.Context, credentials account.Credentials, ip string) (account.User, account.Session, error)
	Login(ctx context.Context, credentials account.Credentials, ip string) (account.User, account.Session, error)
	Current(ctx context.Context, sessionID string) (account.User, error)
	Logout(ctx context.Context, sessionID string) error
}

type AuthOptions struct {
	// SecureCookies marks the session cookie Secure; enable it whenever the
	// frontend is served over HTTPS.
	SecureCookies bool
	// CookieDomain limits the cookie to a host; empty uses the request host
	// without subdomains, which keeps development on localhost working.
	CookieDomain string
}

func registerAuthRoutes(router *gin.Engine, auth AuthQueries, options AuthOptions) {

	router.POST("/api/auth/register", func(c *gin.Context) {
		credentials, ok := decodeCredentials(c)

		if !ok {
			return
		}

		user, session, err := auth.Register(c.Request.Context(), credentials, c.ClientIP())

		if err != nil {
			respondAuthError(c, err)

			return
		}

		setSessionCookie(c, session.ID, session.ExpiresAt, options)
		c.JSON(http.StatusCreated, publicUser(user))
	})

	router.POST("/api/auth/login", func(c *gin.Context) {
		credentials, ok := decodeCredentials(c)

		if !ok {
			return
		}

		user, session, err := auth.Login(c.Request.Context(), credentials, c.ClientIP())

		if err != nil {
			respondAuthError(c, err)

			return
		}

		setSessionCookie(c, session.ID, session.ExpiresAt, options)
		c.JSON(http.StatusOK, publicUser(user))
	})

	router.GET("/api/auth/me", func(c *gin.Context) {
		user, err := auth.Current(c.Request.Context(), sessionID(c))

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация."})

			return
		}

		c.JSON(http.StatusOK, publicUser(user))
	})

	router.POST("/api/auth/logout", func(c *gin.Context) {
		_ = auth.Logout(c.Request.Context(), sessionID(c))

		setSessionCookie(c, "", time.Time{}, options)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}

func decodeCredentials(c *gin.Context) (account.Credentials, bool) {

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxAuthBody)

	var request struct {
		Email    *string `json:"email"`
		Password *string `json:"password"`
	}

	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&request); err != nil || request.Email == nil || request.Password == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Укажите email и пароль."})

		return account.Credentials{}, false
	}

	return account.Credentials{Email: *request.Email, Password: *request.Password}, true
}

func respondAuthError(c *gin.Context, err error) {

	if errors.Is(err, auth_service.ErrInvalidCredentials) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный email или пароль."})

		return
	}

	if errors.Is(err, auth_service.ErrEmailTaken) {
		c.JSON(http.StatusConflict, gin.H{"error": "Этот email уже зарегистрирован."})

		return
	}

	if errors.Is(err, auth_service.ErrInvalidInput) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный email или пароль (минимум 8 символов)."})

		return
	}

	respondError(c, err)
}

func publicUser(user account.User) gin.H {
	return gin.H{"id": user.ID, "email": user.Email, "name": user.Name}
}

func setSessionCookie(c *gin.Context, value string, expires time.Time, options AuthOptions) {

	cookie := &http.Cookie{
		Name:     sessionCookie,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   options.SecureCookies,
		Domain:   options.CookieDomain,
		SameSite: http.SameSiteLaxMode,
	}

	if value == "" {
		cookie.Expires = time.Unix(0, 0)
		cookie.MaxAge = -1
	} else {
		cookie.Expires = expires
	}

	http.SetCookie(c.Writer, cookie)
}

func sessionID(c *gin.Context) string {

	cookie, err := c.Cookie(sessionCookie)

	if err != nil {
		return ""
	}

	return cookie
}
