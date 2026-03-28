package gateway_test

import (
	"testing"

	"backend/internal/gateway"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestGatewayPostgreRepository_Save_DeduplicatesByPublicIdentifier(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, gateway.MigrateGateway(db))

	repo := gateway.NewGatewayPostgreRepository(db)
	savePort, _, _ := gateway.NewGatewayPostgreAdapter(repo)

	tenantID := uuid.New()
	first := gateway.Gateway{
		Id:               uuid.New(),
		Name:             "gw1",
		TenantId:         &tenantID,
		Status:           gateway.GATEWAY_STATUS_ACTIVE,
		IntervalLimit:    10,
		PublicIdentifier: "pub-123",
	}

	err = savePort.Save(first)
	require.NoError(t, err)

	var count int64
	err = db.Table("gateways").Where("public_identifier = ?", "pub-123").Count(&count).Error
	require.NoError(t, err)
	require.Equal(t, int64(1), count)

	second := first
	second.Id = uuid.New()
	second.Name = "gw2"

	err = savePort.Save(second)
	require.NoError(t, err)

	err = db.Table("gateways").Where("public_identifier = ?", "pub-123").Count(&count).Error
	require.NoError(t, err)
	require.Equal(t, int64(1), count)
}
