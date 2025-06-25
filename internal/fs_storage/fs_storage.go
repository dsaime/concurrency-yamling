package fsStorage

import (
	"errors"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"

	"github.com/dsaime/concurrency-yamling/internal/model"
)

// Storage представляет собой реализацию хранилища данных в YAML файлах
type Storage struct {
	dir string     // Путь к директории
	mu  sync.Mutex // Ограничение доступа к директории с файлами
}

type Config struct {
	Dir string // Путь к директории в которой будут храниться yaml-файлы
}

// Init инициализирует хранилище данных в файловой системе
func Init(cfg Config) (*Storage, error) {
	// Создаем директорию, если её нет
	cfgDir := filepath.Clean(cfg.Dir + "/")
	//if err := os.MkdirAll(filepath.Clean(cfg.Dir+"/"), 0755); err != nil {
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		return nil, err
	}
	return &Storage{
		dir: cfgDir,
	}, nil
}

// dataFilePath возвращает путь к файлу по ID данных
func (s *Storage) dataFilePath(dataID string) string {
	return s.dir + "/" + dataID + ".yaml"
}

// Find ищет запись по ID
func (s *Storage) Find(id string) (model.Data, error) {
	// Валидировать входящие параметры
	if id == "" {
		return model.Data{}, errors.New("id не может быть пустым")
	}

	// Ограничить доступ к директории
	s.mu.Lock()
	defer s.mu.Unlock()

	var data model.Data

	// Прочитать файл
	b, err := os.ReadFile(s.dataFilePath(id))
	if os.IsNotExist(err) {
		return model.Data{}, model.ErrNotFound
	} else if err != nil {
		return model.Data{}, err
	}

	// Разобрать содержимое как yaml
	return data, yaml.Unmarshal(b, &data)
}

// Upsert обновляет или добавляет данные в yaml-файл
func (s *Storage) Upsert(data model.Data) error {
	// Валидировать входящие параметры
	if data.ID == "" {
		return errors.New("id не может быть пустым")
	}

	// Сериализовать данные
	b, err := yaml.Marshal(data)
	if err != nil {
		return err
	}

	// Ограничить доступ к директории
	s.mu.Lock()
	defer s.mu.Unlock()

	// Записать новое yaml-содержимое в файл
	return os.WriteFile(s.dataFilePath(data.ID), b, 0644)
}
