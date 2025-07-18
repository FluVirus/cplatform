package postgres

type UserDto struct {
	Id           int64  `db:"id"`
	Name         string `db:"name"`
	Email        string `db:"email"`
	PasswordHash []byte `db:"password_hash"`
	Salt         []byte `db:"salt"`
}
