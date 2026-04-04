package dto

type Pagination struct {
	// Page: pagina di dati (iniziando a contare da 1)
	Page int `uri:"page" form:"page" json:"page" binding:"min=1"`

	// Limit: quanti elementi inserire in una pagina (minimo: 1, massimo: 200)
	Limit int `uri:"limit" form:"limit" json:"limit" binding:"min=1,max=200"`
}

var DEFAULT_PAGINATION = Pagination{Page: 1, Limit: 25}

type ListInfo struct {
	Count uint `uri:"count" form:"count" json:"count" binding:"required"`
	Total uint `uri:"total" form:"total" json:"total" binding:"required"`
}

