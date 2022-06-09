package logic

import (
	"context"

	"github.com/quanxiang-cloud/audit/internal/models"
	"github.com/quanxiang-cloud/audit/internal/models/es"
	"github.com/quanxiang-cloud/audit/pkg/config"
	"github.com/quanxiang-cloud/audit/pkg/misc/elastic2"
	"github.com/quanxiang-cloud/audit/pkg/misc/logger"
)

// Audit Audit
type Audit interface {
	// SearchAudit 查询审计日志
	SearchAudit(ctx context.Context, req *SearchAuditReq) (*SearchAuditResp, error)
}

// NewAudit new audit
func NewAudit(conf *config.Config) (Audit, error) {
	elasticClient, err := elastic2.NewClient(&conf.Elastic, logger.Logger)
	if err != nil {
		return nil, err
	}

	return &audit{
		c:         conf,
		auditRepo: es.NewAuditRepo(elasticClient),
	}, nil
}

type audit struct {
	c *config.Config

	auditRepo models.AuditRepo
}

// SearchAuditReq 查询审计日志[参数]
type SearchAuditReq struct {
	UserName           string `json:"userName"`
	OperationTimeBegin int64  `json:"operationTimeBegin"`
	OperationTimeEnd   int64  `json:"operationTimeEnd"`

	Page int `json:"page"`
	Size int `json:"size"`
}

// SearchAuditVO 查询审计日志VO
type SearchAuditVO struct {
	ID              string `json:"-"`
	RequestID       string `json:"requestID,omitempty"`
	UserID          string `json:"userID,omitempty"`
	UserName        string `json:"userName,omitempty"`
	OperationTime   int64  `json:"operationTime,omitempty"`
	OperationType   string `json:"operationType,omitempty"`
	OperationUA     string `json:"operationUA,omitempty"`
	OperationModule string `json:"operationModule,omitempty"`
	GEO             GEO    `json:"geo,omitempty"`
	Detail          string `json:"detail,omitempty"`
	CreateAt        int64  `json:"createAt,omitempty"`
}

// GEO geo
type GEO struct {
	IP       string   `json:"ip,omitempty"`
	Country  string   `json:"country,omitempty"`
	Province string   `json:"province,omitempty"`
	City     string   `json:"city,omitempty"`
	Location Location `json:"location,omitempty"`
}

// Location location
type Location struct {
	AccuracyRadius uint16  `json:"accuracyRadius,omitempty"`
	Latitude       float64 `json:"latitude,omitempty"`
	Longitude      float64 `json:"longitude,omitempty"`
	MetroCode      uint    `json:"metroCode,omitempty"`
	TimeZone       string  `json:"timeZone,omitempty"`
}

// SearchAuditResp  查询审计日志[返回值]
type SearchAuditResp struct {
	Audit []*SearchAuditVO `json:"audit"`
	Total int64            `json:"total"`
}

const ()

// SearchAudit 查询审计日志
func (a *audit) SearchAudit(ctx context.Context, req *SearchAuditReq) (*SearchAuditResp, error) {
	entiries, total, err := a.auditRepo.Search(ctx,
		req.UserName,
		req.OperationTimeBegin*1e3, req.OperationTimeEnd*1e3,
		req.Page, req.Size)
	if err != nil {
		return nil, err
	}

	resp := &SearchAuditResp{
		Total: total,
		Audit: make([]*SearchAuditVO, 0, len(entiries)),
	}

	for _, elem := range entiries {
		audit := new(SearchAuditVO)
		serializeSearchAuditVO(audit, elem)
		resp.Audit = append(resp.Audit, audit)
	}

	return resp, nil
}

func serializeSearchAuditVO(dst *SearchAuditVO, src *models.Audit) {
	dst.ID = src.ID
	dst.UserID = src.UserID
	dst.UserName = src.UserName
	dst.OperationTime = src.OperationTime
	dst.OperationType = string(src.OperationType)
	dst.OperationUA = src.OperationUA
	dst.OperationModule = src.OperationModule
	if src.GEO != nil {
		dst.GEO = GEO{
			IP:       src.GEO.IP,
			Country:  src.GEO.Country,
			Province: src.GEO.Province,
			City:     src.GEO.City,
			Location: Location{
				AccuracyRadius: src.GEO.Location.AccuracyRadius,
				Latitude:       src.GEO.Location.Latitude,
				Longitude:      src.GEO.Location.Longitude,
				MetroCode:      src.GEO.Location.MetroCode,
				TimeZone:       src.GEO.Location.TimeZone,
			},
		}
	}
	dst.Detail = src.Detail
	dst.CreateAt = src.CreateAt
}
