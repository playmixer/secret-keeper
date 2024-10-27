package uiapi

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/playmixer/secret-keeper/internal/adapter/models"
)

type store interface {
	UpdateDate() int64
	Open(name string) error
	Close() error
	Get(id int64) (*models.FileMetaDataItem, error)
	UpdMeta(m *models.FileMetaDataItem) error
	GetAll() (*[]models.FileMetaDataItem, error)
	NewData(eID uint, updateDT int64, title string, dataType models.DataType, data *[]byte) (
		*models.FileMetaDataItem, error,
	)
	GetData(id int64) (*models.FileMetaDataItem, *[]byte, error)
	EditData(id int64, m *models.FileMetaDataItem, data *[]byte) error
	DelData(id int64) error
	UploadFileToPath(id int64, path string) error
}

const (
	kilobyte = 1024
	byte8    = 8
)

var (
	periodTickWorker time.Duration = 10 * time.Second
)

type keepClient struct {
	store       store
	log         *zap.Logger
	apiURL      string
	token       string
	fileMaxSize int64
}

type option func(*keepClient)

func SetConfig(cfg Config) option {
	return func(k *keepClient) {
		k.apiURL = cfg.APIAddress
	}
}

func SetFileMaxSize(size int64) option {
	return func(kc *keepClient) {
		kc.fileMaxSize = size
	}
}

func New(ctx context.Context, store store, lgr *zap.Logger, options ...option) (*keepClient, error) {
	k := &keepClient{
		store:       store,
		log:         lgr,
		fileMaxSize: kilobyte * byte8,
	}

	for _, opt := range options {
		opt(k)
	}

	go k.worker(ctx)

	return k, nil
}

func (k *keepClient) worker(ctx context.Context) {
	ticker := time.NewTicker(periodTickWorker)
	for {
		select {
		case <-ctx.Done():
			k.log.Info("Worker stoped")
			return

		case <-ticker.C:
			if k.token != "" {
				k.log.Debug("синхронизация данных")
				k.updateStore(ctx)
			}
		}
	}
}

func (k *keepClient) updateStore(ctx context.Context) {
	exData, err := k.eventGetExternalMetaDatas()
	if err != nil {
		k.log.Error("failed get external data", zap.Error(err))
		return
	}

	lData, err := k.store.GetAll()
	if err != nil {
		k.log.Error("failed get data from store", zap.Error(err))
		return
	}
	k.log.Debug("local data", zap.String("data", fmt.Sprint(*lData)))

loopExternal:
	for _, e := range *exData {
		for _, l := range *lData {
			select {
			case <-ctx.Done():
				return
			default:
				if e.ID == l.ExternalID {
					if l.IsDeleted || e.IsDeleted {
						if !l.IsDeleted {
							err = k.eventDeleteData(l.ID)
							if err != nil {
								k.log.Error("failed delete local data", zap.Error(err), zap.Uint("id", e.ID))
							}
						}
						if !e.IsDeleted {
							err = k.eventDeleteExternalData(e.ID)
							if err != nil {
								k.log.Error("failed delete local data", zap.Error(err), zap.Uint("id", e.ID))
							}
						}
						continue loopExternal
					}
					if l.UpdateDT < e.UpdatedDT {
						k.log.Debug("remote newed", zap.Uint("external_id", e.ID), zap.Int64("local_id", l.ID))
						err = k.updateLocalData(l.ID, e.ID)
						if err != nil {
							k.log.Error("failed update local data", zap.Error(err), zap.Uint("id", e.ID))
							continue loopExternal
						}
					}
					if l.UpdateDT > e.UpdatedDT {
						k.log.Debug("local newed", zap.Uint("external_id", e.ID), zap.Int64("local_id", l.ID))
						err = k.updateExternalData(l.ID, e.ID)
						if err != nil {
							k.log.Error("failed update external data", zap.Error(err), zap.Uint("id", e.ID))
							continue loopExternal
						}
					}
					continue loopExternal
				}
			}
		}
		// не нашли данные в локальном сторе, добавляем.
		if !e.IsDeleted {
			k.log.Debug("not found in local", zap.Uint("external_id", e.ID))
			err := k.addLocalData(e.ID)
			if err != nil {
				k.log.Error("failed add local data", zap.Error(err), zap.Uint("id", e.ID))
				continue loopExternal
			}
		}
	}

loopLocal:
	for _, l := range *lData {
		for _, e := range *exData {
			select {
			case <-ctx.Done():
				return
			default:
				if e.ID == l.ExternalID {
					continue loopLocal
				}
			}
		}
		// не нашли данные в удаленном сторе, добавляем.
		if !l.IsDeleted {
			err := k.addExternalData(l.ID)
			if err != nil {
				k.log.Error("failed add external data", zap.Error(err), zap.Int64("id", l.ID))
				continue loopLocal
			}
		}
	}
}

func (k *keepClient) updateLocalData(lID int64, eID uint) error {
	exData, err := k.eventGetExternalData(eID)
	if err != nil {
		return fmt.Errorf("failed get external card: %w", err)
	}

	m, _, err := k.store.GetData(lID)
	if err != nil {
		return fmt.Errorf("failed get data: %w", err)
	}
	m.UpdateDT = exData.UpdatedDT
	m.Title = exData.Title
	err = k.store.EditData(lID, m, exData.Data)
	if err != nil {
		return fmt.Errorf("failed update locale card: %w", err)
	}

	return nil
}

func (k *keepClient) addLocalData(eID uint) error {
	exData, err := k.eventGetExternalData(eID)
	if err != nil {
		return fmt.Errorf("failed get external card: %w", err)
	}
	k.log.Debug("add local data", zap.Uint("id", eID), zap.String("title", exData.Title),
		zap.String("type", string(exData.DataType)))
	_, err = k.store.NewData(eID, exData.UpdatedDT, exData.Title, exData.DataType, exData.Data)
	if err != nil {
		return fmt.Errorf("failed create data: %w", err)
	}
	return nil
}

func (k *keepClient) updateExternalData(lID int64, eID uint) error {
	meta, data, err := k.store.GetData(lID)
	if err != nil {
		return fmt.Errorf("failed get local card: %w", err)
	}
	k.log.Debug("meta", zap.String("data", fmt.Sprint(*meta)))
	err = k.eventUpdExternalData(eID, meta.Title, data, meta.DataType, meta.UpdateDT)
	if err != nil {
		return errors.New("failed upd external data")
	}
	return nil
}

func (k *keepClient) addExternalData(lID int64) error {
	meta, data, err := k.store.GetData(lID)
	if err != nil {
		return fmt.Errorf("failed get local card: %w", err)
	}
	k.log.Debug("meta", zap.String("data", fmt.Sprint(*meta)))
	exData, err := k.eventAddExternalData(meta.Title, data, meta.DataType, meta.UpdateDT)
	if err != nil {
		return errors.New("failed upd external data")
	}

	meta.ExternalID = exData.ID
	err = k.store.EditData(meta.ID, meta, data)
	if err != nil {
		return fmt.Errorf("failed create local data %w", err)
	}

	return nil
}
