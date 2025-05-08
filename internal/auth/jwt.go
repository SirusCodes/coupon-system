package auth

import (
    "fmt"
    "time"

    "github.com/golang-jwt/jwt/v5"
)

// Define your secret key (use a secure method to manage this in production)
var jwtSecret = []byte("your_super_secret_key") // TODO: Replace with a secure key

// Claims defines the custom claims for the JWT.
// Claims defines the custom claims for the JWT.
type Claims struct {
    UserID string `json:"user_id"`
    Role   string `json:"role"` // e.g., "admin", "user"
    jwt.RegisteredClaims
}

// GenerateJWT generates a new JWT for the given user ID and role.
func GenerateJWT(userID, role string) (string, error) {
    expirationTime := time.Now().Add(24 * time.Hour) // Token expires in 24 hours
    claims := &Claims{
        UserID: userID,
        Role:   role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(expirationTime),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            NotBefore: jwt.NewNumericDate(time.Now()),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(jwtSecret)
}

// ParseJWT parses and validates a JWT, returning the claims if valid.
func ParseJWT(tokenString string) (*Claims, error) {
    claims := &Claims{}
    _, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return jwtSecret, nil
    })
    return claims, err
}