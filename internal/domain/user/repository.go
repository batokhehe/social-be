package user

import (
	"database/sql"
)

type Repository struct {
	DB *sql.DB
}

func (r *Repository) Create(name, email, password string) error {
	_, err := r.DB.Exec(`
		INSERT INTO users (name, email, password_hash)
		VALUES ($1, $2, $3)
	`, name, email, password)

	return err
}

func (r *Repository) GetAll() ([]User, error) {
	rows, err := r.DB.Query(`
		SELECT id, name, email 
		FROM users 
		WHERE deleted_at IS NULL
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User

	for rows.Next() {
		var u User
		err := rows.Scan(&u.ID, &u.Name, &u.Email)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}

	return users, nil
}

func (r *Repository) GetByEmail(email string) (*User, string, int, error) {
	var user User
	var passwordHash string
	var role int

	err := r.DB.QueryRow(`
		SELECT id, name, email, password_hash, role
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`, email).Scan(&user.ID, &user.Name, &user.Email, &passwordHash, &role)

	if err != nil {
		return nil, "", 0, err
	}

	return &user, passwordHash, role, nil
}

func (r *Repository) GetByID(id int) (*User, error) {
	var user User

	err := r.DB.QueryRow(`
		SELECT id, name, email
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`, id).Scan(&user.ID, &user.Name, &user.Email)

	if err != nil {
		return nil, err
	}

	return &user, nil
}
