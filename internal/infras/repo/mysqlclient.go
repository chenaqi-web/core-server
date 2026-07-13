package repo

import (
	"backend/core-server/internal/model/entity"
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

	maxIdleConn := mysqlCfg.MaxIdleConn
	if maxIdleConn == 0 {
		maxIdleConn = 10
	}
	maxOpenConn := mysqlCfg.MaxOpenConn
	if maxOpenConn == 0 {
		maxOpenConn = 100
	}
	sqlDB.SetMaxIdleConns(maxIdleConn)
	sqlDB.SetMaxOpenConns(maxOpenConn)

	if err := migrate(db); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}

	log.Println("mysql connected")
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
