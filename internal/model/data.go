package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/exp/maps"
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

func NewData(id string) Data {
	//if err := ValidateDateID(); err != nil { ... }

	return Data{
		ID:        id,
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
	Find(id string) (Data, error)
	Upsert(Data) error
}

func ProcessData(d Data, threadID string) Data {
	start := time.Now()
	newLatencies := make(map[string]time.Duration, len(d.Latencies))
	maps.Copy(newLatencies, d.Latencies)
	newData := Data{
		ID:        d.ID,
		CreatedAt: d.CreatedAt,
		UpdatedAt: time.Now(),
		Chain:     append(d.Chain, threadID),
		Latencies: newLatencies,
		Text:      uuid.NewString(),
	}
	newData.Latencies[threadID] += time.Since(start)
	return newData
}
