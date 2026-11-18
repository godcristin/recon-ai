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

var ProviderSet = wire.NewSet(NewData, NewTransaction, NewReconRepo)

type txKey struct{}

type Transaction interface {
	InTx(context.Context, func(ctx context.Context) error) error
}

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

func NewTransaction(d *Data) Transaction {
	return d
}

func (d *Data) InTx(ctx context.Context, fn func(ctx context.Context) error) error {
	db, err := d.DB(ctx)
	if err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(ctx, txKey{}, tx))
	})
}

func (d *Data) DB(ctx context.Context) (*gorm.DB, error) {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx.WithContext(ctx), nil
	}

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
