package auth

type User struct {
	UserID   string `db:"user_id"`
	Name     string `db:"name"`
	Surname  string `db:"surname"`
	Password string `db:"password"`
	Login    string `db:"login"`
}
