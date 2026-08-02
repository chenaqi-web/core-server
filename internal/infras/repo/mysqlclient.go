package repo

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"core-server/internal/config"
)

type DBClient struct {
	DB *sqlx.DB
}

type dbExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	GetContext(ctx context.Context, dest any, query string, args ...any) error
	SelectContext(ctx context.Context, dest any, query string, args ...any) error
}

func NewDBClient(cfg *config.Config) (*DBClient, error) {
	mysqlCfg := cfg.Mysql
	if mysqlCfg.Host == "" || mysqlCfg.Port == "" || mysqlCfg.DBName == "" {
		return nil, fmt.Errorf("mysql config is incomplete")
	}

	db, err := sqlx.Connect("mysql", mysqlCfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}

	db.SetMaxIdleConns(mysqlCfg.MaxIdleConn)
	db.SetMaxOpenConns(mysqlCfg.MaxOpenConn)

	log.Println("mysql connected successfully")
	return &DBClient{DB: db}, nil
}

func (c *DBClient) Close() error {
	return c.DB.Close()
}

func (c *DBClient) db(ctx context.Context) dbExecutor {
	if tx, ok := ctx.Value(txContextKey{}).(*sqlx.Tx); ok && tx != nil {
		return tx
	}
	return c.DB
}

// =====================================================================================================================

// 将事务对象存入上下文
type txContextKey struct{}

func withTx(ctx context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(ctx, txContextKey{}, tx)
}

// WithTransaction 事务管理器
func (c *DBClient) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	// 1. 开启事务
	tx, err := c.DB.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	//  业务函数可以从中提取事务对象
	txCtx := withTx(ctx, tx)
	if err := fn(txCtx); err != nil {
		// 错误回滚
		_ = tx.Rollback()
		return err
	}
	// 成功提交
	return tx.Commit()
}
