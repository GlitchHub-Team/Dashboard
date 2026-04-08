package dto

import "time"

type TimestampField struct {
	Timestamp time.Time `uri:"timestamp" form:"timestamp" json:"timestamp" binding:"required"`
}
