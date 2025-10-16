# `go-auth-pkg`

**Универсальный пакет для аутентификации на основе JWT (Golang)**

Этот пакет предоставляет ядро бизнес-логики для аутентификации (логин, проверка пароля, генерация и валидация JWT-токенов). Он спроектирован с использованием интерфейсов, что позволяет легко интегрировать его с любой базой данных (PostgreSQL, MySQL, Redis и т.д.) или ORM, используемыми в вашем сервисе.

## 🚀 Установка

Для добавления пакета в ваш Go-проект используйте команду:

```bash
go get github.com/ssharifzoda/go-auth

```

Также убедитесь, что в вашем проекте установлены необходимые зависимости:

```bash
go get golang.org/x/crypto/bcrypt
go get github.com/golang-jwt/jwt/v5
```

## 🛠️ Инструкция по Использованию

Для интеграции пакета в ваш сервис необходимо выполнить три основных шага:

### Шаг 1. Реализация Интерфейсов Моделей

Создайте ваши структуры пользователя (`User`) в вашем проекте, которые должны реализовать интерфейс **`auth.UserIn`** из нашего пакета.

```go
// my-service/models/user.go
package models

import "go-auth-pkg/auth"

// ServiceUser - Ваша конкретная структура с вашими полями.
type ServiceUser struct {
	ID        int64
	Email     string
	Password  string // Хэш пароля
	FirstName string // Дополнительное поле
}

// Реализация auth.Userer (обязательно!)
func (u *ServiceUser) GetID() int64          { return u.ID }
func (u *ServiceUser) GetEmail() string      { return u.Email }
func (u *ServiceUser) GetPasswordHash() string { return u.Password }

// Убедитесь, что импорт 'go-auth-pkg/auth' правильный.
var _ auth.UserIn = (*ServiceUser)(nil) // Проверка интерфейса во время компиляции
```

-----

### Шаг 2. Реализация Интерфейса Хранилища (Storage)

Создайте структуру, которая будет взаимодействовать с вашей базой данных (PostgreSQL) и реализовывать интерфейс **`auth.Storage`**.

**Важно:** Методы `Storage` должны возвращать ваши модели, **приведенные к интерфейсу** `auth.UserIn`.

```go
// my-service/storage/pg_storage.go
package storage

import (
	"context"
	"database/sql"
	"errors"
	"my-service/models"
	
	"go-auth-pkg/auth"
)

type PgStorage struct {
	DB *sql.DB
}

// GetUserByEmail реализует auth.Storage
func (s *PgStorage) GetUserByEmail(ctx context.Context, email string) (auth.UserIn, error) {
	var u models.ServiceUser
	// Здесь код для запроса к Postgres и сканирования в u
	// Пример (упрощенно):
	row := s.DB.QueryRowContext(ctx, "SELECT id, email, password FROM users WHERE email = $1", email)
	if err := row.Scan(&u.ID, &u.Email, &u.Password); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	// Возвращаем вашу конкретную модель, приведенную к нашему интерфейсу
	return &u, nil
}

// GetUserByID реализует auth.Storage
func (s *PgStorage) GetUserByID(ctx context.Context, id int64) (auth.UserIn, error) {
    // Аналогичная логика, но поиск по ID
	return nil, errors.New("not implemented")
}

// Убедитесь, что импорт 'go-auth-pkg/auth' правильный.
var _ auth.Storage = (*PgStorage)(nil) // Проверка интерфейса во время компиляции
```

-----

### Шаг 3. Использование AuthService

В вашей точке входа (например, `main.go`) инициализируйте наш сервис и используйте его методы.

```go
// my-service/main.go
package main

import (
	"context"
	"fmt"
	"time"
	"database/sql"
    _ "github.com/lib/pq" // Драйвер Postgres
    
	"go-auth-pkg/auth"
	"my-service/storage" // Ваш файл с PgStorage
)

func main() {
    // 1. Настройка и подключение к БД (ваш код)
    db, err := sql.Open("postgres", "user=... password=...")
    if err != nil {
        panic(err)
    }

    // 2. Инициализация вашего хранилища
	pgStorage := &storage.PgStorage{DB: db} 

    // 3. Инициализация AuthService из нашего пакета
	secretKey := []byte("ВАШ_СЕКРЕТНЫЙ_КЛЮЧ_ДЛЯ_JWT") 
	tokenTTL := 12 * time.Hour
	
	authService := auth.NewAuthService(pgStorage, secretKey, tokenTTL)

	// --- Использование ---

	// Логин
	token, err := authService.Login(context.Background(), "user@example.com", "mypassword")
	if err != nil {
		fmt.Printf("Ошибка логина: %v\n", err)
		return
	}
	fmt.Printf("Получен токен: %s\n", token)
    
	// Валидация токена
	claims, err := authService.ParseAndValidateToken(token)
	if err != nil {
		fmt.Printf("Ошибка валидации: %v\n", err)
		return
	}
	fmt.Printf("Токен валиден для UserID: %d\n", claims.UserID)
}
```

-----

## 🧩 Расширение Функционала

Если вашему сервису нужны дополнительные методы (например, регистрация, сброс пароля, обновление пользователя), **не нужно менять пакет `go-auth-pkg`**.

### Добавление новых методов

Просто добавьте необходимые методы к **вашей реализации** хранилища (`PgStorage`):

```go
// my-service/storage/pg_storage.go

// CreateUser - Метод, необходимый для регистрации, которого нет в core-интерфейсе auth.Storage
func (s *PgStorage) CreateUser(ctx context.Context, email, passwordHash, name string) (*models.ServiceUser, error) {
    // Логика INSERT в базу данных
    // ...
    return nil, nil
}

// UpdatePassword - Метод для сброса пароля
func (s *PgStorage) UpdatePassword(ctx context.Context, userID int64, newHash string) error {
    // Логика UPDATE в базу данных
    // ...
    return nil
}

// ... и любые другие методы, нужные вашему сервису.
```