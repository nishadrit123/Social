package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"social/internal/store"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func (app *application) getJWTFromHeader(w http.ResponseWriter, r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		app.unauthorizedErrorResponse(w, r, fmt.Errorf("authorization header is missing"))
		return ""
	}
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		app.unauthorizedErrorResponse(w, r, fmt.Errorf("authorization header is malformed"))
		return ""
	}
	token := parts[1]
	return token
}

// Authentication
func (app *application) AuthTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := app.getJWTFromHeader(w, r)
		jwtToken, err := app.authenticator.ValidateToken(token)
		if err != nil {
			app.unauthorizedErrorResponse(w, r, err)
			return
		}

		claims, _ := jwtToken.Claims.(jwt.MapClaims)

		userID, err := strconv.ParseInt(fmt.Sprintf("%.f", claims["sub"]), 10, 64)
		if err != nil {
			app.unauthorizedErrorResponse(w, r, err)
			return
		}

		ctx := r.Context()

		isLoggedIn, _ := app.cacheStorage.Users.Get(ctx, userID, token, "login")
		if !isLoggedIn.(bool) {
			err = errors.New("invalid token")
			app.unauthorizedErrorResponse(w, r, err)
			return
		}

		user, err := app.getUserFromRedisCache(ctx, userID)
		if err != nil {
			app.unauthorizedErrorResponse(w, r, err)
			return
		}

		ctx = context.WithValue(ctx, userCtx, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Authorization
func (app *application) checkOwnership(requiredRole, intender string, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := getUserFromContext(r)

		if intender == "post" {
			post := getPostFromCtx(r)
			if post.UserID == user.ID {
				next.ServeHTTP(w, r)
				return
			}
		} else if intender == "comment" {
			comment := getCommentFromCtx(r)
			if comment.UserID == user.ID {
				next.ServeHTTP(w, r)
				return
			}
		} else if intender == "story" {
			StoryUserID := getStoryUserIDFromCtx(r)
			if StoryUserID == user.ID {
				next.ServeHTTP(w, r)
				return
			}
		}

		allowed, err := app.checkRolePrecedence(r.Context(), user, requiredRole)
		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

		if !allowed {
			app.forbiddenResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) checkRolePrecedence(ctx context.Context, user *store.User, roleName string) (bool, error) {
	role, err := app.store.Roles.GetByName(ctx, roleName)
	if err != nil {
		return false, err
	}

	return user.Role.Level >= role.Level, nil
}

func (app *application) getUserFromRedisCache(ctx context.Context, userID int64) (*store.User, error) {
	user, err := app.cacheStorage.Users.Get(ctx, userID, "", "user")
	if err != nil {
		return nil, err
	}
	if user == nil || user.(*store.User).ID == 0 {
		user, err = app.store.Users.GetByID(ctx, userID)
		if err != nil {
			return nil, err
		}

		if err := app.cacheStorage.Users.Set(ctx, user, userID, "user"); err != nil {
			return nil, err
		}
	}
	return user.(*store.User), nil
}
