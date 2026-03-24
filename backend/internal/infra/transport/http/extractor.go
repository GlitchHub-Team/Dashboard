package http

import (
	"errors"

	"backend/internal/shared/identity"

	"github.com/gin-gonic/gin"
)

var ErrMissingIdentity = errors.New("identity missing from request context")

// TODO: inserisci questo in tutti i controller!
/*
Estrare il requester dal contesto di Gin.
L'unico errore che ritorna è ErrMissingIdentity
*/
func ExtractRequester(ctx *gin.Context) (identity.Requester, error) {
	ctxRequester, exists := ctx.Get("requester")
	if !exists {
		return identity.Requester{}, ErrMissingIdentity
	}
	requester, ok := ctxRequester.(identity.Requester)
	if !ok {
		return identity.Requester{}, ErrMissingIdentity
	}

	return requester, nil
}
