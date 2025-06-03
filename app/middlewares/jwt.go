package middleware

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// JWT configuration from environment variables
var (
	jwtSecret []byte
	tokenTTL  time.Duration
)

// Initialize JWT configuration from environment variables
func init() {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}
	jwtSecret = []byte(secret)

	// Optional: Configure token TTL from environment
	ttlHours := os.Getenv("JWT_TTL_HOURS")
	if ttlHours == "" {
		tokenTTL = 24 * time.Hour // Default 24 hours
	} else {
		// Parse custom TTL if needed
		tokenTTL = 24 * time.Hour
	}
}

// Claims represents the JWT claims structure
type Claims struct {
	ClientID uint   `json:"client_id"`
	UserID   uint   `json:"user_id,omitempty"`
	Role     string `json:"role,omitempty"`
	Service  string `json:"service,omitempty"`
	jwt.RegisteredClaims
}

//GenerateJWT creates a new JWT token (for auth service)
//func GenerateJWT(clientID uint, userID uint, role string) (string, error) {
//	claims := &Claims{
//		ClientID: clientID,
//		//UserID:   userID,
//		//Role:     role,
//		RegisteredClaims: jwt.RegisteredClaims{
//			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenTTL)),
//			IssuedAt:  jwt.NewNumericDate(time.Now()),
//			NotBefore: jwt.NewNumericDate(time.Now()),
//			//Issuer:    os.Getenv("SERVICE_NAME"), // Optional: identify issuing service
//		},
//	}
//
//	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
//	return token.SignedString(jwtSecret)
//}

func GenerateJWT(clientID uint) (string, error) {
	claims := &Claims{
		ClientID: clientID,
		//UserID:   userID,
		//Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			//Issuer:    os.Getenv("SERVICE_NAME"), // Optional: identify issuing service
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ValidateToken validates a JWT token and returns claims
func ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("token parsing error: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Additional validation
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("token has expired")
	}

	return claims, nil
}

// JWTMiddleware middleware for protecting routes
func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error": "Missing Authorization header"}`, http.StatusUnauthorized)
			return
		}

		// Extract token from "Bearer token" format
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenStr == authHeader {
			http.Error(w, `{"error": "Invalid Authorization header format. Use: Bearer <token>"}`, http.StatusUnauthorized)
			return
		}

		// Validate token
		claims, err := ValidateToken(tokenStr)
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "Invalid token: %s"}`, err.Error()), http.StatusUnauthorized)
			return
		}

		// Add claims to request context
		ctx := context.WithValue(r.Context(), "jwt_claims", claims)
		ctx = context.WithValue(ctx, "client_id", claims.ClientID)
		ctx = context.WithValue(ctx, "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "user_role", claims.Role)

		// Call next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalJWTMiddleware - for routes that can work with or without auth
func OptionalJWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authHeader != "" {
			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenStr != authHeader {
				if claims, err := ValidateToken(tokenStr); err == nil {
					ctx := context.WithValue(r.Context(), "jwt_claims", claims)
					ctx = context.WithValue(ctx, "client_id", claims.ClientID)
					r = r.WithContext(ctx)
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

// RequireRole middleware that requires specific role
func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return JWTMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := GetClaimsFromContext(r.Context())
			if claims == nil {
				http.Error(w, `{"error": "No authentication claims found"}`, http.StatusInternalServerError)
				return
			}

			if claims.Role != role {
				http.Error(w, `{"error": "Insufficient permissions"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		}))
	}
}

// Helper functions to extract data from context

// GetClaimsFromContext extracts JWT claims from request context
func GetClaimsFromContext(ctx context.Context) *Claims {
	if claims, ok := ctx.Value("jwt_claims").(*Claims); ok {
		return claims
	}
	return nil
}

// GetClientIDFromContext extracts client ID from request context
func GetClientIDFromContext(ctx context.Context) (uint, bool) {
	if clientID, ok := ctx.Value("client_id").(uint); ok {
		return clientID, true
	}
	return 0, false
}

// GetUserIDFromContext extracts user ID from request context
func GetUserIDFromContext(ctx context.Context) (uint, bool) {
	if userID, ok := ctx.Value("user_id").(uint); ok {
		return userID, true
	}
	return 0, false
}

// GetUserRoleFromContext extracts user role from request context
func GetUserRoleFromContext(ctx context.Context) (string, bool) {
	if role, ok := ctx.Value("user_role").(string); ok {
		return role, true
	}
	return "", false
}

// GenerateServiceToken creates tokens for service-to-service communication
func GenerateServiceToken(serviceName string) (string, error) {
	claims := &Claims{
		ClientID: 0, // Service accounts use 0
		Role:     "service",
		Service:  serviceName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)), // Shorter TTL for service tokens
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    os.Getenv("SERVICE_NAME"),
			Subject:   "service-account",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}
