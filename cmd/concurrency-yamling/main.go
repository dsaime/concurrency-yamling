package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/urfave/cli/v3"

	"github.com/dsaime/concurrency-yamling/internal/app"
)

func main() {
	slog.Info("main: starting")
	ctx, cancel := context.WithCancel(context.Background())
	go appRun(ctx)
	waitInterrupt(cancel)
}

// appRun запускает приложение и обрабатывает результат
func appRun(ctx context.Context) {
	err := initCliCommand().Run(ctx, os.Args)
	if errors.Is(err, context.Canceled) {
		slog.Info("main: appRun: exit by context canceled")
		os.Exit(0)
	} else if err != nil {
		slog.Error("main: appRun: " + err.Error())
		os.Exit(1)
	}
	slog.Info("main: appRun: finished")
	os.Exit(0)
}

// waitInterrupt отменяет контекст, когда в приложение поступает сигнал syscall.SIGINT или syscall.SIGTERM
func waitInterrupt(cancel context.CancelFunc) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGINT, syscall.SIGTERM)
	slog.Info("main: waitInterrupt: Received signal " + (<-interrupt).String())
	cancel()
	slog.Info("main: waitInterrupt: Context canceled")
	time.Sleep(3 * time.Second)
}

// initCliCommand создает команду, для разбора аргументов командной строки и запуска приложения
func initCliCommand() *cli.Command {
	var cfg app.Config
	return &cli.Command{
		Action: func(ctx context.Context, command *cli.Command) error {
			return app.Run(ctx, cfg)
		},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "storage-dir",
				Destination: &cfg.FsStorage.Dir,
				Usage:       "Директория, в которой будут храниться yaml-файлы обрабатываемых данных",
				Value:       "./storage/",
			},
			&cli.UintFlag{
				Name:        "max-iterations",
				Destination: &cfg.MaxIterations,
				Usage:       "Сколько итераций требуется для завершения обработки данных",
				Value:       100,
			},
			&cli.UintFlag{
				Name:        "threads-count",
				Destination: &cfg.ThreadsCount,
				Usage:       "Количество потоков",
				Value:       20,
			},
			&cli.UintFlag{
				Name:        "datas-count",
				Destination: &cfg.DatasCount,
				Usage:       "Количество данных к обработке",
				Value:       10,
			},
		},
	}
}
