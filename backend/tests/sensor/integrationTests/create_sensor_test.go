package sensor_integration_test

import (
	"testing"

	"backend/tests/helper"
)

func TestSuccessfulSensorCreation(t *testing.T) {
	router, clouddb, sensordb, natsConn, natsTestConn, jetstreamCtx, jetstreamTestCtx, ctx := helper.Setup(t)
	tests := []helper.TestCase{
		{
			PreSetups: nil,
			Name:      "Create sensor successfully",
			Method:    "POST",
			Path:      "/sensor",
			Body:      nil, // TODO: add request body

			WantStatusCode:   401,
			WantResponseBody: "", // TODO: add expected response body
			Checks:           nil,

			PostSetups: nil,
		},
	}

	helper.RunTests(router, ctx, tests, t, clouddb, sensordb, natsConn, natsTestConn, jetstreamCtx, jetstreamTestCtx)
}
