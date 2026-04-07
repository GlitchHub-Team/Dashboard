package crypto

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"backend/internal/shared/config"
	sharedCrypto "backend/internal/shared/crypto"
	"backend/internal/shared/identity"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

// NOTA: i nomi delle chiavi devono essere corti per rendere il JWT più piccolo possibile
type jwtObj struct {
	Expiry   int64  `json:"exp"` // Unix timestamp (in SECONDI) della data di scadenza
	TenantId string `json:"tid"` // Se vuoto, allora non vi è tenant id (caso super admin)
	UserId   uint   `json:"uid"` // Id utente
	Role     string `json:"rol"` // Super Admin: "sa", Tenant Admin: "ta", Tenant User: "tu"
}

func (token *jwtObj) ToClaims() (jwt.MapClaims, error) {
	bytes, err := json.Marshal(token)
	if err != nil {
		return jwt.MapClaims{}, err
	}

	var claims jwt.MapClaims
	err = json.Unmarshal(bytes, &claims)
	if err != nil {
		return jwt.MapClaims{}, err
	}

	return claims, nil
}

func jwtTokenFromClaims(claims jwt.MapClaims) (jwtObj, error) {
	bytes, err := json.Marshal(claims)
	if err != nil {
		return jwtObj{}, err
	}

	var token jwtObj
	err = json.Unmarshal(bytes, &token)
	if err != nil {
		return jwtObj{}, err
	}
	return token, nil
}

type JWTManager struct {
	secret        []byte
	tokenDuration time.Duration
}

var ErrInvalidSigningMethod = errors.New("invalid signing method")

var _ sharedCrypto.AuthTokenManager = (*JWTManager)(nil)

func NewJWTManager(
	cfg *config.Config,
) (*JWTManager, error) {
	encodedSecret := []byte(cfg.AuthTokenSecret)

	secret := make([]byte, base64.RawURLEncoding.DecodedLen(len(encodedSecret)))
	_, err := base64.RawURLEncoding.Decode(secret, encodedSecret)
	if err != nil {
		return nil, err
	}
	return &JWTManager{
		secret:        secret,
		tokenDuration: time.Second * time.Duration(cfg.AuthTokenDuration),
	}, nil
}

const (
	JWT_SUPER_ADMIN  = "sa"
	JWT_TENANT_ADMIN = "ta"
	JWT_TENANT_USER  = "tu"
)

func (JWTManager) userRoleToString(role identity.UserRole) (
	roleString string, err error,
) {
	switch role {
	case identity.ROLE_SUPER_ADMIN:
		roleString = JWT_SUPER_ADMIN
	case identity.ROLE_TENANT_ADMIN:
		roleString = JWT_TENANT_ADMIN
	case identity.ROLE_TENANT_USER:
		roleString = JWT_TENANT_USER
	default:
		err = identity.ErrUnknownRole
	}
	return
}

func (JWTManager) stringToUserRole(roleString string) (
	role identity.UserRole, err error,
) {
	switch roleString {
	case JWT_SUPER_ADMIN:
		role = identity.ROLE_SUPER_ADMIN
	case JWT_TENANT_ADMIN:
		role = identity.ROLE_TENANT_ADMIN
	case JWT_TENANT_USER:
		role = identity.ROLE_TENANT_USER
	default:
		err = identity.ErrUnknownRole
	}
	return
}

func (generator *JWTManager) GenerateForRequester(requester identity.Requester) (string, error) {
	// jwtClaims := jwt.MapClaims{}
	tokenObj := jwtObj{}

	// 1. Check requester user id
	if requester.RequesterUserId == 0 {
		return "", identity.ErrInvalidUser
	}
	tokenObj.UserId = requester.RequesterUserId

	// 2. Check requester tenant id
	switch requester.RequesterRole {
	case identity.ROLE_SUPER_ADMIN:
		tokenObj.TenantId = ""
	case identity.ROLE_TENANT_ADMIN, identity.ROLE_TENANT_USER:
		tokenObj.TenantId = requester.RequesterTenantId.String()
	default:
		return "", fmt.Errorf("%v: '%v'", identity.ErrUnknownRole, requester.RequesterRole)
	}

	// 3. Check roleString
	roleString, err := generator.userRoleToString(requester.RequesterRole)
	if err != nil {
		return "", err
	}
	tokenObj.Role = roleString

	// 4. Imposta scadenza
	expiryDate := time.Now().Add(generator.tokenDuration)
	tokenObj.Expiry = expiryDate.Unix()

	// 5. Crea e firma token
	claims, err := tokenObj.ToClaims()
	if err != nil {
		return "", err
	}
	generateToken := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	token, err := generateToken.SignedString(generator.secret)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (generator *JWTManager) GetRequesterFromToken(tokenString string) (identity.Requester, error) {
	// 1. Parse token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidSigningMethod
		}
		return generator.secret, nil
	})
	if err != nil {
		return identity.Requester{}, err
	}

	// 2. Check validità token
	if !token.Valid {
		return identity.Requester{}, sharedCrypto.ErrInvalidAuthToken
	}

	// 3. Imposta oggetto di tipo jwtToken
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return identity.Requester{}, sharedCrypto.ErrInvalidAuthToken
	}

	tokenObj, err := jwtTokenFromClaims(claims)
	if err != nil {
		return identity.Requester{}, nil
	}

	// 4. Check token
	// - Scadenza
	if time.Now().Unix() > tokenObj.Expiry {
		return identity.Requester{}, sharedCrypto.ErrInvalidAuthToken
	}
	// - User Id
	if tokenObj.UserId == 0 {
		return identity.Requester{}, sharedCrypto.ErrInvalidAuthToken
	}
	// - Role
	role, err := generator.stringToUserRole(tokenObj.Role)
	if err != nil {
		return identity.Requester{}, sharedCrypto.ErrInvalidAuthToken
	}
	// - Tenant Id
	if role != identity.ROLE_SUPER_ADMIN && tokenObj.TenantId == "" {
		return identity.Requester{}, sharedCrypto.ErrInvalidAuthToken
	}

	// 5. Crea requester
	var tenantId *uuid.UUID // nil di default
	if role != identity.ROLE_SUPER_ADMIN {
		parsedUuid, err := uuid.Parse(tokenObj.TenantId)
		if err != nil {
			return identity.Requester{}, sharedCrypto.ErrInvalidAuthToken
		}
		tenantId = &parsedUuid
	}

	requester := identity.Requester{
		RequesterUserId:   tokenObj.UserId,
		RequesterTenantId: tenantId,
		RequesterRole:     role,
	}

	return requester, nil
}
