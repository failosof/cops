package opening

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/failosof/cops/util"
)

const DatabaseURL = "https://raw.githubusercontent.com/lichess-org/chess-openings/refs/heads/master/"

var filenames = [...]string{
	"a.tsv",
	"b.tsv",
	"c.tsv",
	"d.tsv",
	"e.tsv",
}

func Cached(dir string) (cached bool) {
	for _, filename := range filenames {
		info, err := os.Stat(filepath.Join(dir, filename))
		if err != nil || info.Size() == 0 {
			return false
		}
	}
	return true
}

func DownloadDatabase(dir string) ([]string, error) {
	var wg sync.WaitGroup
	files := make([]string, len(filenames))
	errsCh := make(chan error, len(filenames))
	defer close(errsCh)

	for i, filename := range filenames {
		files[i] = filepath.Join(dir, filename)
		url := DatabaseURL + filename

		wg.Add(1)
		go func(from, to string) {
			defer wg.Done()
			if err := util.Download(url, files[i]); err != nil {
				errsCh <- fmt.Errorf("failed to download openings db: %w", err)
			}
		}(files[i], url)
	}

	wg.Wait()

	if len(errsCh) > 0 {
		errs := make([]error, len(errsCh))
		for err := range errsCh {
			errs = append(errs, err)
		}
		return nil, errors.Join(errs...)
	}

	return files, nil
}
