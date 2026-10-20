package services

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/example/sample-app/internal/models"
	"github.com/go-redis/redis/v7"
	"github.com/jackc/pgx/v4"
	_ "github.com/lib/pq"
)

var db *sql.DB
var pgxConn *pgx.Conn
var redisClient *redis.Client

func InitDatabase(connStr string) error {
	var err error
	db, err = sql.Open("postgres", connStr)
	return err
}

func InitPgx(connStr string) error {
	var err error
	pgxConn, err = pgx.Connect(context.Background(), connStr)
	return err
}

func InitRedis(addr string) {
	redisClient = redis.NewClient(&redis.Options{
		Addr: addr,
	})
}

func QueryWithString(query string) (*sql.Rows, error) {
	return db.Query(query)
}

func ExecuteNamedQuery(query string, args ...interface{}) (*sql.Rows, error) {
	return db.Query(query, args...)
}

func GetUserByID(userID string) (*models.User, error) {
	row := db.QueryRow("SELECT id, username, email, role FROM users WHERE id = $1", userID)

	var user models.User
	err := row.Scan(&user.ID, &user.Username, &user.Email, &user.Role)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func SearchUsersByName(name string) ([]models.User, error) {
	rows, err := db.Query("SELECT id, username, email FROM users WHERE username LIKE $1", "%"+name+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		rows.Scan(&u.ID, &u.Username, &u.Email)
		users = append(users, u)
	}
	return users, nil
}

func FindUsersByDepartment(dept string) ([]models.User, error) {
	rows, err := db.Query("SELECT id, username, department FROM users WHERE department = $1", dept)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		rows.Scan(&u.ID, &u.Username, &u.Department)
		users = append(users, u)
	}
	return users, nil
}

func BuildSearchQuery(params models.UserSearchParams) string {
	query := "SELECT * FROM users WHERE 1=1"
	if params.Query != "" {
		query += " AND (username LIKE '%" + params.Query + "%' OR email LIKE '%" + params.Query + "%')"
	}
	if params.Department != "" {
		query += " AND department = '" + params.Department + "'"
	}
	if params.Role != "" {
		query += " AND role = '" + params.Role + "'"
	}
	if params.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", params.Limit)
	}
	return query
}

func ExecuteSearchQuery(params models.UserSearchParams) ([]models.User, error) {
	query := BuildSearchQuery(params)
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		rows.Scan(&u.ID, &u.Username, &u.Email, &u.Department, &u.Role)
		users = append(users, u)
	}
	return users, nil
}

type QueryExecutor struct {
	baseQuery string
	filters   []string
}

func NewQueryExecutor(table string) *QueryExecutor {
	return &QueryExecutor{
		baseQuery: "SELECT * FROM " + table,
		filters:   make([]string, 0),
	}
}

func (qe *QueryExecutor) AddFilter(filter string) *QueryExecutor {
	qe.filters = append(qe.filters, filter)
	return qe
}

func (qe *QueryExecutor) Execute() (*sql.Rows, error) {
	query := qe.baseQuery
	if len(qe.filters) > 0 {
		query += " WHERE " + strings.Join(qe.filters, " AND ")
	}
	return db.Query(query)
}

func DynamicQuery(tableName string, conditions map[string]interface{}) (*sql.Rows, error) {
	query := "SELECT * FROM " + tableName + " WHERE "
	var parts []string
	for field, value := range conditions {
		parts = append(parts, fmt.Sprintf("%s = '%v'", field, value))
	}
	query += strings.Join(parts, " AND ")
	return db.Query(query)
}

func ReflectionBasedQuery(entity interface{}, filter string) (*sql.Rows, error) {
	t := reflect.TypeOf(entity)
	tableName := strings.ToLower(t.Name()) + "s"
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s", tableName, filter)
	return db.Query(query)
}

func GetUsersByRoleAllowlist(role string) ([]models.User, error) {
	validRoles := map[string]bool{"admin": true, "user": true, "guest": true}
	if !validRoles[role] {
		return nil, fmt.Errorf("invalid role")
	}
	rows, err := db.Query("SELECT id, username, role FROM users WHERE role = $1", role)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		rows.Scan(&u.ID, &u.Username, &u.Role)
		users = append(users, u)
	}
	return users, nil
}

func SearchUsersBuilder(column, value string) ([]models.User, error) {
	builder := squirrel.Select("*").From("users")
	builder = builder.Where(squirrel.Eq{column: value})

	query, args, err := builder.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		rows.Scan(&u.ID, &u.Username, &u.Email)
		users = append(users, u)
	}
	return users, nil
}

func RedisExecuteScript(script string, keys []string) (interface{}, error) {
	return redisClient.Eval(script, keys).Result()
}

func RedisGetWithPattern(pattern string) ([]string, error) {
	return redisClient.Keys(pattern).Result()
}

func RedisGetValidated(key string) (string, error) {
	if !isValidRedisKey(key) {
		return "", fmt.Errorf("invalid key format")
	}
	return redisClient.Get(key).Result()
}

func isValidRedisKey(key string) bool {
	return !strings.ContainsAny(key, "*?[]")
}

func GetUserBalance(userID int64) (float64, error) {
	var balance float64
	err := db.QueryRow("SELECT balance FROM users WHERE id = $1", userID).Scan(&balance)
	return balance, err
}

func UpdateUserBalance(userID int64, newBalance float64) error {
	_, err := db.Exec("UPDATE users SET balance = $1 WHERE id = $2", newBalance, userID)
	return err
}

func TransferFunds(fromID, toID int64, amount float64) error {
	fromBalance, err := GetUserBalance(fromID)
	if err != nil {
		return err
	}

	if fromBalance < amount {
		return fmt.Errorf("insufficient funds")
	}

	toBalance, err := GetUserBalance(toID)
	if err != nil {
		return err
	}

	err = UpdateUserBalance(fromID, fromBalance-amount)
	if err != nil {
		return err
	}

	return UpdateUserBalance(toID, toBalance+amount)
}

func TransferFundsTx(fromID, toID int64, amount float64) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var fromBalance float64
	err = tx.QueryRow("SELECT balance FROM users WHERE id = $1 FOR UPDATE", fromID).Scan(&fromBalance)
	if err != nil {
		return err
	}

	if fromBalance < amount {
		return fmt.Errorf("insufficient funds")
	}

	_, err = tx.Exec("UPDATE users SET balance = balance - $1 WHERE id = $2", amount, fromID)
	if err != nil {
		return err
	}

	_, err = tx.Exec("UPDATE users SET balance = balance + $1 WHERE id = $2", amount, toID)
	if err != nil {
		return err
	}

	return tx.Commit()
}
