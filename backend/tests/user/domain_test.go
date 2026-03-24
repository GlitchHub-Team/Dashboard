package user_test

import (
	"errors"
	"testing"

	"backend/internal/user"
)

func TestUser_SetPasswordHash(t *testing.T) {
	t.Run(
		"set empty password",
		func(t *testing.T) {
			u := user.User{}
			err := u.SetPasswordHash("")
			if !errors.Is(err, user.ErrEmptyPassword) {
				t.Fatalf("want error '%v', got error '%v'", user.ErrEmptyPassword, err)
			}
		},
	)

	t.Run(
		"set same password",
		func(t *testing.T) {
			hash := "hash123"
			u := user.User{
				PasswordHash: &hash,
			}
			err := u.SetPasswordHash(hash)
			if !errors.Is(err, user.ErrSamePassword) {
				t.Fatalf("want error '%v', got error '%v'", user.ErrSamePassword, err)
			}
		},
	)

	t.Run(
		"set new password",
		func(t *testing.T) {
			oldHash := "hash123"
			newHash := "newHash67"
			u := user.User{
				PasswordHash: &oldHash,
			}
			err := u.SetPasswordHash(newHash)
			if err != nil {
				t.Fatalf("want no error, got error '%v'", err)
			}
		},
	)
}
