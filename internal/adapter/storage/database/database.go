package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/playmixer/secret-keeper/internal/adapter/keeperr"
	"github.com/playmixer/secret-keeper/internal/adapter/models"
)

type Storage struct {
	db *gorm.DB
}

func New(dsn string) (*Storage, error) {
	sqlDB, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed open connect: %w", err)
	}
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}))
	if err != nil {
		return nil, fmt.Errorf("failed open gorm connect: %w", err)
	}

	db := &Storage{
		db: gormDB,
	}

	if err := db.Migration(); err != nil {
		return nil, fmt.Errorf("failed auto migrations: %w", err)
	}

	return db, nil
}

func (s *Storage) Migration() error {
	if err := s.db.AutoMigrate(&models.User{}, &models.Secret{}); err != nil {
		return fmt.Errorf("failed migrations: %w", err)
	}
	return nil
}

func (s *Storage) Registration(ctx context.Context, login, passwordHash string) error {
	user := &models.User{
		Login:        login,
		PasswordHash: passwordHash,
	}

	err := s.db.WithContext(ctx).Create(user).Error
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return fmt.Errorf("login not unique: %w %w", err, keeperr.ErrLoginNotUnique)
		}
		return fmt.Errorf("failed create user: %w", err)
	}

	return nil
}
func (s *Storage) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	user := &models.User{}
	err := s.db.WithContext(ctx).Where("login = ?", login).First(user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.Join(keeperr.ErrNotFound, err)
		}
		return nil, fmt.Errorf("failed find user: %w", err)
	}

	return user, nil
}
func (s *Storage) GetMetaDatasByUserID(ctx context.Context, userID uint) (*[]models.Secret, error) {
	data := []models.Secret{}
	err := s.db.Where("user_id = ?", userID).Find(&data).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &data, nil
		}
		return nil, fmt.Errorf("failed get data: %w", err)
	}
	return &data, nil
}

func (s *Storage) NewSecret(ctx context.Context, secret *models.Secret) (*models.Secret, error) {
	err := s.db.Create(secret).Error
	if err != nil {
		return nil, fmt.Errorf("failed create secret: %w", err)
	}

	return secret, nil
}

func (s *Storage) GetSecret(ctx context.Context, id uint) (*models.Secret, error) {
	secret := &models.Secret{}
	err := s.db.Where("id = ?", id).First(secret).Error
	if err != nil {
		return nil, fmt.Errorf("failed get secret: %w", err)
	}
	return secret, nil
}

func (s *Storage) UpdSecret(ctx context.Context, secret *models.Secret) (*models.Secret, error) {
	err := s.db.Where("id = ?", secret.ID).Save(secret).Error
	if err != nil {
		return nil, fmt.Errorf("failed update secret: %w", err)
	}
	return secret, nil
}

func (s *Storage) DelSecret(ctx context.Context, id uint) error {
	secret, err := s.GetSecret(ctx, id)
	if err != nil {
		return fmt.Errorf("failed get secret: %w", err)
	}
	secret.IsDeleted = true
	secret.Data = []byte{}
	err = s.db.Save(secret).Error
	if err != nil {
		return fmt.Errorf("failed save deleting secret id=`%v`: %w", id, err)
	}
	return nil
}
