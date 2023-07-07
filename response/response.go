package response

import (
	"encoding/json"
	"net/http"
)

type ResponseCode int8

const (
	ResponseCodeOK ResponseCode = iota
	ResponseCodeWarning
	ResponseCodeError
)

const (
	MessageOK                      = "ok"
	MessageDatabaseConnectionError = "error.database"
	MessageTokenExpired            = "error.token.expired"
	MessageBindError               = "bind.error"
	MessageCalculateOffset         = "calculate.offset"
	MessageBadTicket               = "bad.ticket"
	MessageIncorrectPassword       = "incorrect.password"
	MessageBadUrlParse             = "bad.url.parse"
	MessageAppExist                = "app.exist"
	MessageUnauthorized            = "unauthorized"
	MessageGetJWTError             = "get.jwt.error"
	MessageCheckJWTError           = "check.jwt.error"
	MessageUserIsExist             = "user.is.exist"
	MessageUserNotExist            = "user.not.exist"
)

type GenResponse[D any] struct {
	Message string       `json:"message"`
	Code    ResponseCode `json:"code"`
	Data    D            `json:"data"`
}

func NewResponse[D any](code ResponseCode, message string, data D) GenResponse[D] {
	return GenResponse[D]{
		Message: message,
		Code:    code,
		Data:    data,
	}
}

type PaginationData[P any] struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
	List     []P `json:"list"`
}

func NewPaginationData[P any](page, pageSize int, list []P) PaginationData[P] {
	return PaginationData[P]{
		Page:     page,
		PageSize: pageSize,
		List:     list,
	}
}

func WriteOK[T any](rw http.ResponseWriter, msg string, data T) error {
	return JSON(rw, msg, ResponseCodeOK, data)
}

func Warning[T any](rw http.ResponseWriter, msg string, data T) error {
	return JSON(rw, msg, ResponseCodeWarning, data)
}

func Error[T any](rw http.ResponseWriter, msg string, data T) error {
	return JSON(rw, msg, ResponseCodeError, data)
}

func JSON[T any](rw http.ResponseWriter, msg string, code ResponseCode, data T) error {
	rw.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(rw)
	response := NewResponse(code, msg, data)
	if err := enc.Encode(response); err != nil {
		return err
	}
	return nil
}
