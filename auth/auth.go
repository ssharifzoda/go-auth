package auth

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// AuthService - главный сервис аутентификации.
type AuthService struct {
	storage   Storage
	secretKey []byte
	tokenTTL  time.Duration
}

// NewAuthService создает новый экземпляр AuthService.
func NewAuthService(storage Storage, secretKey []byte, ttl time.Duration) *AuthService {
	return &AuthService{
		storage:   storage,
		secretKey: secretKey,
		tokenTTL:  ttl,
	}
}

// Login (Логин)
// Проверяет учетные данные и генерирует JWT-токен.
func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.storage.GetUserByEmail(ctx, email)
	if err != nil {
		// Обычно возвращают универсальную ошибку для безопасности
		return "", errors.New("неверные учетные данные")
	}

	// Сравнение хэша пароля
	err = bcrypt.CompareHashAndPassword([]byte(user.GetPasswordHash()), []byte(password))
	if err != nil {
		return "", errors.New("неверные учетные данные")
	}

	// Генерация JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, JWTClaims{
		UserID: user.GetID(),
		Email:  user.GetEmail(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	})

	tokenString, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", errors.New("ошибка подписи токена")
	}

	return tokenString, nil
}

// ParseAndValidateToken (Проверка JWT)
// Парсит токен, проверяет подпись и срок действия.
func (s *AuthService) ParseAndValidateToken(tokenString string) (*JWTClaims, error) {
	claims := &JWTClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Убеждаемся, что используется ожидаемый алгоритм
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("неожиданный метод подписи")
		}
		return s.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("токен недействителен")
	}

	// Возвращаем полезную нагрузку
	return claims, nil
}

// Logout (Логаут)
// В stateless JWT логаут означает удаление токена клиентом.
// Этот метод можно оставить для будущего функционала (например, черный список токенов).
func (s *AuthService) Logout(ctx context.Context, tokenString string) error {
	// В stateless JWT это NO-OP (не требует действий)
	return nil
}
