package project

import "time"

type Project struct {
	ID          int64      `db:"id"`
	UserID      int64      `db:"user_id"`
	Name        string     `db:"name"`
	Description string     `db:"description"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
	DeletedAt   *time.Time `db:"deleted_at"`
}

func (p Project) IsDeleted() bool {
	return p.DeletedAt != nil
}
