package lichess

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/failosof/cops/util"
	"github.com/notnil/chess"
)

func ExportGames(ids []string) ([]*chess.Game, error) {
	util.Assert(len(ids) <= 300, "can't export more than 300 games")

	body := strings.NewReader(strings.Join(ids, ","))
	req, err := http.NewRequest(http.MethodPost, ExportGamesURL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to construct a request: %w", err)
	}

	req.Header.Set("Accept", "application/x-chess-pgn")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to request lichess: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected lichess response: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read lichess response: %w", err)
	}

	pgns := bytes.Split(data, []byte("\n\n\n"))
	games := make([]*chess.Game, len(ids))
	//errCh := make(chan error)

	// todo: use semaphore
	// todo: check errors
	var wg sync.WaitGroup
	for i, pgn := range pgns {
		if len(pgn) > 0 {
			wg.Add(1)
			go func(i int, pgn []byte) {
				defer wg.Done()
				opt, err := chess.PGN(bytes.NewReader(pgn))
				if err != nil {
					//errCh <- err
					return
				}
				games[i] = chess.NewGame(opt)
			}(i, pgn)
		}
	}
	wg.Wait()

	return games, nil
}
