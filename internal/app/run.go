package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"sync"

	"golang.org/x/exp/maps"

	fsStorage "github.com/dsaime/concurrency-yamling/internal/fs_storage"
	"github.com/dsaime/concurrency-yamling/internal/model"
)

func Run(ctx context.Context, cfg Config) error {
	// Инициализировать хранилище данных
	storage, err := fsStorage.Init(cfg.FsStorage)
	if err != nil {
		slog.Error("initFSDataStorage: " + err.Error())
		os.Exit(1)
	}

	// Собрать ID данных для обработки
	dataIDsForProcessing := make(map[string]struct{}, cfg.DatasCount)
	for i := range cfg.DatasCount {
		id := "data_" + strconv.Itoa(int(i))
		dataIDsForProcessing[id] = struct{}{}
	}

	// Количество обработок на каждые данные
	dataIDToIteration := make(map[string]uint)

	// Ограничение доступа к данным
	dataToMutex := make(map[string]*sync.Mutex)
	for dataID := range dataIDsForProcessing {
		dataToMutex[dataID] = &sync.Mutex{}
	}

	// Ограничение доступа к общим переменным
	var mu sync.Mutex

	// Идентификаторы потоков обработки данных
	threadIDs := make([]string, cfg.ThreadsCount)
	for i := range threadIDs {
		threadIDs[i] = "th_" + strconv.Itoa(i)
	}

	// Ожидание окончания работы всех потоков
	var wg sync.WaitGroup
	wg.Add(len(threadIDs))

	for _, threadID := range threadIDs {
		go func() {
			defer wg.Done()
			for {
				// Проверить жизнь контекста
				select {
				case <-ctx.Done():
					slog.Info("thread canceled",
						"threadID", threadID)
					return
				default:
				}

				// Получить эксклюзивный доступ к переменным
				mu.Lock()
				// Завершить обработчик, если данных на обработку нет
				if len(dataIDsForProcessing) == 0 {
					slog.Info("no data for processing",
						"threadID", threadID)
					mu.Unlock()
					return
				}
				// Взять случайный ID из доступных к обработке данных
				dataIDForProcessing := maps.Keys(dataIDsForProcessing)[0]
				// Если данные обработались нужное количество раз, удалить из доступных и оповестить wg
				if dataIDToIteration[dataIDForProcessing] >= cfg.MaxIterations {
					delete(dataIDsForProcessing, dataIDForProcessing)
					// Освободить доступ к переменным
					mu.Unlock()
					slog.Info("data is end of processing",
						"threadID", threadID,
						"dataID", dataIDForProcessing)
					continue
				}
				// Освободить доступ к переменным
				mu.Unlock()

				slog.Info("thread start processing",
					"threadID", threadID,
					"dataID", dataIDForProcessing)

				// Получить эксклюзивный доступ к данным по ID
				dataMu := dataToMutex[dataIDForProcessing]
				dataMu.Lock()

				// Запустить обработку задачи
				if err := runDataProcessing(threadID, dataIDForProcessing, storage); err != nil {
					slog.Error("data processing failed: "+err.Error(),
						"threadID", threadID,
						"dataID", dataIDForProcessing)
				} else {
					mu.Lock()
					dataIDToIteration[dataIDForProcessing]++
					mu.Unlock()
					slog.Info("data processing finished",
						"threadID", threadID,
						"dataID", dataIDForProcessing)
				}
				// Освободить доступ к данным
				dataMu.Unlock()
			}
		}()
	}

	wg.Wait()
	return ctx.Err()
}

// runDataProcessing получает данные по ID из хранилища, обрабатывает и сохраняет обратно
func runDataProcessing(threadID string, dataIDForProcessing string, storage model.DataStorage) error {
	// Получить данные из хранилища
	dataForProcessing, err := storage.Find(dataIDForProcessing)
	if errors.Is(err, model.ErrNotFound) {
		dataForProcessing = model.NewData(dataIDForProcessing)
	} else if err != nil {
		return fmt.Errorf("storage.Find: %w", err)
	}

	// Обработать данные
	dataProcessed := model.ProcessData(dataForProcessing, threadID)

	// Сохранить обработанные данные
	if err := storage.Upsert(dataProcessed); err != nil {
		return fmt.Errorf("storage.Upsert: %w", err)
	}

	return nil
}
