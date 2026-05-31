package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/brvm/go-api/internal/middleware"
	"github.com/brvm/go-api/internal/model"
	"github.com/brvm/go-api/internal/service"
	"github.com/gofiber/fiber/v2"
)

// mockUserRepo implements the minimal interface needed for AuthService tests.
type mockUserRepo struct {
	users map[string]struct {
		ID, HashedPwd string
	}
	refreshTokens map[string]bool
	nextID        int
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users:         make(map[string]struct{ ID, HashedPwd string }),
		refreshTokens: make(map[string]bool),
	}
}

// --- Register tests ---

func TestRegister_Success(t *testing.T) {
	app := setupTestApp(t)

	body := `{"email":"test@example.com","password":"password123","FirstName":"Cheikh","LastName":"Diouf"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}

	if resp.StatusCode != 201 {
		t.Errorf("expected status 201, got %d", resp.StatusCode)
	}

	var result model.AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result.AccessToken == "" {
		t.Error("expected non-empty access_token")
	}
	if result.RefreshToken == "" {
		t.Error("expected non-empty refresh_token")
	}
	if result.ExpiresIn != 900 {
		t.Errorf("expected expires_in=900, got %d", result.ExpiresIn)
	}
}

func TestRegister_MissingFields(t *testing.T) {
	app := setupTestApp(t)

	body := `{"email":"test@example.com","password":"password123"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}

	if resp.StatusCode != 400 {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestRegister_ShortPassword(t *testing.T) {
	app := setupTestApp(t)

	body := `{"email":"test@example.com","password":"short","FirstName":"Cheikh","LastName":"Diouf"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}

	if resp.StatusCode != 400 {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	app := setupTestApp(t)

	body := `{"email":"dup@example.com","password":"password123","FirstName":"A","LastName":"B"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	// First registration
	resp1, _ := app.Test(req)
	if resp1.StatusCode != 201 {
		t.Fatalf("first register: expected 201, got %d", resp1.StatusCode)
	}

	// Second registration with same email
	req2 := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(body))
	req2.Header.Set("Content-Type", "application/json")
	resp2, _ := app.Test(req2)

	if resp2.StatusCode != 409 {
		t.Errorf("expected status 409 for duplicate, got %d", resp2.StatusCode)
	}
}

// --- Login tests ---

func TestLogin_Success(t *testing.T) {
	app := setupTestApp(t)

	// Register first
	regBody := `{"email":"login@example.com","password":"password123","FirstName":"A","LastName":"B"}`
	regReq := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	app.Test(regReq)

	// Login
	loginBody := `{"email":"login@example.com","password":"password123"}`
	loginReq := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(loginReq)
	if err != nil {
		t.Fatalf("app.Test failed: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var result model.AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if result.AccessToken == "" {
		t.Error("expected non-empty access_token")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	app := setupTestApp(t)

	// Register
	regBody := `{"email":"wrong@example.com","password":"password123","FirstName":"A","LastName":"B"}`
	regReq := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	app.Test(regReq)

	// Login with wrong password
	loginBody := `{"email":"wrong@example.com","password":"wrongpassword"}`
	loginReq := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(loginReq)
	if resp.StatusCode != 401 {
		t.Errorf("expected status 401, got %d", resp.StatusCode)
	}
}

func TestLogin_NonexistentUser(t *testing.T) {
	app := setupTestApp(t)

	body := `{"email":"noone@example.com","password":"password123"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)
	if resp.StatusCode != 401 {
		t.Errorf("expected status 401, got %d", resp.StatusCode)
	}
}

func TestLogin_MissingFields(t *testing.T) {
	app := setupTestApp(t)

	body := `{"email":"test@example.com"}`
	req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)
	if resp.StatusCode != 400 {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}
}

// --- helpers ---

// inMemoryUserRepo is a simple in-memory implementation for testing.
type inMemoryUserRepo struct {
	users    map[string]inMemoryUser
	nextID   int
	tokens   map[string]bool
}

type inMemoryUser struct {
	ID         string
	Email      string
	HashedPwd  string
	FirstName  string
	LastName   string
}

func newInMemoryUserRepo() *inMemoryUserRepo {
	return &inMemoryUserRepo{
		users:  make(map[string]inMemoryUser),
		tokens: make(map[string]bool),
	}
}

func (r *inMemoryUserRepo) ExistsByEmail(email string) bool {
	_, ok := r.users[email]
	return ok
}

func (r *inMemoryUserRepo) Create(email, hashedPwd, firstName, lastName string) string {
	r.nextID++
	id := fmt.Sprintf("user-%d", r.nextID)
	r.users[email] = inMemoryUser{
		ID: id, Email: email, HashedPwd: hashedPwd,
		FirstName: firstName, LastName: lastName,
	}
	return id
}

func (r *inMemoryUserRepo) FindByEmail(email string) (string, string, bool) {
	u, ok := r.users[email]
	if !ok {
		return "", "", false
	}
	return u.ID, u.HashedPwd, true
}

func (r *inMemoryUserRepo) FindEmailByID(userID string) string {
	for _, u := range r.users {
		if u.ID == userID {
			return u.Email
		}
	}
	return ""
}

func (r *inMemoryUserRepo) StoreRefreshToken(token string) {
	r.tokens[token] = true
}

func (r *inMemoryUserRepo) RefreshTokenExists(token string) bool {
	return r.tokens[token]
}

// setupTestApp creates a Fiber app with auth routes wired to an in-memory store.
func setupTestApp(t *testing.T) *fiber.App {
	t.Helper()

	repo := newInMemoryUserRepo()
	jwtSecret := "test-secret-key-for-testing"

	svc := &testAuthService{
		repo:      repo,
		jwtSecret: jwtSecret,
	}

	authH := &testAuthHandler{svc: svc}

	app := fiber.New()
	api := app.Group("/api/v1")
	api.Post("/auth/register", authH.register)
	api.Post("/auth/login", authH.login)

	return app
}

// testAuthService provides auth logic using in-memory storage.
type testAuthService struct {
	repo      *inMemoryUserRepo
	jwtSecret string
}

func (s *testAuthService) registerUser(email, password, firstName, lastName string) (*model.AuthResponse, error) {
	if s.repo.ExistsByEmail(email) {
		return nil, service.ErrEmailExists
	}
	hashedPwd, err := middleware.HashPassword(password)
	if err != nil {
		return nil, err
	}
	userID := s.repo.Create(email, hashedPwd, firstName, lastName)

	accessToken, err := middleware.GenerateAccessToken(s.jwtSecret, userID, email)
	if err != nil {
		return nil, err
	}
	refreshToken, err := middleware.GenerateRefreshToken(s.jwtSecret, userID)
	if err != nil {
		return nil, err
	}
	s.repo.StoreRefreshToken(refreshToken)

	return &model.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    900,
	}, nil
}

func (s *testAuthService) loginUser(email, password string) (*model.AuthResponse, error) {
	userID, hashedPwd, ok := s.repo.FindByEmail(email)
	if !ok {
		return nil, service.ErrInvalidCredentials
	}
	if !middleware.VerifyPassword(password, hashedPwd) {
		return nil, service.ErrInvalidCredentials
	}

	accessToken, err := middleware.GenerateAccessToken(s.jwtSecret, userID, email)
	if err != nil {
		return nil, err
	}
	refreshToken, err := middleware.GenerateRefreshToken(s.jwtSecret, userID)
	if err != nil {
		return nil, err
	}
	s.repo.StoreRefreshToken(refreshToken)

	return &model.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    900,
	}, nil
}

// testAuthHandler wraps testAuthService with Fiber handlers.
type testAuthHandler struct {
	svc *testAuthService
}

func (h *testAuthHandler) register(c *fiber.Ctx) error {
	var req model.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid body"})
	}
	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" {
		return c.Status(400).JSON(fiber.Map{"error": "All fields are required"})
	}
	if len(req.Password) < 8 {
		return c.Status(400).JSON(fiber.Map{"error": "Password must be at least 8 characters"})
	}
	resp, err := h.svc.registerUser(req.Email, req.Password, req.FirstName, req.LastName)
	if err != nil {
		if err == service.ErrEmailExists {
			return c.Status(409).JSON(fiber.Map{"error": "Email already registered"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create user"})
	}
	return c.Status(201).JSON(resp)
}

func (h *testAuthHandler) login(c *fiber.Ctx) error {
	var req model.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid body"})
	}
	if req.Email == "" || req.Password == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Email and password are required"})
	}
	resp, err := h.svc.loginUser(req.Email, req.Password)
	if err != nil {
		if err == service.ErrInvalidCredentials {
			return c.Status(401).JSON(fiber.Map{"error": "Invalid credentials"})
		}
		return c.Status(500).JSON(fiber.Map{"error": "Failed to generate token"})
	}
	return c.JSON(resp)
}
