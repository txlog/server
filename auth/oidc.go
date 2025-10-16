package auth

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"database/sql"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	logger "github.com/txlog/server/logger"
	"github.com/txlog/server/models"
	"golang.org/x/oauth2"
)

type OIDCService struct {
	Provider     *oidc.Provider
	OAuth2Config oauth2.Config
	Verifier     *oidc.IDTokenVerifier
	DB           *sql.DB
	HTTPClient   *http.Client
}

// NewOIDCService creates a new OIDC service instance
// Returns nil if OIDC is not configured (optional authentication)
//
// Environment Variables:
//   - OIDC_CLIENT_ID: OAuth2 client ID
//   - OIDC_CLIENT_SECRET: OAuth2 client secret
//   - OIDC_ISSUER_URL: OIDC provider issuer URL (default: http://localhost:8090)
//   - OIDC_REDIRECT_URL: OAuth2 redirect URL (default: http://localhost:8080/auth/callback)
//   - OIDC_SKIP_TLS_VERIFY: Skip TLS certificate verification (default: false)
//     Set to "true" for self-signed certificates in production environments
func NewOIDCService(db *sql.DB) (*OIDCService, error) {
	clientID := os.Getenv("OIDC_CLIENT_ID")
	clientSecret := os.Getenv("OIDC_CLIENT_SECRET")

	// If OIDC credentials are not provided, return nil (no error)
	// This allows the system to work without authentication
	if clientID == "" || clientSecret == "" {
		return nil, nil
	}

	ctx := context.Background()

	issuerURL := os.Getenv("OIDC_ISSUER_URL")
	if issuerURL == "" {
		issuerURL = "http://localhost:8090" // Default PocketID URL
	}

	redirectURL := os.Getenv("OIDC_REDIRECT_URL")
	if redirectURL == "" {
		redirectURL = "http://localhost:8080/auth/callback"
	}

	// Create HTTP client with TLS configuration
	httpClient := &http.Client{}

	// Check if we should skip TLS verification (useful for self-signed certificates in production)
	skipTLSVerify := strings.ToLower(os.Getenv("OIDC_SKIP_TLS_VERIFY")) == "true"
	if skipTLSVerify {
		httpClient.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	}

	// Create context with custom HTTP client for OIDC provider
	ctx = oidc.ClientContext(ctx, httpClient)

	provider, err := oidc.NewProvider(ctx, issuerURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC provider: %w", err)
	}

	oauth2Config := oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	verifier := provider.Verifier(&oidc.Config{ClientID: clientID})

	return &OIDCService{
		Provider:     provider,
		OAuth2Config: oauth2Config,
		Verifier:     verifier,
		DB:           db,
		HTTPClient:   httpClient,
	}, nil
}

// IsConfigured checks if OIDC is properly configured
func IsConfigured() bool {
	clientID := os.Getenv("OIDC_CLIENT_ID")
	clientSecret := os.Getenv("OIDC_CLIENT_SECRET")
	return clientID != "" && clientSecret != ""
}

// GetAuthURL generates the authorization URL for OIDC flow
func (s *OIDCService) GetAuthURL(state string) string {
	return s.OAuth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

// ExchangeCodeForTokens exchanges authorization code for tokens
func (s *OIDCService) ExchangeCodeForTokens(ctx context.Context, code string) (*oauth2.Token, error) {
	// Use custom HTTP client if configured
	if s.HTTPClient != nil {
		ctx = context.WithValue(ctx, oauth2.HTTPClient, s.HTTPClient)
	}
	return s.OAuth2Config.Exchange(ctx, code)
}

// VerifyIDToken verifies and extracts claims from ID token
func (s *OIDCService) VerifyIDToken(ctx context.Context, rawIDToken string) (*oidc.IDToken, error) {
	return s.Verifier.Verify(ctx, rawIDToken)
}

// CreateOrUpdateUser creates or updates user from OIDC claims
func (s *OIDCService) CreateOrUpdateUser(ctx context.Context, idToken *oidc.IDToken) (*models.User, error) {
	var claims struct {
		Sub     string `json:"sub"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}

	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to parse ID token claims: %w", err)
	}

	// Validate required fields
	if claims.Sub == "" {
		return nil, fmt.Errorf("OIDC subject (sub) claim is empty")
	}
	if claims.Email == "" {
		return nil, fmt.Errorf("OIDC email claim is empty")
	}
	if claims.Name == "" {
		return nil, fmt.Errorf("OIDC name claim is empty")
	}

	// Check if user already exists by email
	existingUser, err := s.getUserByEmail(claims.Email)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check existing user for email '%s': %w", claims.Email, err)
	}

	now := time.Now()

	if existingUser != nil {
		// Update existing user with OIDC info
		updateQuery := `
			UPDATE users 
			SET sub = $1, name = $2, picture = $3, updated_at = $4, last_login_at = $5
			WHERE email = $6
			RETURNING id, sub, email, name, COALESCE(picture, '') as picture, is_active, is_admin, created_at, updated_at, last_login_at
		`

		user := &models.User{}
		err = s.DB.QueryRow(updateQuery, claims.Sub, claims.Name, claims.Picture, now, now, claims.Email).Scan(
			&user.ID, &user.Sub, &user.Email, &user.Name, &user.Picture,
			&user.IsActive, &user.IsAdmin, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to update existing user (email: %s): %w", claims.Email, err)
		}

		return user, nil
	}

	// Check if user already exists by sub
	existingUser, err = s.getUserBySub(claims.Sub)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check existing user for sub '%s': %w", claims.Sub, err)
	}

	if existingUser != nil {
		// Update existing user
		updateQuery := `
			UPDATE users 
			SET email = $2, name = $3, picture = $4, updated_at = $5, last_login_at = $6
			WHERE sub = $1
			RETURNING id, sub, email, name, COALESCE(picture, '') as picture, is_active, is_admin, created_at, updated_at, last_login_at
		`

		user := &models.User{}
		err = s.DB.QueryRow(updateQuery, claims.Sub, claims.Email, claims.Name, claims.Picture, now, now).Scan(
			&user.ID, &user.Sub, &user.Email, &user.Name, &user.Picture,
			&user.IsActive, &user.IsAdmin, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to update existing user (sub: %s, email: %s): %w", claims.Sub, claims.Email, err)
		}

		return user, nil
	}

	// Create new user
	// Check if this is the first user (should be admin)
	var userCount int
	countQuery := `SELECT COUNT(*) FROM users`
	err = s.DB.QueryRow(countQuery).Scan(&userCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count existing users: %w", err)
	}

	// First user should be admin, others are regular users
	isAdmin := userCount == 0

	insertQuery := `
		INSERT INTO users (sub, email, name, picture, is_active, is_admin, created_at, updated_at, last_login_at)
		VALUES ($1, $2, $3, $4, true, $5, $6, $6, $6)
		RETURNING id, sub, email, name, COALESCE(picture, '') as picture, is_active, is_admin, created_at, updated_at, last_login_at
	`

	user := &models.User{}
	err = s.DB.QueryRow(insertQuery, claims.Sub, claims.Email, claims.Name, claims.Picture, isAdmin, now).Scan(
		&user.ID, &user.Sub, &user.Email, &user.Name, &user.Picture,
		&user.IsActive, &user.IsAdmin, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create new user (sub: %s, email: %s): %w", claims.Sub, claims.Email, err)
	}

	// Log if this is the first admin user
	if isAdmin {
		logger.Info(fmt.Sprintf("First user created as administrator: %s (%s)", user.Name, user.Email))
	}

	return user, nil
}

// CreateUserSession creates a new user session
func (s *OIDCService) CreateUserSession(userID int) (string, error) {
	sessionID, err := generateSessionID()
	if err != nil {
		return "", fmt.Errorf("failed to generate session ID: %w", err)
	}

	expiresAt := time.Now().Add(24 * time.Hour * 7) // 7 days

	query := `
		INSERT INTO user_sessions (id, user_id, created_at, expires_at, is_active)
		VALUES ($1, $2, $3, $4, true)
	`

	_, err = s.DB.Exec(query, sessionID, userID, time.Now(), expiresAt)
	if err != nil {
		return "", fmt.Errorf("failed to create user session: %w", err)
	}

	return sessionID, nil
}

// InvalidateUserSession invalidates a user session
func (s *OIDCService) InvalidateUserSession(sessionID string) error {
	query := `UPDATE user_sessions SET is_active = false WHERE id = $1`
	_, err := s.DB.Exec(query, sessionID)
	return err
}

func (s *OIDCService) getUserBySub(sub string) (*models.User, error) {
	query := `
		SELECT id, sub, email, name, COALESCE(picture, '') as picture, is_active, is_admin, created_at, updated_at, last_login_at
		FROM users WHERE sub = $1
	`

	user := &models.User{}
	err := s.DB.QueryRow(query, sub).Scan(
		&user.ID, &user.Sub, &user.Email, &user.Name, &user.Picture,
		&user.IsActive, &user.IsAdmin, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt,
	)

	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}

	return user, err
}

func (s *OIDCService) getUserByEmail(email string) (*models.User, error) {
	query := `
		SELECT id, sub, email, name, COALESCE(picture, '') as picture, is_active, is_admin, created_at, updated_at, last_login_at
		FROM users WHERE email = $1
	`

	user := &models.User{}
	err := s.DB.QueryRow(query, email).Scan(
		&user.ID, &user.Sub, &user.Email, &user.Name, &user.Picture,
		&user.IsActive, &user.IsAdmin, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt,
	)

	if err == sql.ErrNoRows {
		return nil, sql.ErrNoRows
	}

	return user, err
}

func generateSessionID() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// GenerateState generates a random state parameter for OIDC flow
func GenerateState() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
