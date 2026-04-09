package middlewares

import (
	"fmt"
	"net/http"

	"backend/internal/shared/crypto"
	// "backend/internal/shared/identity"

	transportHttp "backend/internal/infra/transport/http"

	"github.com/gin-gonic/gin"
)

/*
Middleware gin di autorizzazione
*/
type AuthzMiddleware struct {
	authTokenManager crypto.AuthTokenManager
}

func NewAuthzMiddleware(authTokenManager crypto.AuthTokenManager) *AuthzMiddleware {
	return &AuthzMiddleware{
		authTokenManager: authTokenManager,
	}
}

/*
Imposta la pagina in modo tale da richiedere un token di autenticazione (JWT al momento)
*/
func (authz *AuthzMiddleware) RequireAuthToken(ctx *gin.Context) {
	authzHeader := ctx.GetHeader("Authorization")
	if authzHeader == "" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": transportHttp.ErrMissingIdentity.Error(),
		})
		return
	}

	// Estrai token da "Bearer <token>"
	tokenString := ""
	if len(authzHeader) > 7 && authzHeader[:7] == "Bearer " {
		tokenString = authzHeader[7:]
	} else {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": transportHttp.ErrMissingIdentity.Error()})
		return
	}

	// Ottieni requester dal token
	requester, err := authz.authTokenManager.GetRequesterFromToken(tokenString)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	if requester.RequesterUserId == 0 {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": transportHttp.ErrMissingIdentity.Error(),
		})
	}

	// Imposta il context
	ctx.Set("requester", requester)
	ctx.Next()
}

/*
Imposta la pagina in modo tale da richiedere il JWT come query string parameter (?jwt=...).

Questa funzione va utilizzata solamente nei casi in cui non è possibile inserire nella richiesta
l'header Authorization (ad esempio, nel caso di richieste WebSocket).

NOTA: Questa è una soluzione temporanea che si può migliorare usando un sistema a "ticket", in cui
il client fa richiesta a un endpoint (ad es. /ws/new-ticket) che restituisce un token one-time a vita breve
(detto "ticket") da usare solo nel contesto della richiesta, salvato in un store in-memory (Redis o memdb).
Tale token poi può essere inserito nella prossima richiesta websocket come query parameter
(ad es. all'endpoint /api/v1/sensor/.../real_time_data?ticket=<ticket>)
*/
func (authz *AuthzMiddleware) RequireAuthTokenInQuery(ctx *gin.Context) {
	tokenString := ctx.Query("jwt")

	fmt.Printf("tok: %v\n", tokenString)

	if tokenString == "" {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": transportHttp.ErrMissingIdentity.Error(),
		})
		return
	}

	requester, err := authz.authTokenManager.GetRequesterFromToken(tokenString)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	if requester.RequesterUserId == 0 {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": transportHttp.ErrMissingIdentity.Error(),
		})
	}

	ctx.Set("requester", requester)
	ctx.Next()
}
