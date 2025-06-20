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
				Name:        "pgsql-dsn",
				Destination: &cfg.Pgsql.DSN,
				Usage:       "PostgreSQL connection string in format 'postgres://user:password@host:port/dbname'",
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "redis-url",
				Destination: &cfg.Redis.RedisURL,
				Usage:       "Redis connection string in format 'redis://[[username][:password]@]host[:port][/db-number]'",
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "http-addr",
				Destination: &cfg.HttpAddr,
				Usage:       "HTTP server address",
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "nats-url",
				Destination: &cfg.Nats.NatsURL,
				Usage:       "NATS connection string in format 'nats://user:password@host:port'",
				Required:    true,
			},
			&cli.DurationFlag{
				Name:        "cache-expiration",
				Destination: &cfg.Redis.Expiration,
				Usage:       "Cache expiration time",
				Value:       time.Minute * 5,
			},
		},
	}
}
