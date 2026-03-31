package helper

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"backend/internal/shared/identity"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"go.uber.org/mock/gomock"
)

func SetupMockUseCase[MockT any](
	constructor func(*gomock.Controller) *MockT,
	setupSteps []MockUseCaseSetupFunc[MockT],
	t *testing.T,
) *MockT {
	ctrl := gomock.NewController(t)
	mockUseCase := constructor(ctrl)

	var expectedCalls []any
	for _, step := range setupSteps {
		if call := step(mockUseCase); call != nil {
			expectedCalls = append(expectedCalls, call)
		}
	}
	if len(expectedCalls) > 0 {
		gomock.InOrder(expectedCalls...)
	}
	return mockUseCase
}

func ExecuteControllerTest[InputT any, MockT any](
	t *testing.T,
	tc GenericControllerTestCase[InputT, MockT],
	mountMethod string,
	mountUrl string,
	controllerFunc func(*gin.Context),
) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	if validatorEngine, ok := binding.Validator.Engine().(*validator.Validate); ok {
		validatorEngine.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})
	}

	router.Use(func(ctx *gin.Context) {
		if !tc.OmitIdentity {
			ctx.Set("requester", tc.Requester)
		}
		ctx.Next()
	})

	router.Handle(mountMethod, mountUrl, controllerFunc)

	reqBody, err := json.Marshal(tc.InputDto)
	if err != nil {
		t.Fatalf("error marshaling request body: %v", err)
	}

	req, err := http.NewRequest(tc.Method, tc.Url, bytes.NewBuffer(reqBody)) //nolint:noctx
	if err != nil {
		t.Fatalf("error creating request: %v", err)
	}
	if reflect.ValueOf(tc.InputDto) != reflect.Zero(reflect.TypeFor[InputT]()) {
		req.Header.Set("Content-Type", "application/json")
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != tc.ExpectedStatus {
		t.Fatalf("expected status %d, got %d. Response: %s", tc.ExpectedStatus, w.Code, w.Body.String())
	}

	CheckHttpResponse(w, tc.ExpectedResponse, t)
}

type GenericControllerTestCase[InputT any, MockT any] struct {
	Name             string
	Method           string
	Url              string
	InputDto         InputT
	Requester        identity.Requester
	OmitIdentity     bool
	SetupSteps       []MockUseCaseSetupFunc[MockT]
	ExpectedStatus   int
	ExpectedResponse any
}

type HasError struct{}

type MockUseCaseSetupFunc[T any] func(*T) *gomock.Call

func CheckHttpResponse[T any](w *httptest.ResponseRecorder, expectedResponse T, t *testing.T) {
	if any(expectedResponse) == nil {
		return
	}

	if reflect.ValueOf(expectedResponse) == reflect.Zero(reflect.TypeFor[T]()) {
		return
	}

	actualBytes := w.Body.Bytes()
	expectedBytes, err := json.Marshal(expectedResponse)
	if err != nil {
		t.Fatalf("failed to marshal expected response: %v", err)
	}

	var actualObj any
	if err := json.Unmarshal(actualBytes, &actualObj); err != nil {
		t.Fatalf("failed to unmarshal actual response: %v. Body: %s", err, string(actualBytes))
	}

	var expectedObj any
	if err := json.Unmarshal(expectedBytes, &expectedObj); err != nil {
		t.Fatalf("failed to unmarshal expected response: %v", err)
	}

	_, onlyCheckError := any(expectedResponse).(HasError)
	if onlyCheckError {
		actualMap, ok := actualObj.(map[string]any)
		if !ok {
			t.Fatalf("expected JSON object, got %#v", actualObj)
		}
		if _, ok := actualMap["error"]; !ok {
			t.Fatalf("expected response containing 'error', got %#v", actualObj)
		}
		return
	}

	if !reflect.DeepEqual(expectedObj, actualObj) {
		t.Errorf("response body mismatch.\nExpected: %#v\nGot:      %#v", expectedObj, actualObj)
	}
}
