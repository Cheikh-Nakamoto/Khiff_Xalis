package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/brvm/go-api/internal/middleware"
	"github.com/brvm/go-api/internal/model"
	"github.com/brvm/go-api/internal/repository"
	"github.com/golang-jwt/jwt/v5"
)

// AuthService contains the business logic for authentication.
type AuthService struct {
	Users     *repository.UserRepo
	JWTSecret string
}

// Register creates a new user and returns auth tokens.
func (s *AuthService) Register(ctx context.Context, req model.RegisterRequest) (*model.AuthResponse, error) {
	// Check if email exists
	exists, err := s.Users.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("db error: %w", err)
	}
	if exists {
		return nil, ErrEmailExists
	}

	// Hash password with Argon2
	hashedPwd, err := middleware.HashPassword(req.Password)
	if err != nil {
		log.Printf("Password hash error: %v", err)
		return nil, ErrInternal
	}

	// Insert user
	userID, err := s.Users.Create(ctx, req.Email, hashedPwd, req.FirstName, req.LastName)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate tokens
	accessToken, err := middleware.GenerateAccessToken(s.JWTSecret, userID, req.Email)
	if err != nil {
		return nil, ErrTokenGen
	}
	refreshToken, err := middleware.GenerateRefreshToken(s.JWTSecret, userID)
	if err != nil {
		return nil, ErrTokenGen
	}

	// Store refresh token
	if err := s.Users.StoreRefreshToken(ctx, userID, refreshToken, time.Now().Add(7*24*time.Hour)); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &model.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    900,
	}, nil
}

// Login authenticates a user and returns auth tokens.
func (s *AuthService) Login(ctx context.Context, req model.LoginRequest) (*model.AuthResponse, error) {
	// Find user
	userID, hashedPwd, err := s.Users.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Verify password
	if !middleware.VerifyPassword(req.Password, hashedPwd) {
		return nil, ErrInvalidCredentials
	}

	// Generate tokens
	accessToken, err := middleware.GenerateAccessToken(s.JWTSecret, userID, req.Email)
	if err != nil {
		return nil, ErrTokenGen
	}
	refreshToken, err := middleware.GenerateRefreshToken(s.JWTSecret, userID)
	if err != nil {
		return nil, ErrTokenGen
	}

	// Store refresh token
	if err := s.Users.StoreRefreshToken(ctx, userID, refreshToken, time.Now().Add(7*24*time.Hour)); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &model.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    900,
	}, nil
}

// Refresh validates a refresh token and returns a new access token.
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (string, error) {
	// Validate JWT structure
	token, err := jwt.Parse(refreshToken, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return "", ErrInvalidRefreshToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", ErrInvalidRefreshToken
	}

	userID, _ := claims["sub"].(string)
	if userID == "" {
		return "", ErrInvalidRefreshToken
	}

	// Check token exists in DB
	exists, err := s.Users.RefreshTokenExists(ctx, refreshToken)
	if err != nil || !exists {
		return "", ErrInvalidRefreshToken
	}

	// Get user email
	email, err := s.Users.FindEmailByID(ctx, userID)
	if err != nil {
		return "", ErrInvalidRefreshToken
	}

	// Generate new access token
	accessToken, err := middleware.GenerateAccessToken(s.JWTSecret, userID, email)
	if err != nil {
		return "", ErrTokenGen
	}

	return accessToken, nil
}

// Sentinel errors for auth flows.
var (
	ErrEmailExists         = fmt.Errorf("email already registered")
	ErrInvalidCredentials  = fmt.Errorf("invalid credentials")
	ErrTokenGen            = fmt.Errorf("failed to generate token")
	ErrInvalidRefreshToken = fmt.Errorf("invalid refresh token")
	ErrInternal            = fmt.Errorf("internal error")
)
