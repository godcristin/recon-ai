package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
)

// ReconRepo 对账任务的持久化
type ReconRepo struct {
	data *Data
	log  *log.Helper
}

func NewReconRepo(d *Data, logger log.Logger) *ReconRepo {
	return &ReconRepo{
		data: d,
		log:  log.NewHelper(log.With(logger, "module", "data/recon")),
	}
}

// CreateTask 插入一条对账任务，返回自增 ID
func (r *ReconRepo) CreateTask(ctx context.Context, name string) (int64, error) {
	db, err := r.data.DB(ctx)
	if err != nil {
		return 0, err
	}

	var id int64
	if err := db.Raw(`INSERT INTO recon_task (name, status) VALUES (?, 'CREATED') RETURNING id`, name).
		Scan(&id).Error; err != nil {
		r.log.WithContext(ctx).Errorw("msg", "CreateTask failed", "err", err, "name", name)
		return 0, err
	}
	return id, nil
}
