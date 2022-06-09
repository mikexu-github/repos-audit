package client

import (
	"encoding/json"
	"fmt"

	"github.com/quanxiang-cloud/audit/internal/models"
	"github.com/quanxiang-cloud/audit/pkg/misc/header2"
	"github.com/quanxiang-cloud/audit/pkg/misc/logger"
	"github.com/quanxiang-cloud/audit/pkg/misc/time2"

	"github.com/Shopify/sarama"
	"github.com/gin-gonic/gin"
)

const (
	// topic kafka topic
	topic = "audit-log"

	// userAgent user agent
	userAgent = "User-Agent"

	// detailFormat format detail
	detailFormat = "%s操作的路由是%s,操作参数是%s"
)

// Audit 审计日志
type Audit struct {
	RequestID       string
	UserID          string
	UserName        string
	OperationTime   int64
	OperationUA     string
	OperationType   string
	OperationModule string
	IP              string
	Detail          string
}

// Client 审计客户端
type Client struct {
	producer sarama.SyncProducer
}

// New 创建一个审计客户端
func New(producer sarama.SyncProducer) *Client {
	return &Client{
		producer: producer,
	}
}

// Send 添加审计日志
func (c *Client) Send(ctx *gin.Context, _t models.OperationType, m models.Module, param interface{}) error {

	context := logger.CTXTransfer(ctx)

	audit := Audit{
		RequestID:       logger.STDRequestID(context).String,
		OperationTime:   time2.NowUnixMill(),
		OperationType:   string(_t),
		OperationUA:     ctx.Request.Header.Get(userAgent),
		OperationModule: string(m),
		IP:              ctx.ClientIP(),
	}

	profile := header2.GetProfile(ctx)
	audit.UserID = profile.UserName
	audit.UserName = profile.UserName

	paramByte, err := json.Marshal(param)
	if err != nil {
		return err
	}
	audit.Detail = fmt.Sprintf(detailFormat, audit.UserName, ctx.Request.RequestURI, string(paramByte))

	auditByte, err := json.Marshal(audit)
	if err != nil {
		return err
	}
	_, _, err = c.producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(auditByte),
	})
	return err
}

// Close 关闭连接
func (c *Client) Close() error {
	return c.producer.Close()
}
