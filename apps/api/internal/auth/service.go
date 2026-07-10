package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Config holds auth-specific configuration values.
type Config struct {
	JWTSecret            string
	AccessTokenTTL       time.Duration
	RefreshTokenTTL      time.Duration
}

// ErrInvalidCredentials is returned when email or password does not match.
var ErrInvalidCredentials = errors.New("invalid email or password")

// ErrAccountInactive is returned when a user account is deactivated.
var ErrAccountInactive = errors.New("account is inactive")

// ErrTokenInvalid is returned for bad, expired, or revoked refresh tokens.
var ErrTokenInvalid = errors.New("token is invalid or expired")

// Service contains the business logic for authentication.
type Service struct {
	repo *Repository
	cfg  Config
}

// NewService creates a new auth service.
func NewService(repo *Repository, cfg Config) *Service {
	return &Service{repo: repo, cfg: cfg}
}

// Login validates credentials and returns a token pair.
func (s *Service) Login(ctx context.Context, email, password string) (*TokenPair, *User, error) {
	user, err := s.repo.FindUserByEmail(ctx, email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, nil, fmt.Errorf("login: %w", err)
	}
	if !user.IsActive {
		return nil, nil, ErrAccountInactive
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, nil, ErrInvalidCredentials
	}

	pair, err := s.issueTokenPair(ctx, user)
	if err != nil {
		return nil, nil, err
	}

	_ = s.repo.UpdateLastLogin(ctx, user.ID)
	return pair, user, nil
}

// Refresh validates the opaque refresh token, revokes it, and issues a new pair.
func (s *Service) Refresh(ctx context.Context, plainRefreshToken string) (*TokenPair, error) {
	hash := HashToken(plainRefreshToken)
	rt, err := s.repo.FindRefreshToken(ctx, hash)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrTokenInvalid
	}
	if err != nil {
		return nil, fmt.Errorf("refresh: %w", err)
	}
	if rt.RevokedAt != nil || rt.ExpiresAt.Before(time.Now().UTC()) {
		return nil, ErrTokenInvalid
	}

	// Revoke the consumed token (rotation)
	if err := s.repo.RevokeRefreshToken(ctx, hash); err != nil {
		return nil, fmt.Errorf("refresh: revoke old: %w", err)
	}

	user, err := s.repo.FindUserByID(ctx, rt.UserID)
	if err != nil {
		return nil, fmt.Errorf("refresh: find user: %w", err)
	}
	if !user.IsActive {
		return nil, ErrAccountInactive
	}

	return s.issueTokenPair(ctx, user)
}

// Logout revokes all refresh tokens for the user identified by the given plain token.
func (s *Service) Logout(ctx context.Context, plainRefreshToken string) error {
	if plainRefreshToken == "" {
		return nil
	}
	hash := HashToken(plainRefreshToken)
	return s.repo.RevokeRefreshToken(ctx, hash)
}

// LogoutAll revokes every active refresh token for a user.
func (s *Service) LogoutAll(ctx context.Context, userID uint64) error {
	return s.repo.RevokeAllUserRefreshTokens(ctx, userID)
}

// Me returns the full user profile with role assignments.
func (s *Service) Me(ctx context.Context, userID uint64) (*MeResponse, error) {
	user, err := s.repo.FindUserByID(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("me: %w", err)
	}
	roles, err := s.repo.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("me: roles: %w", err)
	}
	return &MeResponse{
		ID:             user.ID,
		OrganizationID: user.OrganizationID,
		Name:           user.Name,
		Email:          user.Email,
		Roles:          roles,
	}, nil
}

// HashPassword hashes a plaintext password with bcrypt.
func HashPassword(plain string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(h), nil
}

// issueTokenPair generates a fresh access + refresh token pair and persists the refresh token.
func (s *Service) issueTokenPair(ctx context.Context, user *User) (*TokenPair, error) {
	access, err := GenerateAccessToken(
		s.cfg.JWTSecret, s.cfg.AccessTokenTTL,
		user.ID, user.OrganizationID, user.Email, user.Name,
	)
	if err != nil {
		return nil, fmt.Errorf("issue token pair: access: %w", err)
	}

	plain, hash, err := GenerateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("issue token pair: refresh: %w", err)
	}

	expiresAt := time.Now().UTC().Add(s.cfg.RefreshTokenTTL)
	if err := s.repo.StoreRefreshToken(ctx, user.ID, hash, expiresAt); err != nil {
		return nil, fmt.Errorf("issue token pair: store: %w", err)
	}

	return &TokenPair{
		AccessToken:  access,
		ExpiresIn:    int(s.cfg.AccessTokenTTL.Seconds()),
		RefreshToken: plain,
	}, nil
}
