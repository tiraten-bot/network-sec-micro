package arena

import (
    "errors"
    "strconv"
    "strings"

    "network-sec-micro/pkg/auth"

    "github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT token and sets user in context
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        ah := c.GetHeader("Authorization")
        if ah == "" { c.JSON(401, gin.H{"error": "authorization header required"}); c.Abort(); return }
        parts := strings.Split(ah, " ")
        if len(parts) != 2 || parts[0] != "Bearer" { c.JSON(401, gin.H{"error": "invalid authorization header format"}); c.Abort(); return }
        claims, err := auth.ValidateToken(parts[1])
        if err != nil { c.JSON(401, gin.H{"error": "invalid token"}); c.Abort(); return }
        c.Set("username", claims.Username)
        c.Set("user_id", strconv.FormatUint(uint64(claims.UserID), 10))
        c.Set("role", claims.Role)
        c.Next()
    }
}

type User struct {
    UserID   uint
    Username string
    Role     string
}

func GetCurrentUser(c *gin.Context) (*User, error) {
    username := c.GetString("username")
    if username == "" { return nil, errors.New("username not found") }
    idStr := c.GetString("user_id")
    var idUint uint
    if idStr != "" {
        if v, err := strconv.ParseUint(idStr, 10, 32); err == nil { idUint = uint(v) }
    }
    return &User{UserID: idUint, Username: username, Role: c.GetString("role")}, nil
}


