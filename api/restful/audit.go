package restful

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/quanxiang-cloud/audit/internal/logic"
	"github.com/quanxiang-cloud/audit/pkg/config"
	"github.com/quanxiang-cloud/audit/pkg/misc/logger"
	"github.com/quanxiang-cloud/audit/pkg/misc/resp"
)

// Audit gin audit
type Audit struct {
	audit logic.Audit
}

// NewAudit new audit gin
func NewAudit(conf *config.Config) (*Audit, error) {
	k, err := logic.NewAudit(conf)
	if err != nil {
		return nil, err
	}
	return &Audit{
		audit: k,
	}, nil
}

// SearchAudit 查询
func (a *Audit) SearchAudit(c *gin.Context) {
	req := &logic.SearchAuditReq{}
	if err := c.ShouldBind(req); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	resp.Format(a.audit.SearchAudit(logger.CTXTransfer(c), req)).Context(c)
}
