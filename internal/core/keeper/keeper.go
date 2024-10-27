package keeper

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"time"

	"gorm.io/gorm"

	"github.com/playmixer/secret-keeper/internal/adapter/keeperr"
	"github.com/playmixer/secret-keeper/internal/adapter/models"
)

// Storage интерфейс хранилища.
type Storage interface {
	Registration(ctx context.Context, login, passwordHash string) error
	GetUserByLogin(ctx context.Context, login string) (*models.User, error)
	GetMetaDatasByUserID(ctx context.Context, userID uint) (*[]models.Secret, error)
	NewSecret(ctx context.Context, secret *models.Secret) (*models.Secret, error)
	GetSecret(ctx context.Context, id uint) (*models.Secret, error)
	UpdSecret(ctx context.Context, secret *models.Secret) (*models.Secret, error)
	DelSecret(ctx context.Context, id uint) error
}

// Keeper - Keeper.
type Keeper struct {
	store      Storage
	encryptKey string
}

type option func(*Keeper)

// SetEncryptKey установить секретный ключ.
func SetEncryptKey(key string) option {
	return func(k *Keeper) {
		k.encryptKey = key
	}
}

// New - создаем Keeper.
func New(store Storage, options ...option) (*Keeper, error) {
	k := &Keeper{
		store:      store,
		encryptKey: "",
	}

	for _, opt := range options {
		opt(k)
	}

	return k, nil
}

// Registration регистрация пользователя.
func (k *Keeper) Registration(ctx context.Context, login, password string) error {
	if err := validateLogin(login); err != nil {
		return fmt.Errorf("login invalid: %w", err)
	}
	if err := validatePassword(password); err != nil {
		return fmt.Errorf("password invalid: %w", err)
	}

	passwordHash, err := HashPassword(password)
	if err != nil {
		return fmt.Errorf("failed hashing password: %w", err)
	}

	err = k.store.Registration(ctx, login, passwordHash)
	if err != nil {
		return fmt.Errorf("failed user regstration: %w", err)
	}

	return nil
}

// Login авторизация пользователя.
func (k *Keeper) Login(ctx context.Context, login, password string) (*models.User, error) {
	if err := validateLogin(login); err != nil {
		return nil, fmt.Errorf("login invalid: %w", err)
	}
	if err := validatePassword(password); err != nil {
		return nil, fmt.Errorf("password invalid: %w", err)
	}

	user, err := k.store.GetUserByLogin(ctx, login)
	if err != nil {
		return nil, fmt.Errorf("failed find user: %w", err)
	}

	if !checkPasswordHash(password, user.PasswordHash) {
		return nil, keeperr.ErrLoginOrPasswordNotCorrect
	}

	return user, nil
}

// GetMetaDatasByUserID получить все мето данные пользователя из стора.
func (k *Keeper) GetMetaDatasByUserID(ctx context.Context, userID uint) (*[]models.Secret, error) {
	secret, err := k.store.GetMetaDatasByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed get meta datas by user `%v`: %w", userID, err)
	}
	return secret, nil
}

// GetSecret получить данные из стора.
func (k *Keeper) GetSecret(ctx context.Context, id uint) (*models.Secret, error) {
	secret, err := k.store.GetSecret(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed get secret: %w", err)
	}
	eData, err := k.decrypt([]byte(k.encryptKey), secret.Data)
	if err != nil {
		return nil, fmt.Errorf("failed encrypt data with key `%s`: %w", k.encryptKey, err)
	}
	secret.Data = eData

	return secret, nil
}

// NewSecret создаем данные в сторе.
func (k *Keeper) NewSecret(ctx context.Context,
	data *[]byte, title string, dataType models.DataType, updateDT int64, userID uint) (*models.Secret, error) {
	eData, err := k.encrypt([]byte(k.encryptKey), *data)
	if err != nil {
		return nil, fmt.Errorf("failed encrypt data: %w", err)
	}
	secret := &models.Secret{
		UserID:   userID,
		Title:    title,
		DataType: dataType,
		Data:     eData,
	}

	if updateDT > 0 {
		secret.UpdateDT = updateDT
	} else {
		secret.UpdateDT = time.Now().UTC().Unix()
	}

	secret, err = k.store.NewSecret(ctx, secret)
	if err != nil {
		return nil, fmt.Errorf("failed create secret: %w", err)
	}

	return secret, nil
}

// UpdSecret обновляем данные в сторе.
func (k *Keeper) UpdSecret(
	ctx context.Context, id uint, data *[]byte, title string,
	dataType models.DataType, updateDT int64, userID uint,
) (*models.Secret, error) {
	eData, err := k.encrypt([]byte(k.encryptKey), *data)
	if err != nil {
		return nil, fmt.Errorf("failed encrypt data: %w", err)
	}
	secret := &models.Secret{
		Model: gorm.Model{
			ID: id,
		},
		Data:     eData,
		Title:    title,
		DataType: dataType,
		UpdateDT: updateDT,
		UserID:   userID,
	}
	secret, err = k.store.UpdSecret(ctx, secret)
	if err != nil {
		return nil, fmt.Errorf("failed update secret: %w", err)
	}

	return secret, nil
}

func (k *Keeper) encrypt(key, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed create cipgher: %w", err)
	}
	b := base64.StdEncoding.EncodeToString(text)
	ciphertext := make([]byte, aes.BlockSize+len(b))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, fmt.Errorf("failed read: %w", err)
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(b))
	return ciphertext, nil
}

func (k *Keeper) decrypt(key, text []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed create cipher: %w", err)
	}
	if len(text) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}
	iv := text[:aes.BlockSize]
	text = text[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(text, text)
	data, err := base64.StdEncoding.DecodeString(string(text))
	if err != nil {
		return nil, fmt.Errorf("failed decode: %w", err)
	}
	return data, nil
}

// DelSecret удаляем данные из стора.
func (k *Keeper) DelSecret(ctx context.Context, id uint) error {
	err := k.store.DelSecret(ctx, id)
	if err != nil {
		return fmt.Errorf("failed delete secret: %w", err)
	}
	return nil
}
