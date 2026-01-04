package reddit

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

// OAuthClient handles Reddit OAuth API authentication and requests.
// Uses app-only authentication for accessing public r/soccer data.
type OAuthClient struct {
	httpClient   *http.Client
	clientID     string
	clientSecret string
	username     string
	password     string

	// Token management
	accessToken       string
	refreshTokenValue string
	tokenExpiry       time.Time
	tokenMutex        sync.RWMutex

	// Rate limiting (OAuth allows 600 requests per hour)
	rateLimiter *rateLimiter

	// Debug logging
	debugLogger DebugLogger
}

// debugLog is a helper method to safely call the debug logger if it exists
func (c *OAuthClient) debugLog(message string) {
	if c.debugLogger != nil {
		c.debugLogger(message)
	}
}

// OAuthResponse represents Reddit's OAuth token response.
type OAuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// NewOAuthClient creates a new OAuth client if credentials are available.
// Returns nil if OAuth credentials are not configured (falls back to public API).
func NewOAuthClient() (*OAuthClient, error) {
	clientID := os.Getenv("REDDIT_CLIENT_ID")
	clientSecret := os.Getenv("REDDIT_CLIENT_SECRET")
	username := os.Getenv("REDDIT_USERNAME")
	password := os.Getenv("REDDIT_PASSWORD")

	// If any OAuth credentials are missing, return nil (fallback to public API)
	if clientID == "" || clientSecret == "" || username == "" || password == "" {
		return nil, nil
	}

	client := &OAuthClient{
		httpClient: &http.Client{
			Timeout: 15 * time.Second, // Slightly longer timeout for OAuth
		},
		clientID:     clientID,
		clientSecret: clientSecret,
		username:     username,
		password:     password,
		rateLimiter:  newRateLimiter(10), // 10 requests/minute (well under 600/hour limit)
		debugLogger:  nil,                // No debug logger by default
	}

	// Try to authenticate immediately
	if err := client.authenticate(); err != nil {
		return nil, fmt.Errorf("OAuth authentication failed: %w", err)
	}

	return client, nil
}

// NewOAuthClientWithDebug creates a new OAuth client with debug logging enabled.
func NewOAuthClientWithDebug(debugLogger DebugLogger) (*OAuthClient, error) {
	clientID := os.Getenv("REDDIT_CLIENT_ID")
	clientSecret := os.Getenv("REDDIT_CLIENT_SECRET")
	username := os.Getenv("REDDIT_USERNAME")
	password := os.Getenv("REDDIT_PASSWORD")

	// If any OAuth credentials are missing, return nil (fallback to public API)
	if clientID == "" || clientSecret == "" || username == "" || password == "" {
		return nil, nil
	}

	client := &OAuthClient{
		httpClient: &http.Client{
			Timeout: 15 * time.Second, // Slightly longer timeout for OAuth
		},
		clientID:     clientID,
		clientSecret: clientSecret,
		username:     username,
		password:     password,
		rateLimiter:  newRateLimiter(10), // 10 requests/minute (well under 600/hour limit)
		debugLogger:  debugLogger,
	}

	// Try to authenticate immediately
	if err := client.authenticate(); err != nil {
		return nil, fmt.Errorf("OAuth authentication failed: %w", err)
	}

	return client, nil
}

// authenticate performs OAuth authentication with Reddit.
// Obtains access and refresh tokens for app-only API access.
func (c *OAuthClient) authenticate() error {
	c.debugLog("Starting Reddit OAuth authentication")
	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("username", c.username)
	data.Set("password", c.password)

	req, err := http.NewRequest("POST", "https://www.reddit.com/api/v1/access_token", strings.NewReader(data.Encode()))
	if err != nil {
		c.debugLog(fmt.Sprintf("OAuth authentication failed: create auth request: %v", err))
		return fmt.Errorf("create auth request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", fmt.Sprintf("%s:v1.0.0 (by /u/%s)", c.username, c.username))
	req.SetBasicAuth(c.clientID, c.clientSecret)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.debugLog(fmt.Sprintf("OAuth authentication failed: auth request failed: %v", err))
		return fmt.Errorf("auth request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.debugLog(fmt.Sprintf("OAuth authentication failed: status %d: %s", resp.StatusCode, string(body)))
		return fmt.Errorf("auth failed with status %d: %s", resp.StatusCode, string(body))
	}

	var oauthResp OAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&oauthResp); err != nil {
		c.debugLog(fmt.Sprintf("OAuth authentication failed: parse auth response: %v", err))
		return fmt.Errorf("parse auth response: %w", err)
	}

	c.tokenMutex.Lock()
	c.accessToken = oauthResp.AccessToken
	c.refreshTokenValue = oauthResp.RefreshToken
	c.tokenExpiry = time.Now().Add(time.Duration(oauthResp.ExpiresIn) * time.Second)
	c.tokenMutex.Unlock()

	c.debugLog("OAuth authentication successful")
	return nil
}

// refreshToken refreshes the access token using the refresh token.
func (c *OAuthClient) refreshToken() error {
	c.debugLog("Refreshing OAuth access token")
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", c.refreshTokenValue)

	req, err := http.NewRequest("POST", "https://www.reddit.com/api/v1/access_token", strings.NewReader(data.Encode()))
	if err != nil {
		c.debugLog(fmt.Sprintf("OAuth token refresh failed: create refresh request: %v", err))
		return fmt.Errorf("create refresh request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", fmt.Sprintf("%s:v1.0.0 (by /u/%s)", c.username, c.username))
	req.SetBasicAuth(c.clientID, c.clientSecret)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.debugLog(fmt.Sprintf("OAuth token refresh failed: refresh request failed: %v", err))
		return fmt.Errorf("refresh request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// If refresh fails, we'll need to re-authenticate
		body, _ := io.ReadAll(resp.Body)
		c.debugLog(fmt.Sprintf("OAuth token refresh failed: status %d: %s", resp.StatusCode, string(body)))
		return fmt.Errorf("token refresh failed with status %d: %s", resp.StatusCode, string(body))
	}

	var oauthResp OAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&oauthResp); err != nil {
		c.debugLog(fmt.Sprintf("OAuth token refresh failed: parse refresh response: %v", err))
		return fmt.Errorf("parse refresh response: %w", err)
	}

	c.tokenMutex.Lock()
	c.accessToken = oauthResp.AccessToken
	// Refresh token might be updated
	if oauthResp.RefreshToken != "" {
		c.refreshTokenValue = oauthResp.RefreshToken
	}
	c.tokenExpiry = time.Now().Add(time.Duration(oauthResp.ExpiresIn) * time.Second)
	c.tokenMutex.Unlock()

	c.debugLog("OAuth token refresh successful")
	return nil
}

// ensureValidToken ensures we have a valid access token, refreshing if necessary.
func (c *OAuthClient) ensureValidToken() error {
	c.tokenMutex.RLock()
	tokenValid := c.accessToken != "" && time.Now().Before(c.tokenExpiry.Add(-5*time.Minute)) // Refresh 5 mins early
	c.tokenMutex.RUnlock()

	if !tokenValid {
		c.debugLog("OAuth token expired or missing, attempting to refresh/re-authenticate")
		c.tokenMutex.Lock()
		defer c.tokenMutex.Unlock()

		// Double-check after acquiring write lock
		if c.accessToken == "" || time.Now().After(c.tokenExpiry.Add(-5*time.Minute)) {
			if c.refreshTokenValue != "" {
				// Try to refresh token
				c.debugLog("Attempting OAuth token refresh")
				if err := c.refreshToken(); err != nil {
					c.debugLog(fmt.Sprintf("OAuth token refresh failed, attempting full authentication: %v", err))
					// Refresh failed, try full authentication
					return c.authenticate()
				}
			} else {
				// No refresh token, authenticate from scratch
				c.debugLog("No refresh token available, attempting full OAuth authentication")
				return c.authenticate()
			}
		}
	}

	return nil
}

// Search performs a search using Reddit OAuth API.
// Higher rate limits and no CAPTCHA issues compared to public API.
func (c *OAuthClient) Search(query string, limit int, matchTime time.Time) ([]SearchResult, error) {
	c.rateLimiter.wait()

	// Ensure we have a valid token
	if err := c.ensureValidToken(); err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	// Build timestamp range for filtering
	startTime := matchTime.Add(-24 * time.Hour).Unix()
	endTime := matchTime.Add(48 * time.Hour).Unix()

	// OAuth API uses different endpoint and authorization
	searchURL := fmt.Sprintf(
		"https://oauth.reddit.com/r/soccer/search.json?q=%s+flair:Media+timestamp:%d..%d&restrict_sr=on&sort=relevance&limit=%d",
		url.QueryEscape(query),
		startTime,
		endTime,
		limit,
	)

	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create search request: %w", err)
	}

	c.tokenMutex.RLock()
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	c.tokenMutex.RUnlock()

	req.Header.Set("User-Agent", fmt.Sprintf("%s:v1.0.0 (by /u/%s)", c.username, c.username))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OAuth search failed with status %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var searchResp redditSearchResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		return nil, fmt.Errorf("parse search response: %w", err)
	}

	results := make([]SearchResult, 0, len(searchResp.Data.Children))
	for _, child := range searchResp.Data.Children {
		result := child.Data.toSearchResult()
		// Only include posts with Media flair (same filter as public API)
		if result.Flair == "Media" {
			results = append(results, result)
		}
	}

	return results, nil
}

// IsAvailable returns true if OAuth client is properly configured and authenticated.
func (c *OAuthClient) IsAvailable() bool {
	c.tokenMutex.RLock()
	defer c.tokenMutex.RUnlock()
	return c.accessToken != "" && time.Now().Before(c.tokenExpiry)
}
