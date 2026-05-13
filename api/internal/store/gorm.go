package store

import (
	"ai-workbench-api/internal/model"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var gormDB *gorm.DB

// GetDB returns the GORM DB instance. Panics if not initialized.
func GetDB() *gorm.DB {
	if gormDB == nil {
		logrus.Fatal("gorm db not initialized")
	}
	return gormDB
}

// GormOK returns whether GORM DB is available.
func GormOK() bool {
	return gormDB != nil
}

// InitGormDB initializes GORM using the same DSN as raw sql.DB.
func InitGormDB() {
	dsn := viper.GetString("mysql.dsn")
	if dsn == "" {
		logrus.Warn("gorm: mysql.dsn empty, skipping")
		return
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		logrus.WithError(err).Warn("gorm: open failed, CMDB will use memory fallback")
		return
	}
	gormDB = db
	autoMigrateCmdb()
	SeedCmdbDefaults()
	SeedMcpDefaults()
}

func autoMigrateCmdb() {
	err := gormDB.AutoMigrate(
		&model.CmdbCategory{},
		&model.CmdbObject{},
		&model.CmdbAttribute{},
		&model.CmdbInstance{},
		&model.CmdbRelationType{},
		&model.CmdbInstanceRelation{},
		&model.McpServer{},
		&model.MonitorAlertEventRecord{},
		&model.ProbeCheck{},
		&model.ProbeCheckResult{},
		&model.ProbeStatusPage{},
		&model.ProbeIncident{},
		&model.ProbeSubscription{},
		&model.ProbeNotificationBinding{},
		&model.ProbeAlertBinding{},
	)
	if err != nil {
		logrus.WithError(err).Error("gorm: auto migrate cmdb failed")
	}
}
