package storage

import (
	"auth/internal/models"
	"database/sql"
	"fmt"
	"go.uber.org/zap"
)

type Store interface {
	Ping() error
	Close() error
	CreateUser(user *models.User) (*models.User, error)
	GetUserById(id int, withPwd bool) (*models.User, error)
	GetUserByIdInternal(id int) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	SaveUserSession(token models.UserSession) error
	GetUserSession(userID int) (models.UserSession, error)
	UpdateUserSession(token models.UserSession) error
	GetUserSessionBySessionID(sessionID string) (models.UserSession, error)
	GetAllUsers() ([]models.User, error)
	UpdateUser(user *models.User) error
	DeleteUser(id int) error
	GetAllRoles() ([]models.Role, error)
	CreateRole(role models.Role) (models.Role, error)
	UpdateRole(role models.Role) error
	DeleteRole(id int) error
	GetUsersCount() int
	GetActiveUsersCount() int
	GetLast24hRegisteredCount() int
	UpdateUserPassword(userID int, hashedPassword string) error
}
type FirestoreStorage struct {
	config string
}
type PostgresStorage struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewPostgresStorage(db *sql.DB, logger *zap.Logger) *PostgresStorage {
	return &PostgresStorage{db, logger}
}
func (p *PostgresStorage) Ping() error {
	return p.db.Ping()
}

func (p *PostgresStorage) Close() error {
	return p.db.Close()
}

func (p *PostgresStorage) CreateUser(user *models.User) (*models.User, error) {
	var userCreated models.User
	err := p.db.QueryRow("INSERT INTO users (first_name, last_name, email, pwd, google_id) VALUES ($1, $2, $3, $4, $5) RETURNING id, email, first_name, last_name, google_id, role", user.FirstName, user.LastName, user.Email, user.Password, user.GoogleID).Scan(&userCreated.ID, &userCreated.Email, &userCreated.FirstName, &userCreated.LastName, &userCreated.GoogleID, &userCreated.Role)
	if err != nil {
		return nil, err
	}
	return &userCreated, nil
}

func (p *PostgresStorage) GetUserById(id int, withPwd bool) (*models.User, error) {
	var user models.User
	if withPwd {
		err := p.db.QueryRow("SELECT id, first_name, last_name, email, pwd, role, google_id, verified, created_at FROM users WHERE id = $1", id).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Password, &user.Role, &user.GoogleID, &user.Verified, &user.CreatedAt)
		if err != nil {
			return nil, err
		}
	} else {
		err := p.db.QueryRow("SELECT id, first_name, last_name, email, role, google_id, verified, created_at FROM users WHERE id = $1", id).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Role, &user.GoogleID, &user.Verified, &user.CreatedAt)
		if err != nil {
			return nil, err
		}
	}
	return &user, nil
}
func (p *PostgresStorage) GetUserByIdInternal(id int) (*models.User, error) {
	var user models.User
	err := p.db.QueryRow("SELECT id, email, pwd, created_at, first_name, last_name, role, google_id, verified FROM users WHERE id = $1", id).Scan(&user.ID, &user.Email, &user.Password, &user.CreatedAt, &user.FirstName, &user.LastName, &user.Role, &user.GoogleID, &user.Verified)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (p *PostgresStorage) GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	err := p.db.QueryRow("SELECT id, first_name, last_name, email, pwd, role, google_id, verified FROM users WHERE email = $1", email).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Password, &user.Role, &user.GoogleID, &user.Verified)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (p *PostgresStorage) SaveUserSession(session models.UserSession) error {
	_, err := p.db.Exec("INSERT INTO users_sessions (user_id, session_id, expiration) VALUES ($1, $2, $3)", session.UserID, session.SessionID, session.Expiration)
	return err
}
func (p *PostgresStorage) UpdateUserSession(session models.UserSession) error {
	_, err := p.db.Exec("UPDATE users_sessions SET session_id = $1, expiration = $2 WHERE user_id = $3", session.SessionID, session.Expiration, session.UserID)
	return err
}
func (p *PostgresStorage) GetUserSession(userID int) (models.UserSession, error) {
	var session models.UserSession
	err := p.db.QueryRow("SELECT user_id, session_id, expiration FROM users_sessions WHERE user_id = $1", userID).Scan(&session.UserID, &session.SessionID, &session.Expiration)
	return session, err
}
func (p *PostgresStorage) GetUserSessionBySessionID(sessionID string) (models.UserSession, error) {
	var session models.UserSession
	err := p.db.QueryRow("SELECT user_id, session_id, expiration FROM users_sessions WHERE session_id = $1", sessionID).Scan(&session.UserID, &session.SessionID, &session.Expiration)
	return session, err
}

func (p *PostgresStorage) GetAllUsers() ([]models.User, error) {
	rows, err := p.db.Query("SELECT id, email, first_name, last_name, role, google_id, created_at FROM users ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("error fetching users: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(&user.ID, &user.Email, &user.FirstName, &user.LastName, &user.Role, &user.GoogleID, &user.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("error scanning user: %w", err)
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error with rows iteration: %w", err)
	}

	return users, nil
}

func (p *PostgresStorage) UpdateUser(user *models.User) error {
	query := `
        UPDATE users 
        SET first_name = $1, last_name = $2, email = $3, role = $4, pwd = $5, google_id = $6, verified=$7, updated_at = NOW()
        WHERE id = $8
    `
	_, err := p.db.Exec(query, user.FirstName, user.LastName, user.Email, user.Role, user.Password, user.GoogleID, user.Verified, user.ID)
	if err != nil {
		return fmt.Errorf("error updating user: %w", err)
	}

	return nil
}

func (p *PostgresStorage) DeleteUser(id int) error {
	_, err := p.db.Exec("DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("error deleting user: %w", err)
	}

	return nil
}

func (p *PostgresStorage) GetAllRoles() ([]models.Role, error) {
	rows, err := p.db.Query("SELECT id, name FROM roles")
	if err != nil {
		return nil, fmt.Errorf("error fetching roles: %w", err)
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var role models.Role
		err := rows.Scan(&role.ID, &role.Name)
		if err != nil {
			return nil, fmt.Errorf("error scanning role: %w", err)
		}
		roles = append(roles, role)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error with rows iteration: %w", err)
	}

	return roles, nil
}

func (p *PostgresStorage) CreateRole(role models.Role) (models.Role, error) {
	query := "INSERT INTO roles (name) VALUES ($1) RETURNING id, name"
	var newRole models.Role
	err := p.db.QueryRow(query, role.Name).Scan(&newRole.ID, &newRole.Name)
	if err != nil {
		return models.Role{}, fmt.Errorf("error creating role: %w", err)
	}

	return newRole, nil
}

func (p *PostgresStorage) UpdateRole(role models.Role) error {
	query := "UPDATE roles SET name = $1 WHERE id = $2"
	_, err := p.db.Exec(query, role.Name, role.ID)
	if err != nil {
		return fmt.Errorf("error updating role: %w", err)
	}

	return nil
}

func (p *PostgresStorage) DeleteRole(id int) error {
	_, err := p.db.Exec("DELETE FROM roles WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("error deleting role: %w", err)
	}

	return nil
}

func (p *PostgresStorage) GetUsersCount() int {
	var count int
	err := p.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return 0
	}
	return count
}

func (p *PostgresStorage) GetActiveUsersCount() int {
	var count int
	err := p.db.QueryRow("SELECT COUNT(*) FROM users_sessions").Scan(&count)
	if err != nil {
		return 0
	}
	return count
}

func (p *PostgresStorage) GetLast24hRegisteredCount() int {
	var count int
	err := p.db.QueryRow("SELECT COUNT(*) FROM users WHERE created_at > NOW() - INTERVAL '24 hours'").Scan(&count)
	if err != nil {
		return 0
	}
	return count
}
func (p *PostgresStorage) UpdateUserPassword(userID int, hashedPassword string) error {
	_, err := p.db.Exec("UPDATE users SET pwd = $1 WHERE id = $2", hashedPassword, userID)
	return err
}