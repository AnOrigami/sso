package middleware

import (
	"context"
	"net/http"

	"git.blauwelle.com/go/crate/log"
	"github.com/golang-jwt/jwt/v5"
	"github.com/uptrace/bunrouter"

	"git.blauwelle.com/go/crate/cmd/sso/constants"
	"git.blauwelle.com/go/crate/cmd/sso/response"
	"git.blauwelle.com/go/crate/cmd/sso/util"
)

// 用于中间件和请求方法中claims的传递
type ContextKey[K, T any] struct{}

func (key ContextKey[K, T]) WithValue(ctx context.Context, value T) context.Context {
	return context.WithValue(ctx, key, value)
}

func (key ContextKey[K, T]) Value(ctx context.Context) T {
	return ctx.Value(key).(T)
}

type ContextJWTClaims struct {
	ContextKey[ContextJWTClaims, jwt.RegisteredClaims]
}

func HTTPMiddlewareJWT(jwtService *util.JWT) bunrouter.MiddlewareFunc {
	return func(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
		return func(rw http.ResponseWriter, r bunrouter.Request) error {
			ctx := r.Context()

			// 从请求的 Cookie 中获取 JWT
			cookie, err := r.Request.Cookie(constants.SessionCookieName)
			if err != nil {
				return response.Error(rw, response.MessageGetJWTError, bunrouter.H{})
			}

			// 验证 JWT 并获取声明信息
			claims, err := jwtService.Verify(ctx, cookie.Value)
			if err != nil {
				log.Error(ctx, err.Error())
				return response.Error(rw, response.MessageCheckJWTError, bunrouter.H{})
			}

			// 将声明信息存储到请求的上下文中
			ctx = ContextJWTClaims{}.WithValue(ctx, claims)
			r.Request = r.Request.WithContext(ctx)

			// 继续处理下一个中间件或请求处理函数
			return next(rw, r)
		}
	}
}
