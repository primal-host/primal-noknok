package atproto

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bluesky-social/indigo/atproto/auth/oauth"
	"github.com/bluesky-social/indigo/atproto/syntax"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PgStore implements oauth.ClientAuthStore backed by Postgres JSONB.
type PgStore struct {
	pool *pgxpool.Pool
}

// NewPgStore creates a Postgres-backed OAuth store.
func NewPgStore(pool *pgxpool.Pool) *PgStore {
	return &PgStore{pool: pool}
}

func (s *PgStore) GetSession(ctx context.Context, did syntax.DID, sessionID string) (*oauth.ClientSessionData, error) {
	var data []byte
	err := s.pool.QueryRow(ctx,
		`SELECT data FROM oauth_sessions WHERE did = $1 AND session_id = $2`,
		did.String(), sessionID).Scan(&data)
	if err != nil {
		return nil, fmt.Errorf("get oauth session: %w", err)
	}
	var sess oauth.ClientSessionData
	if err := json.Unmarshal(data, &sess); err != nil {
		return nil, fmt.Errorf("unmarshal oauth session: %w", err)
	}
	return &sess, nil
}

func (s *PgStore) SaveSession(ctx context.Context, sess oauth.ClientSessionData) error {
	data, err := json.Marshal(sess)
	if err != nil {
		return fmt.Errorf("marshal oauth session: %w", err)
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO oauth_sessions (did, session_id, data)
		VALUES ($1, $2, $3)
		ON CONFLICT (did, session_id) DO UPDATE SET data = $3
	`, sess.AccountDID.String(), sess.SessionID, data)
	return err
}

func (s *PgStore) DeleteSession(ctx context.Context, did syntax.DID, sessionID string) error {
	_, err := s.pool.Exec(ctx,
		`DELETE FROM oauth_sessions WHERE did = $1 AND session_id = $2`,
		did.String(), sessionID)
	return err
}

func (s *PgStore) GetAuthRequestInfo(ctx context.Context, state string) (*oauth.AuthRequestData, error) {
	var data []byte
	err := s.pool.QueryRow(ctx,
		`SELECT data FROM oauth_requests WHERE state = $1`,
		state).Scan(&data)
	if err != nil {
		return nil, fmt.Errorf("get auth request: %w", err)
	}
	var info oauth.AuthRequestData
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("unmarshal auth request: %w", err)
	}
	return &info, nil
}

func (s *PgStore) SaveAuthRequestInfo(ctx context.Context, info oauth.AuthRequestData) error {
	data, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("marshal auth request: %w", err)
	}
	_, err = s.pool.Exec(ctx,
		`INSERT INTO oauth_requests (state, data) VALUES ($1, $2)`,
		info.State, data)
	return err
}

func (s *PgStore) DeleteAuthRequestInfo(ctx context.Context, state string) error {
	_, err := s.pool.Exec(ctx,
		`DELETE FROM oauth_requests WHERE state = $1`, state)
	return err
}
