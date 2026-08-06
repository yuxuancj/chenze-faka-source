package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"chenze-faka/internal/config"
	"chenze-faka/internal/model"
	"chenze-faka/internal/repository"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	userRepo   *repository.UserRepository
	jwtConfig  config.JWTConfig
}

func NewAuthService(jwtConfig config.JWTConfig) *AuthService {
	return &AuthService{
		userRepo:  repository.NewUserRepository(),
		jwtConfig: jwtConfig,
	}
}

func (s *AuthService) Register(username, password string) (*model.User, error) {
	if username == "" || password == "" {
		return nil, errors.New("username and password cannot be empty")
	}

	existingUser, err := s.userRepo.GetByUsername(username)
	if err == nil && existingUser != nil {
		return nil, errors.New("username already exists")
	}

	salt := generateSalt()
	hashedPassword := hashPassword(password, salt)

	user := &model.User{
		Username: username,
		Password: hashedPassword,
		Salt:     salt,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Login(username, password string) (string, error) {
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		return "", errors.New("invalid username or password")
	}

	hashedPassword := hashPassword(password, user.Salt)
	if hashedPassword != user.Password {
		return "", errors.New("invalid username or password")
	}

	token, err := s.generateToken(user)
	if err != nil {
		return "", err
	}

	s.userRepo.UpdateLastLogin(user.ID)
	return token, nil
}

func (s *AuthService) ParseToken(tokenString string) (*jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return []byte(s.jwtConfig.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return &claims, nil
	}

	return nil, errors.New("invalid token")
}

func (s *AuthService) GetUserByID(id uint) (*model.User, error) {
	return s.userRepo.GetByID(id)
}

func (s *AuthService) generateToken(user *model.User) (string, error) {
	expireTime := time.Duration(s.jwtConfig.ExpireTime) * time.Hour

	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"exp":      time.Now().Add(expireTime).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtConfig.Secret))
}

func generateSalt() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func hashPassword(password, salt string) string {
	hash := sha256.New()
	hash.Write([]byte(password + salt))
	return hex.EncodeToString(hash.Sum(nil))
}
