package auth

import (
	"context"
	"testing"
	"time"

	"social-be/internal/domain/user"
	"social-be/internal/pkg/cache"
	"social-be/internal/pkg/pagination"
	"social-be/internal/pkg/query"
	"social-be/internal/pkg/security"

	"github.com/redis/go-redis/v9"
)

func init() {
	// The user service touches the global cache; point it at an unreachable
	// server so cache ops fail fast and degrade gracefully (as in production),
	// letting reads fall through to the repository in these unit tests.
	cache.RDB = redis.NewClient(&redis.Options{Addr: "127.0.0.1:1", DialTimeout: 100 * time.Millisecond})
}

// stubUserRepo implements user.Repository, recording last-login writes.
type stubUserRepo struct {
	user         user.User
	passwordHash string
	role         int

	lastLoginCalls int
	gotID          int
	gotIP          string
	gotUA          string
}

func (s *stubUserRepo) GetByEmailOrVIS(ctx context.Context, identifier string) (*user.User, string, int, error) {
	u := s.user
	return &u, s.passwordHash, s.role, nil
}
func (s *stubUserRepo) GetByID(ctx context.Context, id int) (*user.User, error) {
	u := s.user
	return &u, nil
}
func (s *stubUserRepo) UpdateLastLogin(ctx context.Context, userID int, ip, userAgent string) error {
	s.lastLoginCalls++
	s.gotID, s.gotIP, s.gotUA = userID, ip, userAgent
	return nil
}

// unused interface methods
func (s *stubUserRepo) Create(ctx context.Context, req user.CreateRequest, passwordHash string) error {
	return nil
}
func (s *stubUserRepo) GetAll(ctx context.Context) ([]user.User, error) { return nil, nil }
func (s *stubUserRepo) GetPaginated(ctx context.Context, page pagination.Query, filters query.Filters) ([]user.User, int64, error) {
	return nil, 0, nil
}
func (s *stubUserRepo) GetByEmail(ctx context.Context, email string) (*user.User, string, int, error) {
	return nil, "", 0, nil
}
func (s *stubUserRepo) UpdateProfilePhoto(ctx context.Context, userID int, profilePhoto string) error {
	return nil
}

func newAuthService(repo *stubUserRepo) *Service {
	return NewService(&user.Service{Repo: repo}, nil)
}

// Successful login records last-login with the IP/User-Agent from LoginMeta,
// and the login response contract is unchanged.
func TestLogin_Success_RecordsLastLogin(t *testing.T) {
	hash, err := security.HashPassword("secret123")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	repo := &stubUserRepo{
		user:         user.User{ID: 42, Email: "admin@x.io", Role: 1},
		passwordHash: hash,
		role:         1, // admin, so the volunteer branch is skipped
	}
	svc := newAuthService(repo)

	resp, err := svc.Login(context.Background(),
		LoginRequest{Email: "admin@x.io", Password: "secret123"},
		LoginMeta{IP: "198.51.100.9", UserAgent: "unit-test-agent"},
	)
	if err != nil {
		t.Fatalf("unexpected login error: %v", err)
	}

	if repo.lastLoginCalls != 1 {
		t.Fatalf("UpdateLastLogin called %d times, want 1", repo.lastLoginCalls)
	}
	if repo.gotID != 42 || repo.gotIP != "198.51.100.9" || repo.gotUA != "unit-test-agent" {
		t.Fatalf("last-login args wrong: id=%d ip=%q ua=%q", repo.gotID, repo.gotIP, repo.gotUA)
	}
	// Response contract intact.
	if resp.AccessToken == "" || resp.RefreshToken == "" {
		t.Fatalf("tokens missing: %+v", resp)
	}
}

// Failed login (wrong password) must NOT record last-login.
func TestLogin_Failure_DoesNotRecordLastLogin(t *testing.T) {
	hash, err := security.HashPassword("secret123")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	repo := &stubUserRepo{
		user:         user.User{ID: 42, Email: "admin@x.io", Role: 1},
		passwordHash: hash,
		role:         1,
	}
	svc := newAuthService(repo)

	_, err = svc.Login(context.Background(),
		LoginRequest{Email: "admin@x.io", Password: "WRONG"},
		LoginMeta{IP: "198.51.100.9", UserAgent: "unit-test-agent"},
	)
	if err == nil {
		t.Fatalf("expected invalid-credentials error, got nil")
	}
	if repo.lastLoginCalls != 0 {
		t.Fatalf("UpdateLastLogin must not run on failed login, ran %d times", repo.lastLoginCalls)
	}
}
