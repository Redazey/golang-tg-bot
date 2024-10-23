package bottypes

import "time"

type Users struct {
	ID          int64 `gorm:"primaryKey"`
	Access      bool
	Accessed_at time.Time
}
