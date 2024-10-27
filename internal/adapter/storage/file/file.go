package file

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/playmixer/secret-keeper/internal/adapter/models"
	"github.com/playmixer/secret-keeper/pkg/tools"
)

var (
	lengthNameFile uint = 20
)

type Storage struct {
	log      *zap.Logger
	path     string
	filename string
	store    []models.FileMetaDataItem
}

type option func(*Storage)

func SetPath(path string) option {
	return func(s *Storage) {
		s.path = path
	}
}

func SetLogger(log *zap.Logger) option {
	return func(s *Storage) {
		s.log = log
	}
}

func Init(options ...option) (*Storage, error) {
	s := &Storage{
		path: "./data",
		log:  zap.NewNop(),
	}

	for _, opt := range options {
		opt(s)
	}

	return s, nil
}

func (s *Storage) Open(name string) error {
	s.log.Debug("Open store")
	s.filename = tools.GetMD5Hash(name)
	s.store = []models.FileMetaDataItem{}

	err := os.Mkdir(s.path, tools.Mode0755)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return fmt.Errorf("failed create data directory: %w", err)
	}
	f, err := os.OpenFile(s.getFullPath(s.filename), os.O_CREATE|os.O_RDONLY, tools.Mode0600)
	if err != nil {
		return fmt.Errorf("failed open storage file: %w", err)
	}
	defer func() {
		err := f.Close()
		if err != nil {
			s.log.Error("failed close file", zap.Error(err))
		}
	}()

	bData, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("failed read from storage file: %w", err)
	}

	if len(bData) == 0 {
		return nil
	}

	err = json.Unmarshal(bData, &s.store)
	if err != nil {
		return fmt.Errorf("failed unmarshal storage data: %w", err)
	}

	return nil
}

func (s *Storage) save() error {
	s.log.Debug("Save store", zap.String("store", fmt.Sprint(s.store)))
	bStore, err := json.Marshal(s.store)
	if err != nil {
		s.log.Error("failed marshal store", zap.Error(err))
		return fmt.Errorf("failed marshal store: %w", err)
	}

	err = os.WriteFile(s.getFullPath(s.filename), bStore, tools.Mode0600)
	if err != nil {
		s.log.Error("failed write file", zap.Error(err))
		return fmt.Errorf("failed write file: %w", err)
	}

	return nil
}

func (s *Storage) Close() error {
	s.log.Debug("Close store", zap.String("store", fmt.Sprint(s.store)))
	err := s.save()
	if err != nil {
		return fmt.Errorf("failed save storage: %w", err)
	}
	s.store = []models.FileMetaDataItem{}
	return nil
}

func (s *Storage) getFullPath(path string) string {
	return s.path + "/" + path
}

func (s *Storage) UpdateDate() int64 {
	return time.Now().UTC().Unix()
}

func (s *Storage) WriteFile(filename string, data *[]byte) error {
	err := os.WriteFile(s.getFullPath(filename), *data, tools.Mode0600)
	if err != nil {
		return fmt.Errorf("failed write new card: %w", err)
	}

	return nil
}

func (s *Storage) OpenData(id int64) (*[]byte, error) {
	res := []byte{}
	filename := ""
	for _, e := range s.store {
		if e.ID == id {
			filename = e.OriginalPath
			break
		}
	}

	if filename == "" {
		return nil, errors.New("data not found")
	}

	f, err := os.Open(s.getFullPath(filename))
	if err != nil {
		return nil, fmt.Errorf("failed open file: %w", err)
	}
	defer func() {
		err := f.Close()
		if err != nil {
			s.log.Error("failed close file", zap.Error(err))
		}
	}()

	res, err = io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed read data from file: %w", err)
	}

	return &res, nil
}

func (s *Storage) GetAll() (*[]models.FileMetaDataItem, error) {
	return &s.store, nil
}

func (s *Storage) Get(id int64) (*models.FileMetaDataItem, error) {
	for _, v := range s.store {
		if v.ID == id {
			return &v, nil
		}
	}
	return nil, fmt.Errorf("not found id=`%v`", id)
}

func (s *Storage) UpdMeta(m *models.FileMetaDataItem) error {
	s.log.Debug("upd meta store", zap.String("data", fmt.Sprint(*m)))
	newStore := []models.FileMetaDataItem{}
	for _, v := range s.store {
		if v.ID == m.ID {
			newStore = append(newStore, *m)
		} else {
			newStore = append(newStore, v)
		}
	}
	s.store = newStore
	return nil
}

func (s *Storage) GetData(id int64) (*models.FileMetaDataItem, *[]byte, error) {
	m, err := s.Get(id)
	if err != nil {
		return nil, nil, fmt.Errorf("failed found data: %w", err)
	}

	bData, err := s.OpenData(id)
	if err != nil {
		return nil, nil, fmt.Errorf("failed open data: %w", err)
	}

	return m, bData, nil
}

func (s *Storage) NewData(eID uint, updateDT int64, title string, dataType models.DataType, data *[]byte) (
	*models.FileMetaDataItem, error) {
	filename := s.filename + tools.RandomString(lengthNameFile)

	m := models.FileMetaDataItem{
		ID:           time.Now().UnixMicro(),
		ExternalID:   eID,
		OriginalPath: filename,
		Title:        title,
		DataType:     dataType,
		UpdateDT:     s.UpdateDate(),
	}
	if updateDT > 0 {
		m.UpdateDT = updateDT
	}

	err := s.WriteFile(filename, data)
	if err != nil {
		return nil, fmt.Errorf("failed write new card: %w", err)
	}

	s.store = append(s.store, m)

	return &m, nil
}

func (s *Storage) EditData(id int64, m *models.FileMetaDataItem, data *[]byte) error {
	s.log.Debug("edit data", zap.Int64("id", id), zap.String("meta", fmt.Sprint(*m)), zap.Int("len", len(*data)))
	err := s.WriteFile(m.OriginalPath, data)
	if err != nil {
		return fmt.Errorf("faieled write file: %w", err)
	}

	// m.UpdateDT = s.updateDate()
	err = s.UpdMeta(m)
	if err != nil {
		return fmt.Errorf("failed upd meta store: %w", err)
	}

	return nil
}

func (s *Storage) DelData(id int64) error {
	m, _, err := s.GetData(id)
	if err != nil {
		return fmt.Errorf("failed get data: %w", err)
	}

	err = os.Remove(s.getFullPath(m.OriginalPath))
	if err != nil {
		s.log.Error("failed remove data", zap.Error(err), zap.String("filename", m.OriginalPath))
		return fmt.Errorf("failed remove data: %w", err)
	}

	m.IsDeleted = true
	err = s.UpdMeta(m)
	if err != nil {
		return fmt.Errorf("failed upd meta store: %w", err)
	}

	return nil
}

func (s *Storage) UploadFileToPath(id int64, path string) error {
	m, err := s.Get(id)
	if err != nil {
		return fmt.Errorf("failed get meta data: %w", err)
	}

	outFile, err := os.Open(s.getFullPath(m.OriginalPath))
	if err != nil {
		return fmt.Errorf("failed open file: %w", err)
	}
	defer func() {
		err := outFile.Close()
		if err != nil {
			s.log.Error("failed close out file", zap.Error(err))
		}
	}()

	inFile, err := os.Create(path + "/" + m.Filename)
	if err != nil {
		return fmt.Errorf("failed create new file: %w", err)
	}

	_, err = io.Copy(inFile, outFile)
	if err != nil {
		return fmt.Errorf("failed close outfile: %w", err)
	}

	defer func() {
		err := inFile.Close()
		if err != nil {
			s.log.Error("failed close infile", zap.Error(err))
		}
	}()

	return nil
}
