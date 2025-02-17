package app

import "github.com/failosof/cops/cache"

type Files struct {
	OpeningsDatabaseDir string
	OpeningsIndexFile   string
	PuzzlesDatabaseFile string
	PuzzlesIndexFile    string
}

func DefaultFiles() Files {
	return Files{
		OpeningsDatabaseDir: cache.PathTo("database"),
		OpeningsIndexFile:   cache.PathTo("openings.index"),
		PuzzlesDatabaseFile: cache.PathTo("puzzles.csv.zst"),
		PuzzlesIndexFile:    cache.PathTo("puzzles.index"),
	}
}
