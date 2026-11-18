package data

import (
	"context"
	"fmt"
	"sync"

	"recon-ai/app/core/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewReconRepo)

// Data 持有基础设施连接。
// DB 采用「首次使用才连接」的惰性策略：数据库未就绪时进程仍可启动，
// 由 /healthz 与 /readyz 暴露状态，而不是让整个服务起不来。
type Data struct {
	conf *conf.Bootstrap
	log  *log.Helper

	mu  sync.Mutex
	db  *gorm.DB
	err error
}

func NewData(c *conf.Bootstrap, logger log.Logger) (*Data, func(), error) {
	d := &Data{
		conf: c,
		log:  log.NewHelper(log.With(logger, "module", "data/core")),
	}
	cleanup := func() {
		d.close()
	}
	return d, cleanup, nil
}

// DB 取带 ctx 的 DB 句柄；首次调用时建立连接
func (d *Data) DB(ctx context.Context) (*gorm.DB, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.db == nil {
		db, err := gorm.Open(mysql.Open(d.conf.Data.Database.Source), &gorm.Config{
			Logger: gormlogger.Default.LogMode(gormlogger.Warn),
		})
		if err != nil {
			d.err = fmt.Errorf("open mysql: %w", err)
			return nil, d.err
		}
		sqlDB, err := db.DB()
		if err != nil {
			return nil, fmt.Errorf("get sql.DB: %w", err)
		}
		sqlDB.SetMaxIdleConns(int(d.conf.Data.Database.MaxIdle))
		sqlDB.SetMaxOpenConns(int(d.conf.Data.Database.MaxOpen))
		if d.conf.Data.Database.ConnMaxLifetime != nil {
			sqlDB.SetConnMaxLifetime(d.conf.Data.Database.ConnMaxLifetime.AsDuration())
		}
		d.db = db
		d.err = nil
	}

	if d.err != nil {
		return nil, d.err
	}
	return d.db.WithContext(ctx), nil
}

// Ping 检查数据库连通性，供健康检查使用
func (d *Data) Ping(ctx context.Context) error {
	db, err := d.DB(ctx)
	if err != nil {
		return err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

func (d *Data) close() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.db == nil {
		return
	}
	sqlDB, err := d.db.DB()
	if err != nil {
		d.log.Errorw("msg", "get sql.DB for close failed", "err", err)
		return
	}
	if err := sqlDB.Close(); err != nil {
		d.log.Errorw("msg", "close db failed", "err", err)
	}
}
