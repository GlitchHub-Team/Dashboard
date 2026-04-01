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
	sharedCrypto "backend/internal/shared/crypto"

	"github.com/gin-gonic/gin"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/fx"
)

type TestCase struct {
	PreSetups []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool

	Name   string
	Method string
	Path   string
	Header http.Header
	Body   io.Reader

	WantStatusCode   int
	WantResponseBody string
	ResponseChecks   []func(*httptest.ResponseRecorder, clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool

	PostSetups []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream)
}

func Setup(t *testing.T) (*gin.Engine, clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream, sharedCrypto.AuthTokenManager, context.Context) {
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

	return router, cloudDB, sensorDB, natsConn, natsTestConn, jetstreamCtx, jetstreamTestCtx, jwtManager, ctx
}

func RunTests(
	router *gin.Engine,
	ctx context.Context,
	tests []TestCase,
	t *testing.T,
	clouddb clouddb.CloudDBConnection,
	sensordb sensordb.SensorDBConnection,
	natsConn *nats.Conn,
	natsTestConn natsutils.NatsTestConnection,
	jetstreamCtx jetstream.JetStream,
	jetstreamTestCtx jetstream.JetStream,
) {
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			if len(tt.PreSetups) != len(tt.PostSetups) {
				t.Fatalf("Number of PreSetups and PostSetups must be the same for test case %s", tt.Name)
			}

			for index, preSetup := range tt.PreSetups {
				if !preSetup(clouddb, sensordb, natsConn, natsTestConn, jetstreamCtx, jetstreamTestCtx) {
					t.Errorf("Pre-setup failed for test case %s", tt.Name)
				}
				defer tt.PostSetups[index](clouddb, sensordb, natsConn, natsTestConn, jetstreamCtx, jetstreamTestCtx)
			}

			w := httptest.NewRecorder()

			req, err := http.NewRequestWithContext(ctx, tt.Method, tt.Path, tt.Body)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			req.Header.Set("Content-Type", "application/json")
			for key, values := range tt.Header {
				for _, value := range values {
					req.Header.Add(key, value)
				}
			}

			router.ServeHTTP(w, req)

			if w.Code != tt.WantStatusCode {
				t.Errorf("Expected status code %d, got %d. Body: %s", tt.WantStatusCode, w.Code, w.Body.String())
			}

			if tt.WantResponseBody != "" {
				if !strings.Contains(w.Body.String(), tt.WantResponseBody) {
					t.Errorf("Expected body with %s, got %s", tt.WantResponseBody, w.Body.String())
				}
			}

			for _, responseCheck := range tt.ResponseChecks {
				if !responseCheck(w, clouddb, sensordb, natsConn, natsTestConn, jetstreamCtx, jetstreamTestCtx) {
					t.Errorf("Response check failed for test case %s", tt.Name)
				}
			}
		})
	}
}
