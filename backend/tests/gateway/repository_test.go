package gateway_test

import (
	"testing"

	"backend/internal/gateway"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}

	sqlDB, err := db.DB()
	if err == nil {
		t.Cleanup(func() {
			_ = sqlDB.Close()
		})
	}

	if err := gateway.MigrateGateway(db); err != nil {
		if sqlDB != nil {
			_ = sqlDB.Close()
		}
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestGetById_Found(t *testing.T) {
	db := newTestDB(t)

	// inserimento diretto senza usare gatewayEntity (map con colonne)
	id := uuid.New().String()
	if err := db.Table("gateways").Create(map[string]interface{}{
		"id":                id,
		"name":              "gw-test",
		"tenant_id":         "",
		"status":            "inactive",
		"interval_limit":    5,
		"public_identifier": "",
	}).Error; err != nil {
		t.Fatalf("insert: %v", err)
	}

	repo := gateway.NewGatewayPostgreRepository(db) // ora ritorna GatewayRepository
	uuidId, _ := uuid.Parse(id)
	got, err := repo.GetById(uuidId)
	if err != nil {
		t.Fatalf("GetById error: %v", err)
	}
	if got.Id != uuidId {
		t.Fatalf("expected id %v, got %v", uuidId, got.Id)
	}
}

func TestGetById_NotFound(t *testing.T) {
	db := newTestDB(t)
	repo := gateway.NewGatewayPostgreRepository(db)
	_, err := repo.GetById(uuid.New())
	if err == nil || err != gorm.ErrRecordNotFound {
		t.Fatalf("expected gorm.ErrRecordNotFound, got %v", err)
	}
}
