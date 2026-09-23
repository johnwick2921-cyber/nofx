package hook

import (
	"log"
	"net/http"
)

type NewAsterTraderResult struct {
	Err    error
	Client *http.Client
}

func (r *NewAsterTraderResult) Error() error {
	if r.Err != nil {
		log.Printf("⚠️ Error executing NewAsterTraderResult: %v", r.Err)
	}
	return r.Err
}

func (r *NewAsterTraderResult) GetResult() *http.Client {
	r.Error()
	return r.Client
}
