package integration

import (
	"fmt"
	"testing"

	"backend/internal/auth"
	"backend/internal/tenant"
	"backend/internal/user"
	"backend/tests/helper"

	clouddb "backend/internal/infra/database/cloud_db/connection"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

/*
Ritorna il pre setup per creare il tenant con ID tenantId.
*/
func PreSetupCreateTenant(tenantId uuid.UUID, canImpersonate bool) helper.IntegrationTestPreSetup {
	return func(deps helper.IntegrationTestDeps) bool {
		db := (*gorm.DB)(deps.CloudDB)

		tenantEntity := tenant.TenantEntity{
			ID:             tenantId.String(),
			Name:           "test tenant",
			CanImpersonate: canImpersonate,
		}
		if err := db.Clauses().Create(&tenantEntity).Error; err != nil {
			return false
		}
		// create schema
		schemaName := "tenant_" + tenantId.String()
		if err := db.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS \"%s\"", schemaName)).Error; err != nil {
			return false
		}
		// migrate tenant_members in this schema
		if err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Exec(fmt.Sprintf("set local search_path to \"%s\"", schemaName)).Error; err != nil {
				return err
			}
			return tx.AutoMigrate(&user.TenantMemberEntity{}, &auth.TenantConfirmTokenEntity{}, &auth.TenantPasswordTokenEntity{})
		}); err != nil {
			return false
		}
		return true
	}
}

/*
Ritorna il pre setup per aggiungere il Tenant User alla tabella tenant_members del proprio tenant.
Se setPath è impostato a true, allora il Path di tc viene impostato al path per operazioni CRUD su tale tenant user.

NOTA: Se si mettono più di questi preSetup con setPath = truein un TC di eliminazione, quest'ultimo
controllerà l'eliminazione del tenant admin associato all'ULTIMO di questi preSetup nella lista
*/
func PreSetupAddTenantUser(
	t *testing.T,
	tc *helper.IntegrationTestCase,
	/*
		NOTA: Importantissimo che entity sia passato come valore e non come puntatore!
		In modo tale da copiarlo per valore e far funzionare l'ID autoincrementale
	*/
	entity user.TenantMemberEntity,
	setPath bool,
) helper.IntegrationTestPreSetup {
	t.Helper()
	return func(deps helper.IntegrationTestDeps) bool {
		if entity.TenantId == "" {
			t.Fatalf("cannot call preSetupAddTenantUser with TenantMemberEntity with no tenant id")
			return false
		}

		if entity.Role != "tenant_user" {
			t.Fatalf("cannot call preSetupAddTenantUser with TenantMemberEntity with role != \"tenant_user\"")
			return false
		}

		localPreSetup, createdUserId := preSetupAddTenantMember_ReturnUserId(t, &entity)
		ok := localPreSetup(deps)
		if ok && setPath {
			(*tc).Path = fmt.Sprintf("/api/v1/tenant/%v/tenant_user/%d", entity.TenantId, *createdUserId)
		}

		return ok
	}
}

/*
Ritorna il pre setup per aggiungere il Tenant Admin alla tabella tenant_members del proprio tenant. Se setPath è impostato a true, allora
il Path di tc viene impostato al path per operazioni CRUD su tale Tenant Admin.

NOTA: Se si mettono più di questi preSetup con setPath = truein un TC di eliminazione, quest'ultimo
controllerà l'eliminazione del tenant admin associato all'ULTIMO di questi preSetup nella lista
*/
func PreSetupAddTenantAdmin(
	t *testing.T,
	tc *helper.IntegrationTestCase,
	/*
		NOTA: Importantissimo che entity sia passato come valore e non come puntatore!
		In modo tale da copiarlo per valore e far funzionare l'ID autoincrementale
	*/
	entity user.TenantMemberEntity,
	setPath bool,
) helper.IntegrationTestPreSetup {
	t.Helper()
	return func(deps helper.IntegrationTestDeps) bool {
		if entity.TenantId == "" {
			t.Fatalf("cannot call preSetupAddTenantUser with TenantMemberEntity with no tenant id")
			return false
		}

		if entity.Role != "tenant_admin" {
			t.Fatalf("cannot call preSetupAddTenantUser with TenantMemberEntity with role != \"tenant_user\"")
			return false
		}

		localPreSetup, createdUserId := preSetupAddTenantMember_ReturnUserId(t, &entity)
		ok := localPreSetup(deps)
		t.Logf("ok: %v (#%v)", ok, *createdUserId)
		if ok && setPath {
			(*tc).Path = fmt.Sprintf("/api/v1/tenant/%v/tenant_admin/%d", entity.TenantId, *createdUserId)
		}

		return ok
	}
}

/*
Ritorna il pre setup per aggiungere il Super Admin alla tabella tenant_members del proprio tenant. Se setPath è impostato a true, allora
il Path di tc viene impostato al path per operazioni CRUD su tale Super Admin.

NOTA: Se si mettono più di questi preSetup con setPath = truein un TC di eliminazione, quest'ultimo
controllerà l'eliminazione del tenant admin associato all'ULTIMO di questi preSetup nella lista
*/
func PreSetupAddSuperAdmin(
	t *testing.T,
	tc *helper.IntegrationTestCase,
	/*
		NOTA: Importantissimo che entity sia passato come valore e non come puntatore!
		In modo tale da copiarlo per valore e far funzionare l'ID autoincrementale
	*/
	entity user.SuperAdminEntity,
	setPath bool,
) helper.IntegrationTestPreSetup {
	t.Helper()
	return func(deps helper.IntegrationTestDeps) bool {
		localPreSetup, createdUserId := PreSetupAddSuperAdmin_ReturnUserId(t, &entity)
		ok := localPreSetup(deps)
		if ok && setPath {
			tc.Path = fmt.Sprintf("/api/v1/super_admin/%d", *createdUserId)
		}

		return ok
	}
}

/*
Ritorna il pre setup per aggiungere super admin e un puntatore al valore che conterrà l'id dell'utente creato, una
volta chiamato il setup ritornato.
*/
func preSetupAddTenantMember_ReturnUserId(t *testing.T, entity *user.TenantMemberEntity) (helper.IntegrationTestPreSetup, *uint) {
	t.Helper()
	createdUserId := new(uint)

	return func(deps helper.IntegrationTestDeps) bool {
		db := (*gorm.DB)(deps.CloudDB)
		// ensure super_admins table migrated
		if err := db.AutoMigrate(&user.TenantMemberEntity{}); err != nil {
			return false
		}
		// entity := user.SuperAdminEntity{Email: existingEmail, Name: "To Delete"}
		if err := db.
			Scopes(clouddb.WithTenantSchema(entity.TenantId, &user.TenantMemberEntity{})).
			Create(entity).Error; err != nil {
			return false
		}
		*createdUserId = entity.ID
		return true
	}, createdUserId
}

/*
Ritorna il pre setup per aggiungere super admin e un puntatore al valore che conterrà l'id dell'utente creato, una
volta chiamato il setup ritornato.
*/
func PreSetupAddSuperAdmin_ReturnUserId(t *testing.T, entity *user.SuperAdminEntity) (helper.IntegrationTestPreSetup, *uint) {
	t.Helper()
	createdUserId := new(uint)

	return func(deps helper.IntegrationTestDeps) bool {
		db := (*gorm.DB)(deps.CloudDB)
		// ensure super_admins table migrated
		if err := db.AutoMigrate(&user.SuperAdminEntity{}); err != nil {
			return false
		}
		// entity := user.SuperAdminEntity{Email: existingEmail, Name: "To Delete"}
		if err := db.Create(entity).Error; err != nil {
			return false
		}
		*createdUserId = entity.ID
		return true
	}, createdUserId
}
