package bottypes

import "time"

type Users struct {
	ID          int64 `gorm:"primaryKey"`
	Access      bool
	Is_added    bool
	Accessed_at time.Time
}

type Records struct {
	ID      int64 `gorm:"primaryKey"`
	User_id int64
	Amount  int64
	Status  bool
}
