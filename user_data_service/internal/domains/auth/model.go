package auth

type User struct {
	UserID       string `db:"user_id"`
	Name         string `db:"name"`
	Surname      string `db:"surname"`
	PasswordHash string `db:"password_hash"`
	Login        string `db:"login"`
}
