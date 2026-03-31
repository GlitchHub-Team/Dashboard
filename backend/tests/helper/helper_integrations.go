package helper

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	clouddb "backend/internal/infra/database/cloud_db/connection"
	sensordb "backend/internal/infra/database/sensor_db"
	"backend/internal/infra/modules"
	natsutils "backend/internal/infra/nats"

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
	Body   io.Reader

	WantStatusCode   int
	WantResponseBody string
	Checks           []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream) bool

	PostSetups []func(clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream)
}

func Setup(t *testing.T) (*gin.Engine, clouddb.CloudDBConnection, sensordb.SensorDBConnection, *nats.Conn, natsutils.NatsTestConnection, jetstream.JetStream, jetstream.JetStream, context.Context) {
	err := os.Chdir("../../../")
	if err != nil {
		t.Fatalf("Impossibile cambiare directory: %v", err)
	}

	if testing.Short() {
		t.Skip("integration test skipped in short mode")
	}

	var router *gin.Engine
	var cloudDB clouddb.CloudDBConnection
	var sensorDB sensordb.SensorDBConnection
	var natsConn *nats.Conn
	var natsTestConn natsutils.NatsTestConnection
	var jetstreamCtx jetstream.JetStream
	app := fx.New(
		modules.AppModules(),
		fx.Populate(&router),
		fx.Populate(&cloudDB),
		fx.Populate(&sensorDB),
		fx.Populate(&natsConn),
		fx.Populate(&natsTestConn),
		fx.Populate(&jetstreamCtx),
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

	return router, cloudDB, sensorDB, natsConn, natsTestConn, jetstreamCtx, jetstreamTestCtx, ctx
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
			for _, preSetup := range tt.PreSetups {
				if !preSetup(clouddb, sensordb, natsConn, natsTestConn, jetstreamCtx, jetstreamTestCtx) {
					t.Errorf("Pre-setup failed for test case %s", tt.Name)
				}
			}

			w := httptest.NewRecorder()

			path := "http://localhost:80/api/v1" + tt.Path

			req, err := http.NewRequestWithContext(ctx, tt.Method, path, tt.Body)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			req.Header.Set("Content-Type", "application/json")

			router.ServeHTTP(w, req)

			if w.Code != tt.WantStatusCode {
				t.Errorf("Expected status code %d, got %d. Body: %s", tt.WantStatusCode, w.Code, w.Body.String())
			}

			if tt.WantResponseBody != "" {
				if w.Body.String() != tt.WantResponseBody {
					t.Errorf("Expected body %s, got %s", tt.WantResponseBody, w.Body.String())
				}
			}

			for _, check := range tt.Checks {
				if !check(clouddb, sensordb, natsConn, natsTestConn, jetstreamCtx, jetstreamTestCtx) {
					t.Errorf("Check failed for test case %s", tt.Name)
				}
			}

			for _, postSetup := range tt.PostSetups {
				postSetup(clouddb, sensordb, natsConn, natsTestConn, jetstreamCtx, jetstreamTestCtx)
			}
		})
	}
}
