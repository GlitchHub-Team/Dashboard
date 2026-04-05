package helper

import (
	"fmt"
)

/*
Ritorna la sezione di JSON di errore di una richiesta HTTP del controller per l'errore err
*/
func ErrJsonString(err error) string {
	return fmt.Sprintf(`"error":"%v"`, err.Error())
}
