package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/Shopify/sarama"
	"github.com/oschwald/geoip2-golang"
	"github.com/quanxiang-cloud/audit/internal/models"
	"github.com/quanxiang-cloud/audit/internal/models/es"
	"github.com/quanxiang-cloud/audit/pkg/client"
	"github.com/quanxiang-cloud/audit/pkg/config"
	"github.com/quanxiang-cloud/audit/pkg/misc/elastic2"
	"github.com/quanxiang-cloud/audit/pkg/misc/id2"
	"github.com/quanxiang-cloud/audit/pkg/misc/kafka"
	"github.com/quanxiang-cloud/audit/pkg/misc/logger"
	"github.com/quanxiang-cloud/audit/pkg/misc/time2"
	"go.uber.org/zap"
)

// Handler 审计日志
type Handler interface {
	Handler()
	Close() error
}

type handler struct {
	c         *config.Config
	auditRepo models.AuditRepo

	ctx      context.Context
	cancel   context.CancelFunc
	consumer sarama.ConsumerGroup
	ip2      *geoip2.Reader

	requset chan *client.Audit
}

// NewHandler new audit handler
func NewHandler(conf *config.Config) (Handler, error) {
	ip2, err := geoip2.Open(conf.GEOIP)
	if err != nil {
		return nil, err
	}

	elasticClient, err := elastic2.NewClient(&conf.Elastic, logger.Logger)
	if err != nil {
		return nil, err
	}

	return &handler{
		c:         conf,
		ip2:       ip2,
		auditRepo: es.NewAuditRepo(elasticClient),
	}, nil
}

func (h *handler) Handler() {
	ctx, cancel := context.WithCancel(context.Background())
	h.ctx = ctx
	h.cancel = cancel

	h.requset = make(chan *client.Audit, h.c.Handler.Buffer)
	for i := 0; i < h.c.Handler.NumOfProcessor; i++ {
		go h.process(ctx)
	}

	go h.receive(ctx)

}

func (h *handler) process(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("done**")
			return
		case entity := <-h.requset:
			ctx := logger.ReentryRequestID(context.Background(), entity.RequestID)

			audit := &models.Audit{
				ID:              id2.GenID(),
				UserID:          entity.UserID,
				UserName:        entity.UserName,
				OperationTime:   entity.OperationTime,
				OperationUA:     entity.OperationUA,
				OperationModule: entity.OperationModule,
				OperationType:   models.OperationType(entity.OperationType),
				CreateAt:        time2.NowUnixMill(),
				Detail:          entity.Detail,
			}
			// 解析IP
			ip, err := h.analysisIP(entity.IP)
			if err != nil {
				logger.Logger.Error(zap.String("analysisIP", err.Error()), logger.STDRequestID(ctx))
			}
			if ip != nil {
				audit.GEO = ip2City2GEO(ip)
				audit.GEO.IP = entity.IP
			}

			// 存入ES
			err = h.auditRepo.Create(ctx, audit)
			if err != nil {
				logger.Logger.Error(zap.String("audit create", err.Error()), logger.STDRequestID(ctx))
				continue
			}
		}
	}
}

func (h *handler) receive(ctx context.Context) {
	h.c.Kafka.Sarama.Version = sarama.V2_0_0_0
	consumer, err := kafka.NewConsumerGroup(h.c.Kafka, h.c.Handler.Group)
	if err != nil {
		logger.Logger.Error(zap.String("new consumer group", err.Error()))
		return
	}

	err = consumer.Consume(ctx, h.c.Handler.Topic, h)
	if err != nil {
		logger.Logger.Error(zap.String("client Consume", err.Error()))
		return
	}

	h.consumer = consumer
}

func (h *handler) Close() error {
	h.cancel()
	close(h.requset)
	return nil
}

// Setup Setup
func (h *handler) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup Cleanup
func (h *handler) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim ConsumeClaim
func (h *handler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		entity := new(client.Audit)
		err := json.Unmarshal(msg.Value, entity)
		if err != nil {
			logger.Logger.Error(zap.String("ConsumeClaim", err.Error()),
				zap.String("data", string(msg.Value)))
			continue
		}
		logger.Logger.Info(zap.String("data", string(msg.Value)))
		select {
		case <-h.ctx.Done():
			return nil
		case h.requset <- entity:
		}
		sess.MarkMessage(msg, "")
	}
	return nil
}

func (h *handler) analysisIP(ipStr string) (*geoip2.City, error) {
	if ipStr == "" || ipStr == "::1" || ipStr == "0.0.0.0" {
		return nil, nil
	}
	if isInnerIP(ipStr) {
		return nil, nil
	}
	ip := net.ParseIP(ipStr)
	return h.ip2.City(ip)
}

func isInnerIP(ipStr string) bool {
	ipNum := inetAton(ipStr)
	switch {
	case inetAton("10.255.255.255")>>24 == ipNum>>24,
		inetAton("172.16.255.255")>>20 == ipNum>>20,
		inetAton("192.168.255.255")>>16 == ipNum>>16,
		inetAton("100.64.255.255")>>22 == ipNum>>22,
		inetAton("127.255.255.255")>>24 == ipNum>>24:
		return true
	}

	return false
}

func inetAton(ipStr string) int64 {
	bits := strings.Split(ipStr, ".")

	b0, _ := strconv.Atoi(bits[0])
	b1, _ := strconv.Atoi(bits[1])
	b2, _ := strconv.Atoi(bits[2])
	b3, _ := strconv.Atoi(bits[3])

	var sum int64

	sum += int64(b0) << 24
	sum += int64(b1) << 16
	sum += int64(b2) << 8
	sum += int64(b3)

	return sum
}

const (
	zhCN = "zh-CN"
)

func ip2City2GEO(city *geoip2.City) *models.GEO {
	geo := &models.GEO{
		Country: city.Country.Names[zhCN],
		City:    city.City.Names[zhCN],
		Location: models.Location{
			AccuracyRadius: city.Location.AccuracyRadius,
			Latitude:       city.Location.Latitude,
			Longitude:      city.Location.Longitude,
			MetroCode:      city.Location.MetroCode,
			TimeZone:       city.Location.TimeZone,
		},
	}

	if len(city.Subdivisions) > 0 {
		geo.Province = city.Subdivisions[0].Names[zhCN]
	}
	return geo
}
