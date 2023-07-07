package handler

import (
	"encoding/json"
	"net/http"

	"git.blauwelle.com/go/crate/log"
	"github.com/uptrace/bunrouter"

	"git.blauwelle.com/go/crate/cmd/sso/constants"
	"git.blauwelle.com/go/crate/cmd/sso/model"
	"git.blauwelle.com/go/crate/cmd/sso/response"
)

type CreateAdminUserID struct {
	ID uint `json:"id"`
}
type ConcelAdminUserID CreateAdminUserID

func (h *Handler) CreateAdmin() bunrouter.HandlerFunc {
	return func(rw http.ResponseWriter, r bunrouter.Request) error {
		ctx := r.Context()
		var adminUserID CreateAdminUserID

		if err := json.NewDecoder(r.Body).Decode(&adminUserID); err != nil {
			log.Error(ctx, err.Error())
			return response.Error(rw, response.MessageBindError, bunrouter.H{})
		}

		_, ok, err := isExistUserByID(adminUserID.ID, h.db)
		if err != nil {
			return response.Error(rw, response.MessageDatabaseConnectionError, bunrouter.H{})
		}
		if !ok {
			return response.Error(rw, response.MessageUserNotExist, bunrouter.H{})
		}

		userRole := model.UserRole{
			UserID: uint(adminUserID.ID),
			RoleID: constants.AdminID,
		}

		if db := h.db.Create(&userRole); db.Error != nil {
			return response.Error(rw, response.MessageDatabaseConnectionError, bunrouter.H{})
		}
		return response.WriteOK(rw, response.MessageOK, bunrouter.H{})
	}
}

func (h *Handler) ConcelAdmin() bunrouter.HandlerFunc {
	return func(rw http.ResponseWriter, r bunrouter.Request) error {
		ctx := r.Context()
		var adminUserID ConcelAdminUserID

		if err := json.NewDecoder(r.Body).Decode(&adminUserID); err != nil {
			log.Error(ctx, err.Error())
			return response.Error(rw, response.MessageBindError, bunrouter.H{})
		}
		_, ok, err := isExistUserByID(adminUserID.ID, h.db)
		if err != nil {
			return response.Error(rw, response.MessageDatabaseConnectionError, bunrouter.H{})
		}
		if !ok {
			return response.Error(rw, response.MessageUserNotExist, bunrouter.H{})
		}
		if db := h.db.Delete(&model.UserRole{}, "user_id=?", adminUserID.ID); db.Error != nil {
			return response.Error(rw, response.MessageDatabaseConnectionError, bunrouter.H{})
		}
		return response.WriteOK(rw, response.MessageOK, bunrouter.H{})
	}
}
