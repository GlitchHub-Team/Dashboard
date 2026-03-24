package auth

import "time"

type ChangePasswordToken struct {
	id         int
	token      string
	userId     uint
	expiryDate time.Time
}

type ConfirmToken struct {
	id         int
	token      string
	userId     uint
	expiryDate time.Time
}
