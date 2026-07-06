package user

import "time"

type User struct {
	ID           int64      `db:"id"`
	Login        string     `db:"login"`
	Email        string     `db:"email"`
	PasswordHash string     `db:"password_hash"`
	FirstName    string     `db:"first_name"`
	LastName     string     `db:"last_name"`
	CreatedAt    time.Time  `db:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at"`
	DeletedAt    *time.Time `db:"deleted_at"`
}

func (u *User) IsDeleted() bool {
	return u.DeletedAt != nil
}

func (u *User) Delete(now time.Time) {
	u.DeletedAt = &now
	u.UpdatedAt = now
}

func (u *User) Restore(now time.Time) {
	u.DeletedAt = nil
	u.UpdatedAt = now
}
