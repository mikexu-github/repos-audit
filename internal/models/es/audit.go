package es

import (
	"context"
	"encoding/json"
	"github.com/olivere/elastic/v7"

	"github.com/quanxiang-cloud/audit/internal/models"
)

type auditRepo struct {
	db *elastic.Client
}

// NewAuditRepo new audit repo
func NewAuditRepo(db *elastic.Client) models.AuditRepo {
	return &auditRepo{
		db: db,
	}
}

func (a *auditRepo) index() string {
	return "audit"
}

func (a *auditRepo) Create(ctx context.Context, entity *models.Audit) error {
	_, err := a.db.Index().
		Index(a.index()).
		Id(entity.ID).
		BodyJson(entity).
		Do(ctx)
	return err
}

func (a *auditRepo) Search(ctx context.Context, userName string,
	operationTimeBegin, operationTimeEnd int64, page, size int) ([]*models.Audit, int64, error) {
	query := elastic.NewBoolQuery()

	boolQuery := make([]elastic.Query, 0, 2)
	if userName != "" {
		boolQuery = append(boolQuery, elastic.NewMatchQuery("userName", userName))
	}
	if operationTimeBegin != 0 {
		boolQuery = append(boolQuery, elastic.NewRangeQuery("operationTime").Gte(operationTimeBegin))
	}
	if operationTimeEnd != 0 {
		boolQuery = append(boolQuery, elastic.NewRangeQuery("operationTime").Lt(operationTimeEnd))
	}

	query = query.Must(boolQuery...)

	searchResult, err := a.db.Search().
		Index(a.index()).
		Query(query).
		Sort("operationTime", false).
		From((page - 1) * size).Size(size).
		Do(ctx)
	if err != nil {
		return nil, 0, err
	}

	var total int64
	audits := make([]*models.Audit, 0, size)
	if searchResult.Hits != nil {
		for _, hit := range searchResult.Hits.Hits {
			audit := new(models.Audit)
			err := json.Unmarshal(hit.Source, audit)
			if err != nil {
				return nil, 0, err
			}
			audits = append(audits, audit)
		}
		total = searchResult.Hits.TotalHits.Value
	}
	return audits, total, nil
}
