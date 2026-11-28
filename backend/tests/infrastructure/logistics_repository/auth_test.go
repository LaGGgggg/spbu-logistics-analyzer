package logistics_repository_test

import (
	"context"
	"crypto/rand"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/domain/account"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/repository/logistics_repository"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/service/auth_service"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestAuthLifecycle(t *testing.T) {

	url := os.Getenv("TEST_DATABASE_URL")

	if url == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL integration")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	defer cancel()

	admin, err := pgxpool.New(ctx, url)

	if err != nil {
		t.Fatal(err)
	}

	defer admin.Close()

	schema := "auth_test_" + rand.Text()
	quoted := pgx.Identifier{schema}.Sanitize()

	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+quoted); err != nil {
		t.Fatal(err)
	}

	defer func() {
		_, _ = admin.Exec(context.Background(), "DROP SCHEMA "+quoted+" CASCADE")
	}()

	cfg, err := pgxpool.ParseConfig(url)

	if err != nil {
		t.Fatal(err)
	}

	cfg.ConnConfig.RuntimeParams["search_path"] = quoted
	pool, err := pgxpool.NewWithConfig(ctx, cfg)

	if err != nil {
		t.Fatal(err)
	}

	defer pool.Close()

	repo := logistics_repository.New(pool)

	if _, err := pool.Exec(ctx, "CREATE TABLE users (id text PRIMARY KEY, email text NOT NULL UNIQUE, name text NOT NULL, password_hash text NOT NULL, created_at timestamptz NOT NULL DEFAULT now())"); err != nil {
		t.Fatal(err)
	}

	if _, err := pool.Exec(ctx, "CREATE TABLE sessions (id text PRIMARY KEY, user_id text NOT NULL REFERENCES users(id) ON DELETE CASCADE, created_at timestamptz NOT NULL DEFAULT now(), expires_at timestamptz NOT NULL)"); err != nil {
		t.Fatal(err)
	}

	if _, err := pool.Exec(ctx, "CREATE TABLE auth_audit (id bigserial PRIMARY KEY, user_id text, email text NOT NULL, action text NOT NULL, ip text NOT NULL, at timestamptz NOT NULL DEFAULT now())"); err != nil {
		t.Fatal(err)
	}

	svc := auth_service.NewService(repo)
	email := "User@Example.com"

	if _, _, err := svc.Login(ctx, account.Credentials{Email: email, Password: "password123"}, "127.0.0.1"); !errors.Is(err, account.ErrInvalidCredentials) {
		t.Fatalf("unknown user login: %v", err)
	}

	user, session, err := svc.Register(ctx, account.Credentials{Email: email, Password: "password123"}, "127.0.0.1")

	if err != nil {
		t.Fatal(err)
	}

	if user.Email != "user@example.com" || session.ID == "" {
		t.Fatalf("user=%+v session=%+v", user, session)
	}

	if _, _, err := svc.Register(ctx, account.Credentials{Email: email, Password: "password123"}, "127.0.0.1"); !errors.Is(err, account.ErrEmailTaken) {
		t.Fatalf("duplicate email: %v", err)
	}

	if _, _, err := svc.Login(ctx, account.Credentials{Email: email, Password: "wrong-password"}, "127.0.0.1"); !errors.Is(err, account.ErrInvalidCredentials) {
		t.Fatalf("wrong password: %v", err)
	}

	loggedIn, _, err := svc.Login(ctx, account.Credentials{Email: email, Password: "password123"}, "127.0.0.1")

	if err != nil || loggedIn.ID != user.ID {
		t.Fatalf("login: %v %+v", err, loggedIn)
	}

	current, err := svc.Current(ctx, session.ID)

	if err != nil || current.ID != user.ID {
		t.Fatalf("current: %v %+v", err, current)
	}

	if _, err := svc.Current(ctx, "missing"); !errors.Is(err, account.ErrUnauthorized) {
		t.Fatalf("unknown session accepted: %v", err)
	}

	if _, err := svc.Current(ctx, ""); !errors.Is(err, account.ErrUnauthorized) {
		t.Fatalf("empty session accepted: %v", err)
	}

	if err := svc.Logout(ctx, session.ID); err != nil {
		t.Fatal(err)
	}

	if _, err := svc.Current(ctx, session.ID); !errors.Is(err, account.ErrUnauthorized) {
		t.Fatalf("session survived logout: %v", err)
	}

	if _, _, err := svc.Register(ctx, account.Credentials{Email: "not-an-email", Password: "password123"}, "127.0.0.1"); !errors.Is(err, account.ErrInvalidInput) {
		t.Fatalf("invalid email: %v", err)
	}

	if _, _, err := svc.Register(ctx, account.Credentials{Email: "short@example.com", Password: "short"}, "127.0.0.1"); !errors.Is(err, account.ErrInvalidInput) {
		t.Fatalf("short password: %v", err)
	}

	var stored string

	if err := pool.QueryRow(ctx, "SELECT password_hash FROM users WHERE id=$1", user.ID).Scan(&stored); err != nil {
		t.Fatal(err)
	}

	if !strings.HasPrefix(stored, "$2") || strings.Contains(stored, "password123") {
		t.Fatalf("password not hashed: %q", stored)
	}

	var failed int

	if err := pool.QueryRow(ctx, "SELECT count(*) FROM auth_audit WHERE action='login_failed'").Scan(&failed); err != nil {
		t.Fatal(err)
	}

	if failed != 2 {
		t.Fatalf("login_failed audits=%d", failed)
	}
}
