package tenant

import (
	"github.com/google/uuid"
)

type Tenant struct {
	Id             uuid.UUID
	Name           string
	CanImpersonate bool
}

func (t Tenant) IsZero() bool {
	return t == (Tenant{})
}

func (t *Tenant) GetId() uuid.UUID { return t.Id }
