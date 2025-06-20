package processData

import "github.com/dsaime/concurrency-yamling/internal/model"

type Params struct {
	Data     model.Data
	ThreadID string
}

func Run(p Params) (model.Data, error) {

}
