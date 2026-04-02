package dto

/*
Campi DTO per identificare un token: il token stesso e il potenziale tenant Id a cui appartiene (campo
nullabile)
*/
type TokenFields struct {
	Token string `uri:"token" form:"token" json:"token" binding:"required"`
	TenantIdField_NotRequired
}

/* Campo DTO da usare per verificare password di un utente */
type PasswordField struct {
	Password string `uri:"password" form:"password" json:"password" binding:"required,min=8"`
}

/*
Campo DTO da usare quando si vuole inserire la nuova password per qualcosa che prima non aveva password.
Per inserire campi DTO per cambio password (old e new), usare ChangePasswordFields
*/
type NewPasswordField struct {
	NewPassword string `uri:"new_password" form:"new_password" json:"new_password" binding:"required,min=8"`
}

/* Campi DTO da usare quando si vuole inserire nuova password per qualcosa che prima ne aveva un'altra */
type ChangePasswordFields struct {
	OldPassword string `uri:"old_password" form:"old_password" json:"old_password" binding:"required"`
	NewPassword string `uri:"new_password" form:"new_password" json:"new_password" binding:"required,min=8,nefield=OldPassword"`
}
