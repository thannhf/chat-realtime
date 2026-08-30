package postgres

import (
	"media-service/config"
	"log"
	"media-service/internal/domain"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

func InitDatabase(cfg *config.DBConfig) *gorm.DB {
	masterDSN := cfg.WriteHost
	if masterDSN == "" {
		log.Fatalf("Lỗi: DB_MASTER_DSN không đươc để trống")
	}

	replicaDSN := cfg.ReadHost
	if replicaDSN == "" {
		log.Fatalf("Lỗi: DB_REPLICA_DSN không được để trống")
	}

	db, err := gorm.Open(postgres.Open(masterDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("Không thể connect đến MasterDB của media-service: %v", err)
	}

	err = db.Use(dbresolver.Register(dbresolver.Config{
		Sources: []gorm.Dialector{postgres.Open(masterDSN)},
		Replicas: []gorm.Dialector{postgres.Open(replicaDSN)},
		Policy: dbresolver.RandomPolicy{},
	}))

	if err != nil {
		log.Fatalf("Không thể cấu hình master-replica resolver cho media service %v", err)
	}

	log.Println("Kết nối Db media thành công")

	log.Println("Đang tiến hành Auto Migration cho media Service")
	err = db.AutoMigrate(
		&domain.Media{},
		&domain.MediaVariant{},
	)
	if err != nil {
		log.Fatalf("Auto Migration cho media service thất bại %v", err)
	}
	log.Println("Auto migration cho media thành công")

	return db 
}