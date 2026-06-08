package services

import (
	"errors"
	"os"
	"time"

	"payment-gateway/internal/repositories"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	adminRepo *repositories.AdminRepository
}

func NewAuthService() *AuthService {
	return &AuthService{
		adminRepo: repositories.NewAdminRepository(),
	}
}

func (s *AuthService) Login(username, password string) (string, error) {
	admin, err := s.adminRepo.FindByUsername(username)
	if err != nil {
		return "", errors.New("username atau password salah")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password)); err != nil {
		return "", errors.New("username atau password salah")
	}

	// Generate JWT Token (Expired 24 Jam)
	claims := jwt.MapClaims{
		"admin_id": admin.ID,
		"username": admin.Username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", errors.New("gagal generate token")
	}

	return t, nil
}
