package dto

type CommandResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
