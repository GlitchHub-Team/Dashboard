package gateway_connection_test

import (
	"errors"
	"testing"

	"backend/internal/gateway_connection"
	gatewayMocks "backend/tests/gateway/mocks"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGatewayHelloService_ProcessHello(t *testing.T) {
	cases := []struct {
		name          string
		input         gateway_connection.GatewayHelloMessage
		setupMock     func(mock *gatewayMocks.MockSaveGatewayPort)
		expectedError bool
		expectedCall  bool
	}{
		{
			name: "Success: valid message",
			input: gateway_connection.GatewayHelloMessage{
				GatewayId:        uuid.New().String(),
				PublicIdentifier: "pub-123",
			},
			setupMock: func(mock *gatewayMocks.MockSaveGatewayPort) {
				mock.EXPECT().Save(gomock.Any()).Return(nil).Times(1)
			},
			expectedError: false,
			expectedCall:  true,
		},
		{
			name: "Error: save fails",
			input: gateway_connection.GatewayHelloMessage{
				GatewayId:        uuid.New().String(),
				PublicIdentifier: "pub-456",
			},
			setupMock: func(mock *gatewayMocks.MockSaveGatewayPort) {
				mock.EXPECT().Save(gomock.Any()).Return(errors.New("save failed")).Times(1)
			},
			expectedError: true,
			expectedCall:  true,
		},
		{
			name: "Error: invalid UUID",
			input: gateway_connection.GatewayHelloMessage{
				GatewayId:        "invalid-uuid",
				PublicIdentifier: "pub-789",
			},
			setupMock: func(mock *gatewayMocks.MockSaveGatewayPort) {
				// No call expected
			},
			expectedError: true,
			expectedCall:  false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mockController := gomock.NewController(t)
			defer mockController.Finish()

			mockSavePort := gatewayMocks.NewMockSaveGatewayPort(mockController)
			tc.setupMock(mockSavePort)

			service := gateway_connection.NewGatewayHelloService(mockSavePort)

			err := service.ProcessHello(tc.input)

			if tc.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
