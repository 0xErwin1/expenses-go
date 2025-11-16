package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"

	"github.com/iperez/new-expenses-go/internal/domain/models"
	"github.com/iperez/new-expenses-go/pkg/apperror"
)

// AuthService authenticates users and manages session state in Redis.
type AuthService struct {
	users *UserService
	redis *redis.Client
	ttl   time.Duration
}

// NewAuthService builds a new AuthService instance.
func NewAuthService(users *UserService, redis *redis.Client, ttl time.Duration) *AuthService {
	return &AuthService{users: users, redis: redis, ttl: ttl}
}

// Login validates credentials and stores the user identifier in Redis using the provided session key.
func (s *AuthService) Login(ctx context.Context, email, password, sessionID string) (*models.User, string, error) {
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return nil, "", err
	}

	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)) != nil {
		return nil, "", apperror.New(apperror.AuthBadAuth, nil)
	}

	sessionKey := sessionID
	if sessionKey == "" {
		sessionKey = uuid.NewString()
	}

	redisKey := s.userSessionKey(sessionKey)
	if err := s.redis.Set(ctx, redisKey, user.UserID, s.ttl).Err(); err != nil {
		return nil, "", err
	}

	return user, sessionKey, nil
}

// Logout removes the redis entry connected with the incoming session.
func (s *AuthService) Logout(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return nil
	}

	if err := s.redis.Del(ctx, s.userSessionKey(sessionID)).Err(); err != nil && err != redis.Nil {
		return err
	}

	return nil
}

// ResolveSession returns the user identifier associated with the session if it exists.
func (s *AuthService) ResolveSession(ctx context.Context, sessionID string) (string, error) {
	if sessionID == "" {
		return "", nil
	}

	userID, err := s.redis.Get(ctx, s.userSessionKey(sessionID)).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil
		}

		return "", err
	}

	return userID, nil
}

func (s *AuthService) userSessionKey(sessionID string) string {
	return fmt.Sprintf("user:%s", sessionID)
}
