package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"git.blauwelle.com/go/crate/log"
	"github.com/uptrace/bunrouter"

	"git.blauwelle.com/go/crate/cmd/sso/model"
	"git.blauwelle.com/go/crate/cmd/sso/response"
)

func (h *Handler) CreateApp() bunrouter.HandlerFunc {
	return func(rw http.ResponseWriter, r bunrouter.Request) error {
		var application model.Application
		if err := json.NewDecoder(r.Body).Decode(&application); err != nil {
			return response.Error(rw, response.MessageBindError, bunrouter.H{})
		}
		//生成app_key
		application.AppKey = h.r.RandString(20)
		if dbCreate := h.db.Create(&application); dbCreate.Error != nil {
			return response.Error(rw, response.MessageDatabaseConnectionError, bunrouter.H{})
		}
		return response.WriteOK(rw, response.MessageOK, bunrouter.H{"app_key": application.AppKey})
	}
}

func (h *Handler) SearchApp() bunrouter.HandlerFunc {
	return func(rw http.ResponseWriter, r bunrouter.Request) error {
		pageSize := r.URL.Query().Get("pageSize")
		name := r.URL.Query().Get("name")
		page := r.URL.Query().Get("page")
		// 设置默认每页记录数
		defaultPageSize := 20

		var applications []model.Application
		var count int64

		// 构建查询条件
		query := h.db.Model(&model.Application{})

		// 添加模糊查询条件
		if name != "" {
			query = query.Where("name LIKE ?", "%"+name+"%")
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
		if dbFind := query.Offset(offset).Limit(pageSizeInt).Find(&applications); dbFind.Error != nil {
			return response.Error(rw, response.MessageDatabaseConnectionError, bunrouter.H{})
		}
		return response.WriteOK(rw, response.MessageOK, response.NewPaginationData(pageInt, pageSizeInt, applications))
	}
}

type DeleteAppRequest struct {
	ID uint `json:"id"`
}

func (h *Handler) DeleteApp() bunrouter.HandlerFunc {
	return func(rw http.ResponseWriter, req bunrouter.Request) error {
		ctx := req.Context()
		var request DeleteAppRequest

		if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
			log.Error(ctx, err.Error())
			return response.Error(rw, response.MessageBindError, bunrouter.H{})
		}
		if db := h.db.Delete(&model.Application{}, "id=?", request.ID); db.Error != nil {

			return response.Error(rw, response.MessageDatabaseConnectionError, bunrouter.H{})
		}

		return response.WriteOK(rw, response.MessageOK, bunrouter.H{})
	}
}

type UpdateAppRequest struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Site     string `json:"site"`
	Redirect string `json:"redirect"`
}

func (h *Handler) UpdateApp() bunrouter.HandlerFunc {
	return func(rw http.ResponseWriter, req bunrouter.Request) error {
		ctx := req.Context()
		var request UpdateAppRequest

		if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
			log.Error(ctx, err.Error())
			return response.Error(rw, response.MessageBindError, bunrouter.H{})
		}
		updates := map[string]interface{}{
			"name":     request.Name,
			"site":     request.Site,
			"redirect": request.Redirect,
		}
		if h.db.Model(&model.Application{}).Where("id=?", request.ID).Updates(updates).Error != nil {
			return response.Error(rw, response.MessageDatabaseConnectionError, bunrouter.H{})
		}
		return response.WriteOK(rw, response.MessageOK, bunrouter.H{})
	}
}

// 计算偏移量
func calculateOffset(page string, pageSize int, totalRecords int64) (int, error) {
	pageNumber, err := strconv.Atoi(page)
	if err != nil {
		return 0, errors.New("invalid page number")
	}

	if pageNumber <= 0 {
		pageNumber = 1
	}

	offset := (pageNumber - 1) * pageSize
	if offset >= int(totalRecords) {
		return 0, errors.New("page out of range")
	}

	return offset, nil
}
