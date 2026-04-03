package middlewares

import (
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

	// Imposta il context
	ctx.Set("requester", requester)
	ctx.Next()
}
