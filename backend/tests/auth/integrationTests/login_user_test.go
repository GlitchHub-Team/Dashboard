package auth_integration_test

// import (
// 	"bytes"
// 	"crypto/sha512"
// 	"encoding/json"
// 	"fmt"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"

// 	"backend/internal/auth"
// 	"backend/internal/shared/identity"
// 	"backend/internal/tenant"
// 	"backend/internal/user"
// 	"backend/tests/helper"

// 	clouddb "backend/internal/infra/database/cloud_db/connection"
// 	"backend/internal/infra/transport/http/dto"

// 	"github.com/google/uuid"
// 	"golang.org/x/crypto/bcrypt"
// 	"gorm.io/gorm"
// )

// // helper: create tenant and migrate tenant_members schema
// func createTenantForTest(t *testing.T, deps helper.IntegrationTestDeps, tenantId uuid.UUID, canImpersonate bool) {
//     db := (*gorm.DB)(deps.CloudDB)
//     tenantEntity := tenant.TenantEntity{ID: tenantId.String(), Name: "t-login", CanImpersonate: canImpersonate}
//     if err := db.Clauses().Create(&tenantEntity).Error; err != nil {
//         t.Fatalf("cannot create tenant: %v", err)
//     }
//     schemaName := "tenant_" + tenantId.String()
//     if err := db.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS \"%s\"", schemaName)).Error; err != nil {
//         t.Fatalf("cannot create schema: %v", err)
//     }
//     if err := db.Transaction(func(tx *gorm.DB) error {
//         if err := tx.Exec(fmt.Sprintf("set local search_path to \"%s\"", schemaName)).Error; err != nil {
//             return err
//         }
//         return tx.AutoMigrate(&user.TenantMemberEntity{})
//     }); err != nil {
//         t.Fatalf("cannot migrate tenant_members: %v", err)
//     }
// }

// // helper: hash password using same pre-hash + bcrypt approach used in production
// func hashPasswordForTest(plaintext string) (string, error) {
//     pre := sha512.Sum512([]byte(plaintext))
//     h, err := bcrypt.GenerateFromPassword(pre[:], bcrypt.DefaultCost)
//     return string(h), err
// }

// func TestLoginUserIntegration(t *testing.T) {
//     deps := helper.SetupIntegrationTest(t)

//     // common values
//     tenantId := uuid.New()
//     email := "ilogin@domain.test"
//     password := "P@ssw0rd"

//     // prepare DB: tenant + user confirmed with password
//     createTenantForTest(t, deps, tenantId, true)
//     hashed, err := hashPasswordForTest(password)
//     if err != nil {
//         t.Fatalf("cannot hash password: %v", err)
//     }

//     tm := &user.TenantMemberEntity{TenantId: tenantId.String(), Email: email, Name: "ILogin", Password: &hashed, Confirmed: true, Role: string(identity.ROLE_TENANT_USER)}
//     db := (*gorm.DB)(deps.CloudDB)
//     if err := db.Scopes(clouddb.WithTenantSchema(tenantId.String(), &user.TenantMemberEntity{})).Create(tm).Error; err != nil {
//         t.Fatalf("cannot create tenant member: %v", err)
//     }

//     tests := []*helper.IntegrationTestCase{}

//     body := auth.LoginUserDTO{
// 		TenantIdField_NotRequired: dto.TenantIdField_NotRequired{
// 			TenantId: &tenantId,
// 		},
// 		Email: email, Password: password}

//     // Success: valid credentials
//     b, _ := json.Marshal(body)
//     tests = append(tests, &helper.IntegrationTestCase{
//         PreSetups: []helper.IntegrationTestPreSetup{},
//         Name:      "Success: valid credentials",
//         Method:    http.MethodPost,
//         Path:      "/api/v1/auth/login",
//         Header:    http.Header{},
//         Body:      helper.MustJSONBody(t, body),

//         WantStatusCode:   http.StatusOK,
//         WantResponseBody: "\"jwt\":",
//         ResponseChecks: []helper.IntegrationTestCheck{func(r *httptest.ResponseRecorder, deps helper.IntegrationTestDeps) bool {
//             // validate JWT returns a requester with correct user id
//             var resp map[string]any
//             if err := json.Unmarshal(r.Body.Bytes(), &resp); err != nil {
//                 t.Logf("invalid json: %v", err)
//                 return false
//             }
//             jwtStr, ok := resp["jwt"].(string)
//             if !ok || jwtStr == "" {
//                 t.Logf("missing jwt in response")
//                 return false
//             }
//             requester, err := deps.AuthTokenManager.GetRequesterFromToken(jwtStr)
//             if err != nil {
//                 t.Logf("invalid token: %v", err)
//                 return false
//             }
//             // requester user id must match created user id
//             if requester.RequesterUserId != tm.ID {
//                 t.Logf("token user id %v != created %v", requester.RequesterUserId, tm.ID)
//                 return false
//             }
//             return true
//         }},
//         PostSetups: []helper.IntegrationTestPostSetup{func(deps helper.IntegrationTestDeps) { // cleanup tenant
//             db := (*gorm.DB)(deps.CloudDB)
//             schemaName := "tenant_" + tenantId.String()
//             _ = db.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS \"%s\" CASCADE", schemaName)).Error
//             _ = db.Exec(fmt.Sprintf("DELETE FROM tenants WHERE id = '%s'", tenantId.String())).Error
//         }},
//     })

//     // Fail: binding JSON invalid
//     tests = append(tests, &helper.IntegrationTestCase{
//         PreSetups: []helper.IntegrationTestPreSetup{},
//         Name:      "Fail: binding JSON",
//         Method:    http.MethodPost,
//         Path:      "/api/v1/auth/login",
//         Header:    http.Header{},
//         Body:      bytes.NewBufferString("not-a-json"),

//         WantStatusCode:   http.StatusBadRequest,
//         WantResponseBody: "error",
//         ResponseChecks:   nil,
//         PostSetups:       []helper.IntegrationTestPostSetup{func(deps helper.IntegrationTestDeps) {}},
//     })

//     // Fail: account not confirmed
//     email2 := "unconfirmed@d.test"
//     hashed2, _ := hashPasswordForTest("pw2")
//     tm2 := &user.TenantMemberEntity{TenantId: tenantId.String(), Email: email2, Name: "U1", Password: &hashed2, Confirmed: false, Role: string(identity.ROLE_TENANT_USER)}
//     if err := db.Scopes(clouddb.WithTenantSchema(tenantId.String(), &user.TenantMemberEntity{})).Create(tm2).Error; err != nil {
//         t.Fatalf("cannot create unconfirmed user: %v", err)
//     }
//     body2 := auth.LoginUserDTO{TenantId: &tenantId, Email: email2, Password: "pw2"}
//     bb2, _ := json.Marshal(body2)
//     tests = append(tests, &helper.IntegrationTestCase{
//         PreSetups: []helper.IntegrationTestPreSetup{},
//         Name:      "Fail: account not confirmed",
//         Method:    http.MethodPost,
//         Path:      "/api/v1/auth/login",
//         Header:    http.Header{},
//         Body:      bytes.NewBuffer(bb2),

//         WantStatusCode:   http.StatusNotFound,
//         WantResponseBody: helper.ErrJsonString(auth.ErrAccountNotConfirmed),
//         ResponseChecks:   nil,
//         PostSetups: []helper.IntegrationTestPostSetup{func(deps helper.IntegrationTestDeps) {}},
//     })

//     // Fail: wrong credentials (email missing)
//     body3 := auth.LoginUserDTO{TenantId: &tenantId, Email: "missing@x.test", Password: "whatever"}
//     b3, _ := json.Marshal(body3)
//     tests = append(tests, &helper.IntegrationTestCase{
//         PreSetups: []helper.IntegrationTestPreSetup{},
//         Name:      "Fail: wrong credentials - email missing",
//         Method:    http.MethodPost,
//         Path:      "/api/v1/auth/login",
//         Header:    http.Header{},
//         Body:      bytes.NewBuffer(b3),

//         WantStatusCode:   http.StatusNotFound,
//         WantResponseBody: helper.ErrJsonString(auth.ErrWrongCredentials),
//         ResponseChecks:   nil,
//         PostSetups:       []helper.IntegrationTestPostSetup{func(deps helper.IntegrationTestDeps) {}},
//     })

//     // Fail: wrong password for existing email
//     body4 := auth.LoginUserDTO{TenantId: &tenantId, Email: email, Password: "badpwd"}
//     b4, _ := json.Marshal(body4)
//     tests = append(tests, &helper.IntegrationTestCase{
//         PreSetups: []helper.IntegrationTestPreSetup{},
//         Name:      "Fail: wrong password",
//         Method:    http.MethodPost,
//         Path:      "/api/v1/auth/login",
//         Header:    http.Header{},
//         Body:      bytes.NewBuffer(b4),

//         WantStatusCode:   http.StatusNotFound,
//         WantResponseBody: helper.ErrJsonString(auth.ErrWrongCredentials),
//         ResponseChecks:   nil,
//         PostSetups:       []helper.IntegrationTestPostSetup{func(deps helper.IntegrationTestDeps) {}},
//     })

//     helper.RunIntegrationTests(t, tests, deps)
// }
