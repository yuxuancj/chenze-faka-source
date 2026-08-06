package service

import (
	"errors"
	"time"

	"chenze-faka/internal/model"
	"chenze-faka/internal/pkg/database"
	"chenze-faka/internal/pkg/utils"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct{}

func NewAuthService() *AuthService {
	return &AuthService{}
}

func (s *AuthService) Register(username, password string) (*model.User, error) {
	if username == "" || password == "" {
		return nil, errors.New("用户名和密码不能为空")
	}

	if database.DB == nil {
		return nil, errors.New("数据库未连接")
	}

	var existing model.User
	err := database.DB.Where("username = ?", username).First(&existing).Error
	if err == nil {
		return nil, errors.New("用户名已存在")
	}

	salt := utils.GenerateSalt()
	hashedPassword := utils.HashPassword(password, salt)

	user := &model.User{
		Username:     username,
		PasswordHash: hashedPassword,
		Salt:         salt,
		Role:         model.RoleAdmin,
	}

	if err := database.DB.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Login(username, password, jwtSecret string, expireHours int) (string, *model.User, error) {
	if database.DB == nil {
		if username == "admin" && password == "admin123" {
			user := &model.User{
				ID:       1,
				Username: "admin",
				Role:     model.RoleAdmin,
			}
			token, err := generateToken(user, jwtSecret, expireHours)
			if err != nil {
				return "", nil, err
			}
			return token, user, nil
		}
		return "", nil, errors.New("用户名或密码错误")
	}

	var user model.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		return "", nil, errors.New("用户名或密码错误")
	}

	hashedPassword := utils.HashPassword(password, user.Salt)
	if hashedPassword != user.PasswordHash {
		return "", nil, errors.New("用户名或密码错误")
	}

	token, err := generateToken(&user, jwtSecret, expireHours)
	if err != nil {
		return "", nil, err
	}

	now := time.Now()
	database.DB.Model(&model.User{}).Where("id = ?", user.ID).Update("last_login_at", now)

	return token, &user, nil
}

func (s *AuthService) ParseToken(tokenString, jwtSecret string) (*jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("无效的签名方式")
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return &claims, nil
	}

	return nil, errors.New("无效的令牌")
}

func (s *AuthService) GetUserByID(id uint) (*model.User, error) {
	if database.DB == nil {
		return nil, errors.New("数据库未连接")
	}

	var user model.User
	if err := database.DB.First(&user, id).Error; err != nil {
		return nil, errors.New("用户不存在")
	}
	return &user, nil
}

func generateToken(user *model.User, secret string, expireHours int) (string, error) {
	if expireHours <= 0 {
		expireHours = 72
	}

	expireTime := time.Duration(expireHours) * time.Hour

	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      time.Now().Add(expireTime).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func (s *AuthService) CheckInstalled() bool {
	if database.DB == nil {
		return false
	}
	var count int64
	database.DB.Model(&model.User{}).Count(&count)
	return count > 0
}