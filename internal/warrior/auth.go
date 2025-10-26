package warrior

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var jwtSecret = []byte("your-secret-key-change-in-production")

// Claims represents JWT claims
type Claims struct {
	WarriorID uint   `json:"warrior_id"`
	Username  string `json:"username"`
	Role      Role   `json:"role"`
	jwt.RegisteredClaims
}

// hashPassword hashes a password using bcrypt
func hashPassword(password string) string {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return string(hashed)
}

// comparePassword compares a password with a hash
func comparePassword(hashedPassword string, password []byte) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), password)
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token   string `json:"token"`
	Warrior Warrior `json:"warrior"`
}

// Login authenticates a warrior and returns a JWT token
func Login(loginReq LoginRequest) (*LoginResponse, error) {
	var warrior Warrior
	if err := DB.Where("username = ? OR email = ?", loginReq.Username, loginReq.Username).First(&warrior).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid credentials")
		}
		return nil, err
	}

	if err := comparePassword(warrior.Password, []byte(loginReq.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	token, err := generateToken(&warrior)
	if err != nil {
		return nil, err
	}

	// Remove password from response
	warrior.Password = ""

	return &LoginResponse{
		Token:   token,
		Warrior: warrior,
	}, nil
}

// generateToken generates a JWT token for a warrior
func generateToken(warrior *Warrior) (string, error) {
	claims := Claims{
		WarriorID: warrior.ID,
		Username:  warrior.Username,
		Role:      warrior.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ValidateToken validates a JWT token and returns the claims
func ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
