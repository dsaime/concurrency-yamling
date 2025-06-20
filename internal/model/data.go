package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Data struct {
	// ID это идентификатор записи, создается случайная строка при создании файла и не меняется при обновлении
	ID string

	// CreatedAt это время создания записи, создается при создании файла и не изменяется
	CreatedAt time.Time

	// UpdatedAt это время последнего обновления записи, обновляется при каждом обновлении записи/файла
	UpdatedAt time.Time

	// Chain содержит идентификаторы потоков, которые обновляют данный файл, например: 1,4,3,5,2,1,6 и т.д.
	// Каждая итерация записи в файл добавляет сюда идентификатор потока
	Chain []string

	// Latencies это время, которое затрачено на генерацию данных данным потоком {1: 10ms, 4: 11ms, 3:9ms ит.д.}
	// каждый поток дописывает сюда или обновляет свое значение
	Latencies map[string]time.Duration

	// Text случайный текст
	Text string
}

func NewData() Data {
	return Data{
		ID:        uuid.NewString(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Time{},
		Chain:     nil,
		Latencies: make(map[string]time.Duration),
		Text:      uuid.NewString(),
	}
}

var (
	ErrNotFound = errors.New("not found")
)

type DataStorage interface {
	Random() (Data, error)
	Get(id string) (Data, error)
	Upsert(datas ...Data) error
	InTransaction(fn func(DataStorage) error) error
}
