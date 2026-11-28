package logistics_repository

import (
	"context"
	"errors"

	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/domain/account"
	"github.com/LaGGgggg/spbu-logistics-analyzer/backend/main/service/auth_service"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var _ auth_service.Repository = (*Repository)(nil)

func (p *Repository) CreateUser(ctx context.Context, user account.User) (account.User, error) {

	err := p.pool.QueryRow(ctx,
		"INSERT INTO users(id,email,name,password_hash,created_at) VALUES($1,$2,$3,$4,$5) RETURNING created_at",
		user.ID, user.Email, user.Name, user.PasswordHash, user.CreatedAt,
	).Scan(&user.CreatedAt)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return account.User{}, account.ErrEmailTaken
		}

		return account.User{}, err
	}

	return user, nil
}

func (p *Repository) FindUserByEmail(ctx context.Context, email string) (account.User, error) {
	return p.findUser(ctx, "email", email)
}

func (p *Repository) FindUserByID(ctx context.Context, id string) (account.User, error) {
	return p.findUser(ctx, "id", id)
}

func (p *Repository) findUser(ctx context.Context, column, value string) (account.User, error) {

	columns := map[string]bool{"id": true, "email": true}

	if !columns[column] {
		return account.User{}, account.ErrUnauthorized
	}

	var user account.User

	err := p.pool.QueryRow(ctx,
		"SELECT id,email,name,password_hash,created_at FROM users WHERE "+column+"=$1", value,
	).Scan(&user.ID, &user.Email, &user.Name, &user.PasswordHash, &user.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return account.User{}, account.ErrUnauthorized
	}

	if err != nil {
		return account.User{}, err
	}

	return user, nil
}

func (p *Repository) CreateSession(ctx context.Context, session account.Session) error {
	_, err := p.pool.Exec(ctx,
		"INSERT INTO sessions(id,user_id,expires_at) VALUES($1,$2,$3)",
		session.ID, session.UserID, session.ExpiresAt,
	)

	return err
}

func (p *Repository) FindSession(ctx context.Context, id string) (account.Session, error) {

	var session account.Session

	err := p.pool.QueryRow(ctx,
		"SELECT id,user_id,expires_at FROM sessions WHERE id=$1 AND expires_at>now()", id,
	).Scan(&session.ID, &session.UserID, &session.ExpiresAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return account.Session{}, account.ErrUnauthorized
	}

	if err != nil {
		return account.Session{}, err
	}

	return session, nil
}

func (p *Repository) DeleteSession(ctx context.Context, id string) error {
	_, err := p.pool.Exec(ctx, "DELETE FROM sessions WHERE id=$1", id)

	return err
}

func (p *Repository) Audit(ctx context.Context, userID, email, action, ip string) error {
	_, err := p.pool.Exec(ctx,
		"INSERT INTO auth_audit(user_id,email,action,ip) VALUES($1,$2,$3,$4)",
		userID, email, action, ip,
	)

	return err
}
