package repov2

import (
	"context"
	"core-server/internal/config"
	"core-server/internal/infras/repov2/ent"
	"fmt"

	"entgo.io/ent/dialect/sql"
)

type EntClient struct {
	db *ent.Client
}

func NewEntClient(cfg *config.Config) (*EntClient, error) {
	dsn := cfg.Mysql.DSN()

	drv, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	db := drv.DB()
	db.SetMaxIdleConns(cfg.Mysql.MaxIdleConn)
	db.SetMaxOpenConns(cfg.Mysql.MaxOpenConn)

	// 开启测试
	//if os.Getenv("ENV_LOCAL_TEST") != "" {
	//	client = client.Debug()
	//}

	return &EntClient{
		db: ent.NewClient(ent.Driver(drv)),
	}, nil
}

func (c *EntClient) Close() {
	if err := c.db.Close(); err != nil {
		fmt.Printf("fail to close sql connection: %v\n", err)
	}
}

func (b *EntClient) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) (err error) {
	tx, err := b.db.Tx(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if v := recover(); v != nil {
			// 回滚失败也只是记录，不掩盖原始 panic
			if rerr := tx.Rollback(); rerr != nil {
				err = fmt.Errorf("rolling back transaction: %w", rerr)
			}
			panic(v)
		}
	}()

	ctx = ent.NewContext(ctx, tx.Client())

	if err := fn(ctx); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			err = fmt.Errorf("%w: rolling back transaction: %v", err, rerr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (b *EntClient) DB(ctx context.Context) *ent.Client {
	db := ent.FromContext(ctx)
	return db
}
