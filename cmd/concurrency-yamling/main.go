package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/google/uuid"

	"github.com/dsaime/concurrency-yamling/internal/common"
	"github.com/dsaime/concurrency-yamling/internal/model"
	processData "github.com/dsaime/concurrency-yamling/internal/process_data"
)

func main() {
	ctx := context.TODO()

	// Инициализировать хранилище данных
	storage, err := initFSDataStorage()
	if err != nil {
		slog.Error("initFSDataStorage: " + err.Error())
		os.Exit(1)
	}

	// Создать много данных
	datas, err := newRandomDatas(1000000)
	if err != nil {
		slog.Error("newRandomDatas: " + err.Error())
		os.Exit(1)
	}

	// Сохранить созданные данные в хранилище
	if err = storage.Upsert(datas...); err != nil {
		slog.Error("storage.Upsert: " + err.Error())
		os.Exit(1)
	}

	// Создать еще много данных, но не сохранять
	datasUnsaved, err := newRandomDatas(1000000)
	if err != nil {
		slog.Error("newRandomDatas: " + err.Error())
		os.Exit(1)
	}

	// Собрать ID данных для обработки
	dataIDs := make([]string, len(datas)+len(datasUnsaved))
	for i, data := range append(datas, datasUnsaved...) {
		dataIDs[i] = data.ID
	}

	// Инициализировать потоки обработки данных
	var threads []DataProcessingThread

	for i, th := range threads {
		go func() {
			for {
				select {
				case <-ctx.Done():
					slog.Info("thread canceled",
						"threadID", th.ThreadID())
					return
				default:
					randomDataID := common.RndElem(dataIDs)
					if err := storage.InTransaction(func(storage model.DataStorage) error {
						slog.Info("thread start processing",
							"threadID", th.ThreadID(),
							"dataID", randomDataID)

						// Получить данные из хранилища
						randomData, err := storage.Get(randomDataID)
						if errors.Is(err, model.ErrNotFound) {
							if randomData, err = model.NewData(uuid.NewString()); err != nil {
								return fmt.Errorf("model.NewData: %w", err)
							}
						} else if err != nil {
							return fmt.Errorf("storage.Get: %w", err)
						}
						slog.Info("data found")

						// Обработать данные
						newData, err := th.ProcessData(randomData)
						if err != nil {
							return fmt.Errorf("th.ProcessData: %w", err)
						}
						return storage.Upsert(newData)
					}); err != nil {
						slog.Error("data processing failed: "+err.Error(),
							"threadID", th.ThreadID(),
							"dataID", randomDataID)
						continue
					}
					datas = append(datas[:i], datas[i+1:]...)

					slog.Info("data processing finished",
						"threadID", th.ThreadID(),
						"dataID", randomDataID)
				}
			}
		}()
	}

	//threadIDs := make([]string, 1000000)
	//err := loop(context.TODO(), nil, "1")
	//if errors.Is(err, context.Canceled) {
	//	slog.Info("loop: canceled")
	//} else if err != nil {
	//	slog.Error("loop: " + err.Error())
	//	os.Exit(1)
	//}
}

func initFSDataStorage() (model.DataStorage, error) {

}

func loop(ctx context.Context, dataCh <-chan model.Data, threadID string) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case dataForProcessing := <-dataCh:
			dataProcessed, err := processData.Run(processData.Params{
				Data:     dataForProcessing,
				ThreadID: threadID,
			})
			if err != nil {
				slog.Error("processData.Run: "+err.Error(),
					"data.ID", dataForProcessing.ID)
			}
		}
	}
}

func newRandomDatas(count int) ([]model.Data, error) {
	datas := make([]model.Data, count)
	for i := range datas {
		data, err := model.NewData(uuid.NewString())
		if err != nil {
			return nil, fmt.Errorf("model.NewData i=%d: %w", i, err)
		}
		datas[i] = data
	}

	return datas, nil
}

type DataProcessingThread interface {
	ThreadID() string
	ProcessData(model.Data) (model.Data, error)
}
