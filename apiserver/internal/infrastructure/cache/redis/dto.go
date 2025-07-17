package redis

type UserDto struct {
	Id           int64  `redis:"id"`
	Name         string `redis:"name"`
	PasswordHash []byte `redis:"password_hash"`
	Salt         []byte `redis:"salt"`
}
