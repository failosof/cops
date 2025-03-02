package main

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/notnil/chess"
	"golang.org/x/time/rate"
)

const (
	MaxExportNumber = 300
	ExportGamesURL  = "https://lichess.org/games/export/_ids"
)

func ExportFromLichess(ctx context.Context, ids []string) ([]*chess.Game, error) {
	body := strings.NewReader(strings.Join(ids, ","))
	req, err := http.NewRequest(http.MethodPost, ExportGamesURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to construct a request: %w", err)
	}

	req.Header.Set("Accept", "application/x-chess-pgn")

	resp, err := perform(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to request lichess: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected lichess response: %s", resp.Status)
	}

	data := bytes.NewBuffer(make([]byte, 0, 300*1024))
	if _, err := data.ReadFrom(resp.Body); err != nil {
		return nil, fmt.Errorf("failed to read lichess response: %w", err)
	}

	pgns := bytes.Split(data.Bytes(), []byte("\n\n\n"))
	games := make([]*chess.Game, 0, len(ids))

	for _, pgn := range pgns {
		if len(pgn) > 0 {
			opt, err := chess.PGN(bytes.NewReader(pgn))
			if err != nil {
				return nil, fmt.Errorf("failed to parse pgn: %w", err)
			}
			games = append(games, chess.NewGame(opt))
		}
	}

	return games, nil
}

var limit = 3 * time.Second
var limiter = rate.NewLimiter(rate.Every(limit), 1)

func incLimit() {
	limit += time.Second
	limiter.SetLimit(rate.Every(limit))
}

func perform(ctx context.Context, req *http.Request) (*http.Response, error) {
	for i := 0; i < 3; i++ {
		if err := limiter.Wait(ctx); err != nil {
			return nil, err
		}

		slog.Debug("lichess api request", "url", req.URL.String())
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, err
		}
		slog.Debug("lichess api responded", "status", resp.Status)

		if resp.StatusCode != http.StatusTooManyRequests {
			return resp, nil
		}

		time.Sleep(time.Minute)
	}

	return nil, fmt.Errorf("rate limited")
}
