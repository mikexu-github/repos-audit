package restful

import (
	"github.com/gin-gonic/gin"
	"github.com/quanxiang-cloud/audit/pkg/config"
	"github.com/quanxiang-cloud/audit/pkg/misc/logger"
)

const (
	// DebugMode indicates mode is debug.
	DebugMode = "debug"
	// ReleaseMode indicates mode is release.
	ReleaseMode = "release"
)

// Router 路由
type Router struct {
	c *config.Config

	engine *gin.Engine
}

// NewRouter 开启路由
func NewRouter(c *config.Config) (*Router, error) {
	engine, err := newRouter(c)
	if err != nil {
		return nil, err
	}
	v1 := engine.Group("api/v1/audit")

	audit, err := NewAudit(c)
	if err != nil {
		return nil, err
	}

	k := v1.Group("/search")
	{
		k.POST("", audit.SearchAudit)
	}

	return &Router{
		c:      c,
		engine: engine,
	}, nil
}

func newRouter(c *config.Config) (*gin.Engine, error) {
	if c.Model == "" || (c.Model != ReleaseMode && c.Model != DebugMode) {
		c.Model = ReleaseMode
	}
	gin.SetMode(c.Model)
	engine := gin.New()

	engine.Use(logger.GinLogger(), logger.GinRecovery())

	return engine, nil
}

// Run 启动服务
func (r *Router) Run() {
	r.engine.Run(r.c.Port)
}

// Close 关闭服务
func (r *Router) Close() {
}
