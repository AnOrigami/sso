package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"git.blauwelle.com/go/crate/log"
	"github.com/uptrace/bunrouter"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"git.blauwelle.com/go/crate/cmd/sso/middleware"
	"git.blauwelle.com/go/crate/cmd/sso/model"
	"git.blauwelle.com/go/crate/cmd/sso/response"
)

type CreateUserRequest LoginRequest

func (h *Handler) CreateUser() bunrouter.HandlerFunc {
	return func(rw http.ResponseWriter, r bunrouter.Request) error {
		ctx := r.Context()
		var request CreateUserRequest

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			log.Error(ctx, err.Error())
			return response.Error(rw, response.MessageBindError, bunrouter.H{})
		}
		var user model.User
		result := h.db.Where("username = ?", request.Username).First(&user)
		if result.RowsAffected > 0 {
			return response.Error(rw, response.MessageUserIsExist, bunrouter.H{})
		}

		passwordHash, _ := bcrypt.GenerateFromPassword([]byte(request.Password), 12)
		h.db.Create(&model.User{
			Username:     request.Username,
			PasswordHash: string(passwordHash),
		})
		return response.WriteOK(rw, response.MessageOK, bunrouter.H{})
	}
}

func (h *Handler) SearchUser() bunrouter.HandlerFunc {
	return func(rw http.ResponseWriter, r bunrouter.Request) error {
		pageSize := r.URL.Query().Get("pageSize")
		userName := r.URL.Query().Get("username")
		page := r.URL.Query().Get("page")
		// 设置默认每页记录数
		defaultPageSize := 20

		var users []model.User
		var count int64

		// 构建查询条件
		query := h.db.Model(&model.User{})

		// 添加模糊查询条件
		if userName != "" {
			query = query.Where("username LIKE ?", "%"+userName+"%")
		}

		// 查询总记录数
		if dbCount := query.Count(&count); dbCount.Error != nil {
			return response.Error(rw, response.MessageDatabaseConnectionError, bunrouter.H{})
		}

		// 将字符串类型的 pageSize 转换为整数类型
		pageSizeInt, err := strconv.Atoi(pageSize)
		pageInt, _ := strconv.Atoi(page)
		if err != nil || pageSizeInt <= 0 {
			// 如果转换失败或者 pageSize 小于等于 0，则使用默认每页记录数
			pageSizeInt = defaultPageSize
		}

		// 计算偏移量
		offset, err := calculateOffset(page, pageSizeInt, count)
		if err != nil {
			return response.Error(rw, response.MessageCalculateOffset, bunrouter.H{})
		}

		// 分页查询应用程序
		if dbFind := query.Offset(offset).Limit(pageSizeInt).Select("id, created_at, updated_at, username").Find(&users); dbFind.Error != nil {
			return response.Error(rw, response.MessageDatabaseConnectionError, bunrouter.H{})
		}
		return response.WriteOK(rw, response.MessageOK, response.NewPaginationData(pageInt, pageSizeInt, users))
	}
}

type UpdateUsernameRequest struct {
	Username string `json:"username"`
}

func (h *Handler) UpdateUsername() bunrouter.HandlerFunc {
	return func(rw http.ResponseWriter, r bunrouter.Request) error {
		ctx := r.Context()
		var request UpdateUsernameRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			log.Error(ctx, err.Error())
			return response.Error(rw, response.MessageBindError, bunrouter.H{})
		}
		_, ok, err := isExistUserByName(request.Username, h.db)
		if err != nil {
			return response.Error(rw, response.MessageDatabaseConnectionError, bunrouter.H{})
		}
		if ok {
			return response.Error(rw, response.MessageUserIsExist, bunrouter.H{})
		}
		id, err := strconv.Atoi(middleware.ContextJWTClaims{}.Value(r.Context()).Subject)
		if err != nil {
			return err
		}
		if result := h.db.Model(&model.User{}).Where("id = ?", id).
			Update("username", request.Username); result.Error != nil {
			return response.Error(rw, response.MessageDatabaseConnectionError, bunrouter.H{})
		}
		return response.WriteOK(rw, response.MessageOK, bunrouter.H{})
	}
}

type UpdatePDRequest struct {
	Password    string `json:"password"`
	NewPassword string `json:"newPassword"`
}

func (h *Handler) UpdatePassword() bunrouter.HandlerFunc {
	return func(rw http.ResponseWriter, r bunrouter.Request) error {
		ctx := r.Context()
		var request UpdatePDRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			log.Error(ctx, err.Error())
			return response.Error(rw, response.MessageBindError, bunrouter.H{})
		}
		id, err := strconv.Atoi(middleware.ContextJWTClaims{}.Value(r.Context()).Subject)
		if err != nil {
			return err
		}
		user, ok, err := isExistUserByID(uint(id), h.db)
		if err != nil {
			return response.Error(rw, response.MessageDatabaseConnectionError, bunrouter.H{})
		}
		if !ok {
			return response.Error(rw, response.MessageUserNotExist, bunrouter.H{})
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(request.Password)); err != nil {
			log.Error(ctx, err.Error())
			return response.Error(rw, response.MessageIncorrectPassword, bunrouter.H{})
		}
		passwordHash, err := bcrypt.GenerateFromPassword([]byte(request.NewPassword), 12)
		if err != nil {
			return err
		}
		if result := h.db.Model(&model.User{}).Where("id = ?", id).
			Update("password_hash", passwordHash); result.Error != nil {
			return response.Error(rw, response.MessageDatabaseConnectionError, bunrouter.H{})
		}
		return response.WriteOK(rw, response.MessageOK, bunrouter.H{})
	}
}

func isExistUserByID(UserID uint, db *gorm.DB) (model.User, bool, error) {
	var user model.User
	DB := db.Where("id = ?", UserID).Find(&user)
	if DB.Error != nil || DB.RowsAffected != 1 {
		return model.User{}, false, DB.Error
	}
	return user, true, nil
}
func isExistUserByName(Username string, db *gorm.DB) (model.User, bool, error) {
	var user model.User
	DB := db.Where("username = ?", Username).Find(&user)
	if DB.Error != nil || DB.RowsAffected != 1 {
		return model.User{}, false, DB.Error
	}
	return user, true, nil
}

type DeleteUserRequest struct {
	ID uint `json:"id"`
}

func (h *Handler) DeleteUser() bunrouter.HandlerFunc {
	return func(rw http.ResponseWriter, r bunrouter.Request) error {
		ctx := r.Context()
		var request DeleteUserRequest

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			log.Error(ctx, err.Error())
			return response.Error(rw, response.MessageBindError, bunrouter.H{})
		}

		if db := h.db.Delete(&model.User{}, "id=?", request.ID); db.Error != nil {
			return response.Error(rw, response.MessageDatabaseConnectionError, bunrouter.H{})
		}
		return response.WriteOK(rw, response.MessageOK, bunrouter.H{})
	}
}
