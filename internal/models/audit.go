package models

import "context"

// OperationType 操作类型
type OperationType string

const (
	// LoginOperationName 登录
	LoginOperationName = "登录"
	// LoginOperationType login
	LoginOperationType OperationType = "login"

	// GETOperationName 查询
	GETOperationName = "查询"
	// GETOperationType get
	GETOperationType OperationType = "get"

	// POSTOperationName 新增
	POSTOperationName = "新增"
	// POSTOperationType post
	POSTOperationType OperationType = "post"

	// PUTOperationName 修改
	PUTOperationName = "修改"
	// PUTOperationType put
	PUTOperationType OperationType = "put"

	// DELETEOperationName 删除
	DELETEOperationName = "删除"
	// DELETEOperationType delete
	DELETEOperationType OperationType = "delete"
)

// Module 模块
type Module string

const (
	// LoginModule 登录模块
	LoginModule Module = "login"
	// LoginModuleName 登录注册
	LoginModuleName = "登录注册"

	// GoalieModule 权限模块
	GoalieModule Module = "goalie"
	// GoalieModuleName 权限管理
	GoalieModuleName = "权限管理"
)

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

// Audit 审计日志
type Audit struct {
	ID              string        `json:"-"`
	RequestID       string        `json:"requestID,omitempty"`
	UserID          string        `json:"userID,omitempty"`
	UserName        string        `json:"userName,omitempty"`
	OperationTime   int64         `json:"operationTime,omitempty"`
	OperationUA     string        `json:"operationUA,omitempty"`
	OperationModule string        `json:"operationModule,omitempty"`
	OperationType   OperationType `json:"operationType,omitempty"`
	GEO             *GEO          `json:"geo,omitempty"`
	Detail          string        `json:"detail,omitempty"`
	CreateAt        int64         `json:"createAt,omitempty"`
}

// AuditRepo 审计日志[存储服务]
type AuditRepo interface {
	// Create 添加审计日志
	Create(context.Context, *Audit) error

	// Search 查询审计日志
	Search(ctx context.Context, userName string,
		operationTimeBegin, operationTimeEnd int64, page, size int) ([]*Audit, int64, error)
}
