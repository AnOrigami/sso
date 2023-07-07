package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"git.blauwelle.com/go/crate/log"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/uptrace/bunrouter"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"git.blauwelle.com/go/crate/cmd/sso/constants"
	"git.blauwelle.com/go/crate/cmd/sso/middleware"
	"git.blauwelle.com/go/crate/cmd/sso/model"
	"git.blauwelle.com/go/crate/cmd/sso/response"
	"git.blauwelle.com/go/crate/cmd/sso/util"
)

type Handler struct {
	db      *gorm.DB
	redisDB *redis.Client
	r       util.StringRand
	j       *util.JWT
}

func NewHandler(db *gorm.DB, redisDB *redis.Client, jwtService *util.JWT) *Handler {
	return &Handler{
		db:      db,
		redisDB: redisDB,
		r:       util.NewStringRand(),
		j:       jwtService,
	}
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (h *Handler) Login() bunrouter.HandlerFunc {
	return func(rw http.ResponseWriter, r bunrouter.Request) error {
		ctx := r.Context()
		var request LoginRequest

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			log.Error(ctx, err.Error())
			return response.Error(rw, response.MessageBindError, bunrouter.H{})
		}
		user, ok, err := isExistUserByName(request.Username, h.db)
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

		exp := time.Now().AddDate(1, 0, 0)
		tokenString, err := h.j.Sign(ctx, jwt.RegisteredClaims{
			Subject:   strconv.Itoa(int(user.ID)),
			ExpiresAt: jwt.NewNumericDate(exp),
		})
		if err != nil {
			return response.Error(rw, response.MessageTokenExpired, bunrouter.H{})
		}
		http.SetCookie(rw, &http.Cookie{
			Name:    constants.SessionCookieName,
			Value:   tokenString,
			Expires: exp,
		})
		return response.WriteOK(rw, response.MessageOK, bunrouter.H{})
	}
}

type SSOLoginRequest struct {
	Redirect string `json:"redirect"`
}

type SSOLoginResponse struct {
	Redirect string `json:"redirect"`
}

// 返回时，ticket应该放到响应头
// token已经在中间件验证
// 获取重定向，返回重定向
func (h *Handler) SSOLogin() bunrouter.HandlerFunc {
	return func(rw http.ResponseWriter, r bunrouter.Request) error {
		ctx := r.Context()
		var request SSOLoginRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			return response.Error(rw, response.MessageBindError, bunrouter.H{})
		}

		var site string
		{
			redirectURL, err := url.Parse(request.Redirect)
			if err != nil {
				log.Error(ctx, err.Error())
				return response.Error(rw, response.MessageBadUrlParse, bunrouter.H{})
			}
			site = redirectURL.Scheme + "://" + redirectURL.Host
		}

		var app model.Application
		{
			db := h.db.Find(&app, "site=?", site)
			if err := db.Error; err != nil {
				log.Error(ctx, err.Error())
				return response.Error(rw, response.MessageDatabaseConnectionError, bunrouter.H{})
			} else if db.RowsAffected != 1 {
				return response.Error(rw, response.MessageAppExist, bunrouter.H{})
			}
		}
		id, err := strconv.Atoi(middleware.ContextJWTClaims{}.Value(ctx).Subject)
		if err != nil {
			return err
		}
		user, ok, err := isExistUserByID(uint(id), h.db)
		if err != nil {
			return response.Error(rw, response.MessageUnauthorized, bunrouter.H{})
		}
		if !ok {
			return response.Error(rw, response.MessageUnauthorized, bunrouter.H{})
		}
		ticket := h.r.RandString(20)
		if err := util.SetTicketToRedis(ctx, ticket, h.redisDB, util.UserInfo{
			ID:       user.ID,
			Username: user.Username,
		}); err != nil {
			return response.Error(rw, response.MessageBadTicket, bunrouter.H{})
		}

		u, err := url.Parse(app.Redirect)
		if err != nil {
			log.Error(ctx, err.Error())
			return response.Error(rw, response.MessageBadUrlParse, bunrouter.H{})
		}
		params := url.Values{}
		params.Set("redirect", url.QueryEscape(request.Redirect))
		params.Set("ticket", ticket)
		u.RawQuery = params.Encode()
		return response.WriteOK(rw, response.MessageOK, SSOLoginResponse{
			Redirect: u.String(),
		})

	}
}

type SSOVerifyRequest struct {
	Ticket string `json:"ticket"`
}

func (h *Handler) SSOVerify() bunrouter.HandlerFunc {
	return func(rw http.ResponseWriter, r bunrouter.Request) error {
		ctx := r.Context()
		appKey := r.Header.Get(constants.HTTPHeaderAppKey)
		if appKey == "" {
			return response.Error(rw, response.MessageBindError, bunrouter.H{})
		}
		var app model.Application
		{
			db := h.db.Find(&app, "app_key=?", appKey)
			if err := db.Error; err != nil {
				log.Error(ctx, err.Error())
				return response.Error(rw, response.MessageDatabaseConnectionError, bunrouter.H{})
			} else if db.RowsAffected != 1 {
				return response.Error(rw, response.MessageDatabaseConnectionError, bunrouter.H{})
			}
		}
		var request SSOVerifyRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			return response.Error(rw, response.MessageBindError, bunrouter.H{})
		}

		info, err := util.GetTicketFromRedis(ctx, request.Ticket, h.redisDB)
		if err != nil {
			if errors.Is(err, util.ErrTicketNotExists) {
				return response.Error(rw, response.MessageBadTicket, bunrouter.H{})
			}
			return err
		}

		tokenString, err := h.j.Sign(ctx, jwt.RegisteredClaims{
			Subject: info.Username,
		})
		if err != nil {
			return err
		}
		return response.WriteOK(rw, response.MessageOK, bunrouter.H{"token": tokenString})
	}
}
