package helper

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	clouddb "backend/internal/infra/database/cloud_db/connection"
	sensordb "backend/internal/infra/database/sensor_db"
	"backend/internal/infra/modules"
	natsutils "backend/internal/infra/nats"
	"backend/internal/shared/config"
	sharedCrypto "backend/internal/shared/crypto"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/fx"
)
/*
Oggetto che contiene tutte le possibili dipendenze di un test d'integrazione, compreso
il context della richiesta e il router.

NOTA: Ogni parametro che viene aggiunto deve essere puntatore a un tipo concreto o interfaccia.
*/
type IntegrationTestDeps struct {
	Ctx context.Context

	Router *gin.Engine

	CloudDB  clouddb.CloudDBConnection
	SensorDB sensordb.SensorDBConnection

	NatsConn     *nats.Conn
	NatsTestConn natsutils.NatsTestConnection

	JetStreamCtx     jetstream.JetStream
	JetStreamTestCtx jetstream.JetStream

	AuthTokenManager sharedCrypto.AuthTokenManager
}

type IntegrationTestPreSetup func(deps IntegrationTestDeps) bool

type IntegrationTestPostSetup func(deps IntegrationTestDeps)

type IntegrationTestCheck func(
	r *httptest.ResponseRecorder, deps IntegrationTestDeps,
) bool

type IntegrationTestCase struct {
	/*
		Insieme di funzioni da eseguire in ordine PRIMA di eseguire il test di integrazione.

		ATTENZIONE: ciascun preSetup chiama usando defer il postSetup nello stesso indice nel campo PostSetups.
		Per cui se si hanno 3 coppie pre-post, allora il test eseguirà le seguenti chiamate in ordine:

		pre 1, pre 2, pre 3, CORPO TEST, post 3, post 2, post 1
	*/
	PreSetups []IntegrationTestPreSetup

	Name   string      // Nome del test
	Method string      // Metodo HTTP da usare nel router
	Path   string      // URL da usare nel router
	Header http.Header // Header HTTP da usare nella richiesta al router
	Body   io.Reader   // Corpo della richiesta HTTP

	WantStatusCode   int    // Status code HTTP atteso
	WantResponseBody string // Matching della risposta

	/*
		Controlli da eseguire dopo che il controller ha inviato la risposta al client
	*/
	ResponseChecks []IntegrationTestCheck

	/*
		Insieme di funzioni da eseguire DOPO aver eseguito il test di integrazione.

		ATTENZIONE: ciascun preSetup chiama usando defer il postSetup nello stesso indice nel campo PostSetups.
	*/
	PostSetups []IntegrationTestPostSetup
}

func SetupIntegrationTest(t *testing.T) IntegrationTestDeps {
	if testing.Short() {
		t.Skip("integration test skipped in short mode")
	}

	var router *gin.Engine
	var cloudDB clouddb.CloudDBConnection
	var sensorDB sensordb.SensorDBConnection
	var natsConn *nats.Conn
	var natsTestConn natsutils.NatsTestConnection
	var jetstreamCtx jetstream.JetStream
	var jwtManager sharedCrypto.AuthTokenManager

	app := fx.New(
		modules.AppModules(),
		fx.Populate(&router),
		fx.Populate(&cloudDB),
		fx.Populate(&sensorDB),
		fx.Populate(&natsConn),
		fx.Populate(&natsTestConn),
		fx.Populate(&jetstreamCtx),
		fx.Populate(&jwtManager),

		// Imposta configurazione di test per i database a prescindere dall'env
		fx.Decorate(func(cfg *config.Config) *config.Config {
			cfg.CloudDBTest = true
			cfg.SensorDBTest = true
			return cfg
		}),

		fx.Replace(
			natsutils.NatsCredsPath("../../../"+os.Getenv("DASHBOARD_CREDS_PATH")),
			natsutils.NatsTestCredsPath("../../../"+os.Getenv("TEST_CREDS_PATH")),
			natsutils.NatsCAPemPath("../../../"+os.Getenv("CA_PEM_PATH")),
		),

		fx.NopLogger,
	)

	ctx := context.Background()
	if err := app.Start(ctx); err != nil {
		t.Fatalf("Failed to start Fx app for test: %v", err)
	}
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		app.Stop(stopCtx) //nolint:errcheck
	}()

	jetstreamTestCtx, err := natsutils.NewJetStreamContext(natsTestConn)
	if err != nil {
		t.Fatalf("Failed to create JetStream context for test connection: %v", err)
	}

	return IntegrationTestDeps{
		Ctx:              ctx,
		Router:           router,
		CloudDB:          cloudDB,
		SensorDB:         sensorDB,
		NatsConn:         natsConn,
		NatsTestConn:     natsTestConn,
		JetStreamCtx:     jetstreamCtx,
		JetStreamTestCtx: jetstreamTestCtx,
		AuthTokenManager: jwtManager,
	}
	// return router, cloudDB, sensorDB, natsConn, natsTestConn, jetstreamCtx, jetstreamTestCtx, jwtManager, ctx
}

func RunIntegrationTests(
	t *testing.T,
	tests []*IntegrationTestCase,
	deps IntegrationTestDeps,
) {
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			if len(tc.PreSetups) != len(tc.PostSetups) {
				t.Fatalf("len(PreSetups) must be == len(PostSetups) for test case %s", tc.Name)
			}

			for index, preSetup := range tc.PreSetups {
				// if !preSetup(clouddb, sensordb, natsConn, natsTestConn, jetstreamCtx, jetstreamTestCtx) {
				if preSetup != nil && !preSetup(deps) {
					t.Errorf("Pre-setup failed for test case %s", tc.Name)
				}

				postSetup := tc.PostSetups[index]
				if postSetup != nil {
					defer postSetup(deps)
				}
			}

			w := httptest.NewRecorder()

			req, err := http.NewRequestWithContext(deps.Ctx, tc.Method, tc.Path, tc.Body)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			req.Header.Set("Content-Type", "application/json")
			for key, values := range tc.Header {
				for _, value := range values {
					req.Header.Add(key, value)
				}
			}

			deps.Router.ServeHTTP(w, req)

			if w.Code != tc.WantStatusCode {
				t.Errorf("Expected status code %d, got %d. Body: %s", tc.WantStatusCode, w.Code, w.Body.String())
			}

			if tc.WantResponseBody != "" {
				if !strings.Contains(w.Body.String(), tc.WantResponseBody) {
					t.Errorf("Expected body with %s, got %s", tc.WantResponseBody, w.Body.String())
				}
			}

			for i, responseCheck := range tc.ResponseChecks {
				if !responseCheck(w, deps) {
					t.Errorf("Response check at index %v failed for test case %s", i, tc.Name)
				}
			}
		})
	}
}
