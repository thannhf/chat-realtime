package postgres

import (
	"chat-service/config"
	"chat-service/internal/domain"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

type Database struct {
	DB *gorm.DB
}

func InitDatabase(cfg *config.DBConfig) *gorm.DB {
	masterDSN := cfg.WriteHost
	if masterDSN == "" {
		log.Fatalf("Lỗi: DB_MASTER_DSN (WriteHost) không được để trống")
	}

	replicaDSN := cfg.ReadHost
	if replicaDSN == "" {
		replicaDSN = masterDSN
		log.Println("Cảnh báo: Không tìm thấy DB_REPLICA_DSN, tự động fallback về Master")
	}

	db, err := gorm.Open(postgres.Open(masterDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("Không thể kết nối tới Master DB: %v", err)
	}
	
	err = db.Use(dbresolver.Register(dbresolver.Config{
		Sources: []gorm.Dialector{postgres.Open(masterDSN)},
		Replicas: []gorm.Dialector{postgres.Open(replicaDSN)},
		Policy: dbresolver.RandomPolicy{},
	}))
	if err != nil {
		log.Fatalf("Không thể cấu hình Master-Replica Resolver: %v", err)
	}

	log.Println("Kết nối cơ sở dữ liệu Master-Replica thành công!")
	log.Println("Đang tiến hành Auto Migration cho Chat Service...")

	err = db.AutoMigrate(
		&domain.Conversation{},
		&domain.ConversationMember{},
		&domain.Message{},
		&domain.ConversationRead{},
	)
	if err != nil {
		log.Fatalf("Auto Migration thất bại: %v", err)
	}
	log.Println("Auto Migration hoàn tất thành công!")

	return db
}