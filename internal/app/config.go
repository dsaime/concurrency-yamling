package app

import fsStorage "github.com/dsaime/concurrency-yamling/internal/fs_storage"

type Config struct {
	FsStorage     fsStorage.Config
	MaxIterations uint
	ThreadsCount  uint
	DatasCount    uint
}
