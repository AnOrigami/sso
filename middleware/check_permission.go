package middleware

import (
	"net/http"
	"strconv"

	"github.com/uptrace/bunrouter"
	"gorm.io/gorm"

	"git.blauwelle.com/go/crate/cmd/sso/constants"
	"git.blauwelle.com/go/crate/cmd/sso/model"
	"git.blauwelle.com/go/crate/cmd/sso/response"
)

func CheckPermission(db *gorm.DB) bunrouter.MiddlewareFunc {
	return func(next bunrouter.HandlerFunc) bunrouter.HandlerFunc {
		return func(rw http.ResponseWriter, r bunrouter.Request) error {
			id, err := strconv.Atoi(ContextJWTClaims{}.Value(r.Context()).Subject)
			if err != nil {
				return err
			}
			var roles []model.Role
			db.Joins("JOIN user_roles ON user_roles.role_id = roles.id").
				Joins("JOIN users ON users.id = user_roles.user_id").
				Where("users.id = ?", id).
				Find(&roles)
			if len(roles) == 0 {
				return response.Error(rw, response.MessageUnauthorized, bunrouter.H{})
			}
			var isAdmin bool
			// 可以遍历 roles 切片获取每个角色的名称
			for _, role := range roles {
				if role.Name == constants.Admin {
					isAdmin = true
					break
				}
			}
			if isAdmin {
				return next(rw, r)
			}
			return response.Error(rw, response.MessageUnauthorized, bunrouter.H{})
		}
	}
}
