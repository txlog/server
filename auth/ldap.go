package auth

import (
	"crypto/tls"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-ldap/ldap/v3"
	logger "github.com/txlog/server/logger"
	"github.com/txlog/server/models"
)

// CategorizeAuthError categorizes LDAP authentication errors for user feedback
func CategorizeAuthError(err error) string {
	if err == nil {
		return "auth_failed"
	}

	errStr := err.Error()

	// Check for configuration issues
	if strings.Contains(errStr, "Naming Violation") ||
		strings.Contains(errStr, "Not a subtree of the base tree") {
		return "ldap_config_error"
	}

	if strings.Contains(errStr, "failed to connect to LDAP") ||
		strings.Contains(errStr, "dial tcp") ||
		strings.Contains(errStr, "connection refused") {
		return "ldap_connection_error"
	}

	if strings.Contains(errStr, "failed to bind with service account") {
		return "ldap_bind_error"
	}

	if strings.Contains(errStr, "invalid credentials") {
		return "invalid_credentials"
	}

	if strings.Contains(errStr, "user not found") {
		return "user_not_found"
	}

	if strings.Contains(errStr, "not a member of any authorized group") {
		return "unauthorized_group"
	}

	// Default fallback
	return "auth_failed"
}

type LDAPService struct {
	DB *sql.DB
}

// NewLDAPService creates a new LDAP service instance
// Returns nil if LDAP is not configured (optional authentication)
//
// Environment Variables:
//   - LDAP_HOST: LDAP server host (e.g., ldap.example.com)
//   - LDAP_PORT: LDAP server port (default: 389 for LDAP, 636 for LDAPS)
//   - LDAP_USE_TLS: Use TLS connection (default: false)
//   - LDAP_SKIP_TLS_VERIFY: Skip TLS certificate verification (default: false)
//   - LDAP_BIND_DN: Bind DN for LDAP authentication (e.g., cn=admin,dc=example,dc=com)
//   - LDAP_BIND_PASSWORD: Password for bind DN
//   - LDAP_BASE_DN: Base DN for user searches (e.g., ou=users,dc=example,dc=com)
//   - LDAP_USER_FILTER: LDAP filter for user search (default: (uid=%s))
//   - LDAP_ADMIN_GROUP: DN of admin group (e.g., cn=admins,ou=groups,dc=example,dc=com)
//   - LDAP_VIEWER_GROUP: DN of viewer group (e.g., cn=viewers,ou=groups,dc=example,dc=com)
//   - LDAP_GROUP_FILTER: LDAP filter for group membership (default: (member=%s))
func NewLDAPService(db *sql.DB) (*LDAPService, error) {
	host := os.Getenv("LDAP_HOST")

	// If LDAP host is not provided, return nil (no error)
	if host == "" {
		return nil, nil
	}

	// Validate required configuration
	baseDN := os.Getenv("LDAP_BASE_DN")
	if baseDN == "" {
		return nil, fmt.Errorf("LDAP_BASE_DN is required when LDAP_HOST is set")
	}

	adminGroup := os.Getenv("LDAP_ADMIN_GROUP")
	viewerGroup := os.Getenv("LDAP_VIEWER_GROUP")

	if adminGroup == "" && viewerGroup == "" {
		return nil, fmt.Errorf("at least one of LDAP_ADMIN_GROUP or LDAP_VIEWER_GROUP must be configured")
	}

	return &LDAPService{
		DB: db,
	}, nil
}

// IsLDAPConfigured checks if LDAP is properly configured
func IsLDAPConfigured() bool {
	host := os.Getenv("LDAP_HOST")
	baseDN := os.Getenv("LDAP_BASE_DN")
	adminGroup := os.Getenv("LDAP_ADMIN_GROUP")
	viewerGroup := os.Getenv("LDAP_VIEWER_GROUP")

	return host != "" && baseDN != "" && (adminGroup != "" || viewerGroup != "")
}

// Authenticate authenticates a user against LDAP
func (s *LDAPService) Authenticate(username, password string) (*models.User, error) {
	if username == "" || password == "" {
		return nil, fmt.Errorf("username and password are required")
	}

	logger.Info(fmt.Sprintf("LDAP authentication attempt for user: %s", username))

	// Connect to LDAP
	conn, err := s.connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to LDAP: %w", err)
	}
	defer conn.Close()

	logger.Debug("LDAP connection established")

	// Bind with service account if configured
	bindDN := os.Getenv("LDAP_BIND_DN")
	bindPassword := os.Getenv("LDAP_BIND_PASSWORD")

	if bindDN != "" && bindPassword != "" {
		logger.Debug(fmt.Sprintf("Binding with service account: %s", bindDN))
		err = conn.Bind(bindDN, bindPassword)
		if err != nil {
			return nil, fmt.Errorf("failed to bind with service account: %w", err)
		}
		logger.Debug("Service account bind successful")
	} else {
		logger.Debug("No service account configured, using anonymous bind")
	}

	// Search for user
	userDN, userAttrs, err := s.searchUser(conn, username)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	logger.Info(fmt.Sprintf("User found in LDAP: %s", userDN))

	// Authenticate user
	logger.Debug("Attempting user bind with provided credentials")
	err = conn.Bind(userDN, password)
	if err != nil {
		logger.Warn(fmt.Sprintf("User bind failed for %s: %v", userDN, err))
		return nil, fmt.Errorf("invalid credentials: %w", err)
	}

	logger.Debug("User credentials validated successfully")

	// Re-bind with service account to check group membership
	if bindDN != "" && bindPassword != "" {
		err = conn.Bind(bindDN, bindPassword)
		if err != nil {
			return nil, fmt.Errorf("failed to re-bind with service account: %w", err)
		}
	}

	// Check group membership
	isAdmin, isViewer, err := s.checkGroupMembership(conn, userDN)
	if err != nil {
		return nil, fmt.Errorf("failed to check group membership: %w", err)
	}

	if !isAdmin && !isViewer {
		return nil, fmt.Errorf("user is not a member of any authorized group")
	}

	// Extract user information
	email := s.getAttributeValue(userAttrs, "mail")
	if email == "" {
		email = username + "@local" // Fallback if email is not set
	}

	name := s.getAttributeValue(userAttrs, "cn")
	if name == "" {
		name = s.getAttributeValue(userAttrs, "displayName")
	}
	if name == "" {
		name = username
	}

	// Create or update user in database
	user, err := s.createOrUpdateUser(username, email, name, isAdmin)
	if err != nil {
		return nil, fmt.Errorf("failed to create/update user: %w", err)
	}

	return user, nil
}

func (s *LDAPService) connect() (*ldap.Conn, error) {
	host := os.Getenv("LDAP_HOST")
	port := os.Getenv("LDAP_PORT")
	useTLS := strings.ToLower(os.Getenv("LDAP_USE_TLS")) == "true"
	skipTLSVerify := strings.ToLower(os.Getenv("LDAP_SKIP_TLS_VERIFY")) == "true"

	// Default ports
	if port == "" {
		if useTLS {
			port = "636"
		} else {
			port = "389"
		}
	}

	address := fmt.Sprintf("%s:%s", host, port)

	var conn *ldap.Conn
	var err error

	if useTLS {
		tlsConfig := &tls.Config{
			ServerName:         host,
			InsecureSkipVerify: skipTLSVerify,
		}
		conn, err = ldap.DialTLS("tcp", address, tlsConfig)
	} else {
		conn, err = ldap.Dial("tcp", address)
	}

	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (s *LDAPService) searchUser(conn *ldap.Conn, username string) (string, map[string]string, error) {
	baseDN := os.Getenv("LDAP_BASE_DN")
	userFilter := os.Getenv("LDAP_USER_FILTER")
	if userFilter == "" {
		userFilter = "(uid=%s)"
	}

	filter := fmt.Sprintf(userFilter, ldap.EscapeFilter(username))

	// Log search parameters for debugging
	logger.Debug(fmt.Sprintf("LDAP user search: baseDN=%s, filter=%s", baseDN, filter))

	searchRequest := ldap.NewSearchRequest(
		baseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		filter,
		[]string{"dn", "cn", "mail", "displayName", "uid"},
		nil,
	)

	result, err := conn.Search(searchRequest)
	if err != nil {
		logger.Error(fmt.Sprintf("LDAP search failed: %v", err))
		return "", nil, err
	}

	if len(result.Entries) == 0 {
		logger.Warn(fmt.Sprintf("LDAP user not found with filter '%s' in base '%s'", filter, baseDN))
		return "", nil, fmt.Errorf("user not found")
	}

	if len(result.Entries) > 1 {
		logger.Warn(fmt.Sprintf("Multiple LDAP users found with filter '%s': %d entries", filter, len(result.Entries)))
		return "", nil, fmt.Errorf("multiple users found")
	}

	entry := result.Entries[0]
	userDN := entry.DN

	logger.Debug(fmt.Sprintf("LDAP user found: %s", userDN))

	attrs := make(map[string]string)
	for _, attr := range entry.Attributes {
		if len(attr.Values) > 0 {
			attrs[attr.Name] = attr.Values[0]
		}
	}

	return userDN, attrs, nil
}

func (s *LDAPService) checkGroupMembership(conn *ldap.Conn, userDN string) (bool, bool, error) {
	adminGroup := os.Getenv("LDAP_ADMIN_GROUP")
	viewerGroup := os.Getenv("LDAP_VIEWER_GROUP")
	groupFilter := os.Getenv("LDAP_GROUP_FILTER")
	if groupFilter == "" {
		groupFilter = "(member=%s)"
	}

	isAdmin := false
	isViewer := false

	// Check admin group membership
	if adminGroup != "" {
		isMember, err := s.isGroupMember(conn, userDN, adminGroup, groupFilter)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to check admin group membership: %v", err))
		} else {
			isAdmin = isMember
		}
	}

	// Check viewer group membership
	if viewerGroup != "" {
		isMember, err := s.isGroupMember(conn, userDN, viewerGroup, groupFilter)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to check viewer group membership: %v", err))
		} else {
			isViewer = isMember
		}
	}

	return isAdmin, isViewer, nil
}

func (s *LDAPService) isGroupMember(conn *ldap.Conn, userDN, groupDN, groupFilter string) (bool, error) {
	// Check if the filter uses memberUid (posixGroup) instead of member/uniqueMember
	// posixGroup uses only the uid value, not the full DN
	filterValue := userDN
	if strings.Contains(groupFilter, "memberUid") {
		// Extract uid from DN (e.g., "uid=john,ou=users,dc=example,dc=com" -> "john")
		filterValue = extractUIDFromDN(userDN)
		logger.Debug(fmt.Sprintf("Using memberUid filter, extracted uid: %s from DN: %s", filterValue, userDN))
	}

	filter := fmt.Sprintf(groupFilter, ldap.EscapeFilter(filterValue))

	searchRequest := ldap.NewSearchRequest(
		groupDN,
		ldap.ScopeBaseObject,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		filter,
		[]string{"dn"},
		nil,
	)

	result, err := conn.Search(searchRequest)
	if err != nil {
		// If group doesn't exist or search fails, user is not a member
		return false, err
	}

	return len(result.Entries) > 0, nil
}

// extractUIDFromDN extracts the uid value from a DN
// Example: "uid=john.doe,ou=users,dc=example,dc=com" -> "john.doe"
func extractUIDFromDN(dn string) string {
	// Split by comma to get RDN components
	parts := strings.Split(dn, ",")
	if len(parts) == 0 {
		return dn
	}

	// Get the first component (should be uid=value)
	firstPart := strings.TrimSpace(parts[0])

	// Split by = to separate attribute from value
	kvPair := strings.SplitN(firstPart, "=", 2)
	if len(kvPair) != 2 {
		return dn
	}

	// Check if it's a uid attribute
	attr := strings.ToLower(strings.TrimSpace(kvPair[0]))
	value := strings.TrimSpace(kvPair[1])

	if attr == "uid" {
		return value
	}

	// If not uid, return the whole DN (fallback)
	return dn
}

func (s *LDAPService) getAttributeValue(attrs map[string]string, key string) string {
	if val, ok := attrs[key]; ok {
		return val
	}
	return ""
}

func (s *LDAPService) createOrUpdateUser(username, email, name string, isAdmin bool) (*models.User, error) {
	// Use username as the unique identifier (sub field)
	sub := "ldap:" + username

	// Check if user already exists
	existingUser, err := s.getUserBySub(sub)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	now := time.Now()

	if existingUser != nil {
		// Update existing user
		updateQuery := `
			UPDATE users 
			SET email = $2, name = $3, is_admin = $4, updated_at = $5, last_login_at = $6
			WHERE sub = $1
			RETURNING id, sub, email, name, COALESCE(picture, '') as picture, is_active, is_admin, created_at, updated_at, last_login_at
		`

		user := &models.User{}
		err = s.DB.QueryRow(updateQuery, sub, email, name, isAdmin, now, now).Scan(
			&user.ID, &user.Sub, &user.Email, &user.Name, &user.Picture,
			&user.IsActive, &user.IsAdmin, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to update user: %w", err)
		}

		return user, nil
	}

	// Create new user
	insertQuery := `
		INSERT INTO users (sub, email, name, picture, is_active, is_admin, created_at, updated_at, last_login_at)
		VALUES ($1, $2, $3, '', true, $4, $5, $5, $5)
		RETURNING id, sub, email, name, COALESCE(picture, '') as picture, is_active, is_admin, created_at, updated_at, last_login_at
	`

	user := &models.User{}
	err = s.DB.QueryRow(insertQuery, sub, email, name, isAdmin, now).Scan(
		&user.ID, &user.Sub, &user.Email, &user.Name, &user.Picture,
		&user.IsActive, &user.IsAdmin, &user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	logger.Info(fmt.Sprintf("LDAP user created: %s (%s) - Admin: %v", user.Name, user.Email, isAdmin))

	return user, nil
}

func (s *LDAPService) getUserBySub(sub string) (*models.User, error) {
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

// CreateUserSession creates a new user session (reuses OIDC session logic)
func (s *LDAPService) CreateUserSession(userID int) (string, error) {
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
func (s *LDAPService) InvalidateUserSession(sessionID string) error {
	query := `UPDATE user_sessions SET is_active = false WHERE id = $1`
	_, err := s.DB.Exec(query, sessionID)
	return err
}
