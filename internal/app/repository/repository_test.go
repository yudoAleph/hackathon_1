package repository

import (
	"context"
	"testing"
	"time"

	"user-service/internal/app/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock: %v", err)
	}

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      db,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm connection: %v", err)
	}

	cleanup := func() {
		db.Close()
	}

	return gormDB, mock, cleanup
}

func TestUserRepository_Create(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &models.User{
		FullName: "John Doe",
		Email:    "john@example.com",
		Phone:    "1234567890",
		Password: "hashedpassword",
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `users`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Create(ctx, user)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByID(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewUserRepository(db)
	ctx := context.Background()

	now := time.Now()
	expectedUser := &models.User{
		ID:        1,
		FullName:  "John Doe",
		Email:     "john@example.com",
		Phone:     "1234567890",
		CreatedAt: now,
		UpdatedAt: now,
	}

	rows := sqlmock.NewRows([]string{"id", "full_name", "email", "phone", "created_at", "updated_at"}).
		AddRow(expectedUser.ID, expectedUser.FullName, expectedUser.Email, expectedUser.Phone, expectedUser.CreatedAt, expectedUser.UpdatedAt)

	mock.ExpectQuery("SELECT \\* FROM `users` WHERE `users`.`id` = \\? AND `users`.`deleted_at` IS NULL").
		WithArgs(1).
		WillReturnRows(rows)

	user, err := repo.GetByID(ctx, 1)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, expectedUser.Email, user.Email)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewUserRepository(db)
	ctx := context.Background()

	now := time.Now()
	expectedUser := &models.User{
		ID:        1,
		FullName:  "John Doe",
		Email:     "john@example.com",
		Phone:     "1234567890",
		CreatedAt: now,
		UpdatedAt: now,
	}

	rows := sqlmock.NewRows([]string{"id", "full_name", "email", "phone", "created_at", "updated_at"}).
		AddRow(expectedUser.ID, expectedUser.FullName, expectedUser.Email, expectedUser.Phone, expectedUser.CreatedAt, expectedUser.UpdatedAt)

	mock.ExpectQuery("SELECT \\* FROM `users` WHERE email = \\? AND `users`.`deleted_at` IS NULL").
		WithArgs("john@example.com").
		WillReturnRows(rows)

	user, err := repo.GetByEmail(ctx, "john@example.com")
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, expectedUser.Email, user.Email)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestContactRepository_List(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewContactRepository(db)
	ctx := context.Background()

	favorite := true
	req := &models.ListContactsRequest{
		Page:     1,
		Limit:    10,
		Search:   "John",
		Favorite: &favorite,
	}

	// Mock count query
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `contacts`").
		WithArgs(1, "%John%", "%John%", true).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	// Mock select query
	rows := sqlmock.NewRows([]string{"id", "user_id", "full_name", "phone", "email", "favorite", "created_at", "updated_at"}).
		AddRow(1, 1, "John Doe", "1234567890", "john@example.com", true, time.Now(), time.Now()).
		AddRow(2, 1, "John Smith", "0987654321", "smith@example.com", true, time.Now(), time.Now())

	mock.ExpectQuery("SELECT \\* FROM `contacts` WHERE user_id = \\?").
		WithArgs(1, "%John%", "%John%", true, 10, 0).
		WillReturnRows(rows)

	contacts, total, err := repo.List(ctx, 1, req)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, contacts, 2)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestContactRepository_Create(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewContactRepository(db)
	ctx := context.Background()

	email := "jane@example.com"
	contact := &models.Contact{
		UserID:   1,
		FullName: "Jane Doe",
		Phone:    "1234567890",
		Email:    &email,
		Favorite: false,
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `contacts`").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.Create(ctx, contact)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestContactRepository_GetByID(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewContactRepository(db)
	ctx := context.Background()

	now := time.Now()
	email := "jane@example.com"
	expectedContact := &models.Contact{
		ID:        1,
		UserID:    1,
		FullName:  "Jane Doe",
		Phone:     "1234567890",
		Email:     &email,
		Favorite:  false,
		CreatedAt: now,
		UpdatedAt: now,
	}

	rows := sqlmock.NewRows([]string{"id", "user_id", "full_name", "phone", "email", "favorite", "created_at", "updated_at"}).
		AddRow(expectedContact.ID, expectedContact.UserID, expectedContact.FullName, expectedContact.Phone, expectedContact.Email, expectedContact.Favorite, expectedContact.CreatedAt, expectedContact.UpdatedAt)

	mock.ExpectQuery("SELECT \\* FROM `contacts` WHERE id = \\? AND user_id = \\? AND `contacts`.`deleted_at` IS NULL").
		WithArgs(1, 1).
		WillReturnRows(rows)

	contact, err := repo.GetByID(ctx, 1, 1)
	assert.NoError(t, err)
	assert.NotNil(t, contact)
	assert.Equal(t, expectedContact.Phone, contact.Phone)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestContactRepository_Update(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewContactRepository(db)
	ctx := context.Background()

	email := "jane.updated@example.com"
	contact := &models.Contact{
		ID:       1,
		UserID:   1,
		FullName: "Jane Doe Updated",
		Phone:    "9999999999",
		Email:    &email,
		Favorite: true,
	}

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `contacts`").
		WithArgs(sqlmock.AnyArg(), contact.FullName, contact.Phone, contact.Email, contact.Favorite, contact.ID, contact.UserID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.Update(ctx, contact)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestContactRepository_Delete(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := NewContactRepository(db)
	ctx := context.Background()

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `contacts` SET `deleted_at`").
		WithArgs(sqlmock.AnyArg(), 1, 1).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.Delete(ctx, 1, 1)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
