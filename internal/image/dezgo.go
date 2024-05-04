package image

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/go-logr/logr"
	"io"
	"net/http"
	"os"
)

const DEZGO_URL string = "https://api.dezgo.com/text2image"

type Dezgo struct {
}

func (d *Dezgo) Generate(ctx context.Context, input Input, logger logr.Logger) (Output, error) {
	logger.Info("Making API CALL now!!!!")

	res, err := json.Marshal(input)
	if err != nil {
		return Output{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, DEZGO_URL, bytes.NewReader(res))
	if err != nil {
		return Output{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Dezgo-Key", os.Getenv("DEZGO_API_KEY"))
	logger.Info("APIKEY:" + os.Getenv("DEZGO_API_KEY"))
	client := http.DefaultClient
	rsp, err := client.Do(req)
	if err != nil {
		return Output{}, err
	}
	defer rsp.Body.Close() //TODO read more about defer

	seed := rsp.Header.Get("x-input-seed")
	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		return Output{}, err
	}

	logger.Info("", "seed: ", seed)
	return Output{
		Seed: seed,
		Data: body,
	}, nil
}

var _ Generator = &Dezgo{}
