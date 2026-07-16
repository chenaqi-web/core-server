package repo

import (
	"backend/core-server/internal/model/entity"
	"context"
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"backend/core-server/internal/config"
)

type DBClient struct {
	DB *gorm.DB
}

func NewDBClient(cfg *config.Config) (*DBClient, error) {
	mysqlCfg := cfg.Mysql
	if mysqlCfg.Host == "" || mysqlCfg.Port == "" || mysqlCfg.DBName == "" {
		return nil, fmt.Errorf("mysql config is incomplete")
	}

	db, err := gorm.Open(mysql.Open(mysqlCfg.DSN()), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql db: %w", err)
	}

	sqlDB.SetMaxIdleConns(mysqlCfg.MaxIdleConn)
	sqlDB.SetMaxOpenConns(mysqlCfg.MaxOpenConn)

	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}

	log.Println("mysql connected successfully")
	return &DBClient{DB: db}, nil
}

func migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&entity.User{},
		&entity.InteractionLike{},
		&entity.InteractionCount{},
	)
}

func (c *DBClient) Close() error {
	sqlDB, err := c.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

func (c *DBClient) GetDB() *gorm.DB {
	return c.DB
}

func (c *DBClient) db(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txContextKey{}).(*gorm.DB); ok && tx != nil {
		return tx.WithContext(ctx)
	}
	return c.DB.WithContext(ctx)
}

type txContextKey struct{}

func withTx(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, txContextKey{}, tx)
}

func (c *DBClient) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return c.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(withTx(ctx, tx))
	})
}
