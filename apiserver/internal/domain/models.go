package domain

type User struct {
	Id           UserId
	Name         string
	Email        string
	Salt         []byte
	PasswordHash []byte
}
