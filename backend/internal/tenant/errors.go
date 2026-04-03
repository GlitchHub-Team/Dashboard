package tenant

import "errors"

var ErrTenantNotFound = errors.New("Tenant not found")

var ErrImpersonationFailded = errors.New("impersonation failed: Tenant must have impersonation enabled")

var ErrUnauthorized = errors.New("unauthorized: You do not have permission to perform this action")

var ErrInvalidTenantID = errors.New("invalid tenant ID: Tenant ID must be a valid UUID")

var ErrGetListPort = errors.New("unexpected error from GetTenants port")

var ErrTenantAlreadyExists = errors.New("tenant already exists")
