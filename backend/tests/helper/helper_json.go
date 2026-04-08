package helper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
)

/*
Ritorna la sezione di JSON di errore di una richiesta HTTP del controller per l'errore err
*/
func ErrJsonString(err error) string {
	return fmt.Sprintf(`"error":"%v"`, err.Error())
}

func MustJSONBody(t *testing.T, payload any) *bytes.Reader {
	t.Helper()

	jsonBody, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("failed to marshal json payload: %v", err)
	}
	return bytes.NewReader(jsonBody)
}
