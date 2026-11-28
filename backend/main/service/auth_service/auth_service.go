package auth_service

import (
	"context"
	"crypto/rand"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/domain/account"
	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

var (
	ErrInvalidCredentials = account.ErrInvalidCredentials
	ErrEmailTaken         = account.ErrEmailTaken
	ErrUnauthorized       = account.ErrUnauthorized
	ErrInvalidInput       = account.ErrInvalidInput
)

type Repository interface {
	CreateUser(context.Context, account.User) (account.User, error)
	FindUserByEmail(context.Context, string) (account.User, error)
	FindUserByID(context.Context, string) (account.User, error)
	CreateSession(context.Context, account.Session) error
	FindSession(context.Context, string) (account.Session, error)
	DeleteSession(context.Context, string) error
	Audit(context.Context, string, string, string, string) error
}

type Service struct {
	repository Repository
	now        func() time.Time
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository, now: time.Now}
}

func (s *Service) Register(ctx context.Context, credentials account.Credentials, ip string) (account.User, account.Session, error) {

	email, err := account.NormalizeEmail(credentials.Email)

	if err != nil {
		return account.User{}, account.Session{}, err
	}

	if err := account.ValidatePassword(credentials.Password); err != nil {
		return account.User{}, account.Session{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(credentials.Password), bcryptCost)

	if err != nil {
		return account.User{}, account.Session{}, err
	}

	user := account.User{ID: "user_" + rand.Text(), Email: email, Name: nameFromEmail(email), PasswordHash: hash, CreatedAt: s.now()}
	user, err = s.repository.CreateUser(ctx, user)

	if err != nil {
		return account.User{}, account.Session{}, err
	}

	session := account.Session{ID: newSessionID(), UserID: user.ID, ExpiresAt: s.now().Add(account.SessionLifetime)}

	if err := s.repository.CreateSession(ctx, session); err != nil {
		return account.User{}, account.Session{}, err
	}

	s.audit(ctx, user.ID, user.Email, "register", ip)

	return user, session, nil
}

func (s *Service) Login(ctx context.Context, credentials account.Credentials, ip string) (account.User, account.Session, error) {

	email, err := account.NormalizeEmail(credentials.Email)

	if err != nil {
		return account.User{}, account.Session{}, ErrInvalidCredentials
	}

	user, err := s.repository.FindUserByEmail(ctx, email)

	if errors.Is(err, ErrUnauthorized) {
		s.audit(ctx, "", email, "login_failed", ip)

		return account.User{}, account.Session{}, ErrInvalidCredentials
	}

	if err != nil {
		return account.User{}, account.Session{}, err
	}

	if bcrypt.CompareHashAndPassword(user.PasswordHash, []byte(credentials.Password)) != nil {
		s.audit(ctx, user.ID, email, "login_failed", ip)

		return account.User{}, account.Session{}, ErrInvalidCredentials
	}

	session := account.Session{ID: newSessionID(), UserID: user.ID, ExpiresAt: s.now().Add(account.SessionLifetime)}

	if err := s.repository.CreateSession(ctx, session); err != nil {
		return account.User{}, account.Session{}, err
	}

	s.audit(ctx, user.ID, user.Email, "login", ip)

	return user, session, nil
}

func (s *Service) Current(ctx context.Context, sessionID string) (account.User, error) {

	if sessionID == "" {
		return account.User{}, ErrUnauthorized
	}

	session, err := s.repository.FindSession(ctx, sessionID)

	if err != nil {
		return account.User{}, ErrUnauthorized
	}

	user, err := s.repository.FindUserByID(ctx, session.UserID)

	if err != nil {
		return account.User{}, ErrUnauthorized
	}

	return user, nil
}

func (s *Service) Logout(ctx context.Context, sessionID string) error {

	if sessionID == "" {
		return nil
	}

	return s.repository.DeleteSession(ctx, sessionID)
}

func (s *Service) audit(ctx context.Context, userID, email, action, ip string) {

	if err := s.repository.Audit(ctx, userID, email, action, ip); err != nil {
		slog.Error("auth audit", "error", err)
	}
}

func newSessionID() string {
	return rand.Text() + rand.Text()
}

func nameFromEmail(email string) string {

	name, _, _ := strings.Cut(email, "@")

	if len(name) > 64 {
		name = name[:64]
	}

	return name
}
