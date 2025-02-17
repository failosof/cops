package puzzle

import (
	"fmt"

	"github.com/failosof/cops/util"
)

const DatabaseURL = "https://database.lichess.org/lichess_db_puzzle.csv.zst"

func DownloadDatabase(to string) error {
	if err := util.Download(DatabaseURL, to); err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	return nil
}
