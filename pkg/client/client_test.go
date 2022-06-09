package client

import (
	"context"
	"testing"

	"github.com/Shopify/sarama"
	"github.com/quanxiang-cloud/audit/pkg/misc/kafka"
	"github.com/quanxiang-cloud/audit/pkg/misc/logger"
)

func TestClient(t *testing.T) {
	ctx := context.Background()
	ctx = logger.GenRequestID(ctx)

	const (
		topic = "test"
	)

	conf := kafka.Config{
		Broker: []string{"192.168.200.20:9092", "192.168.200.19:9092", "192.168.200.18:9092"},
	}
	conf.Sarama.Version = sarama.V2_0_0_0
	producer, err := kafka.NewSyncProducer(conf)
	if err != nil {
		t.Fatal(err)
		return
	}
	defer producer.Close()

	// client := New(producer)
	// err = client.Send(ctx, &Audit{
	// 	RequestID:     logger.STDRequestID(ctx).String,
	// 	UserID:        "1",
	// 	UserName:      "demo",
	// 	OperationTime: time2.NowUnixMill(),
	// 	OperationType: "login",
	// 	IP:            "222.212.94.41",
	// 	Detail:        "this is detail.",
	// }, Topic)

	if err != nil {
		t.Fatal(err)
		return
	}
}
