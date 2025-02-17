package app

import (
	"log/slog"
	"os"

	"github.com/failosof/cops/opening"
	"github.com/failosof/cops/puzzle"
	"github.com/failosof/cops/util"
)

type State struct {
	Config   Config
	Files    Files
	Log      *slog.Logger
	Openings *opening.Index
	Puzzles  *puzzle.Index
}

func (s *State) LoadOpenings() error {
	util.Assert(s.Log != nil, "app log must be set")

	dir := s.Files.OpeningsDatabaseDir
	index := s.Files.OpeningsIndexFile

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		s.Log.Warn("failed to create openings database dir", "in", dir)
		return err
	}

	if !opening.Cached(dir) {
		s.Log.Info("downloading openings database", "from", opening.DatabaseURL, "to", dir)
		if _, err := opening.DownloadDatabase(dir); err != nil {
			s.Log.Warn("failed to download openings database")
			return err
		}
		s.Log.Info("removing old openings index", "from", index)
		if err := util.RemoveFile(index); err != nil {
			s.Log.Warn("failed to remove old openings index")
			return err
		}
	}

	var err error
	if util.FileExists(index) {
		s.Openings, err = util.LoadBinary[opening.Index](index)
		if err != nil {
			s.Log.Warn("failed to load openings index")
			return err
		}
		s.Log.Info("loaded openings index", "from", index, "size", len(s.Openings.Names))
	} else {
		s.Log.Info("creating openings index", "from", dir)
		s.Openings, err = opening.CreateIndex(dir)
		if err != nil {
			s.Log.Warn("failed to create openings index")
			return err
		}
		s.Log.Info("saving openings index", "to", index)
		if err := util.SaveBinary(index, s.Openings); err != nil {
			s.Log.Warn("failed to save openings index")
			return err
		}
	}

	return nil
}

func (s *State) LoadPuzzles() error {
	util.Assert(s.Log != nil, "app log must be set")

	db := s.Files.PuzzlesDatabaseFile
	index := s.Files.PuzzlesIndexFile

	if !util.FileExists(db) {
		s.Log.Info("downloading puzzles database", "from", puzzle.DatabaseURL, "to", db)
		if err := puzzle.DownloadDatabase(db); err != nil {
			s.Log.Warn("failed to download puzzles database")
			return err
		}
		s.Log.Info("removing old puzzles index", "from", index)
		if err := util.RemoveFile(index); err != nil {
			s.Log.Warn("failed to remove old puzzles index")
			return err
		}
	}

	var err error
	if util.FileExists(index) {
		s.Puzzles, err = util.LoadBinary[puzzle.Index](index)
		if err != nil {
			s.Log.Warn("failed to load puzzles index")
			return err
		}
		s.Log.Info("loaded puzzles index", "from", index, "size", len(s.Puzzles.Collection))
	} else {
		s.Log.Info("creating puzzles index", "from", db)
		s.Puzzles, err = puzzle.CreateIndex(db)
		if err != nil {
			s.Log.Warn("failed to create puzzles index")
			return err
		}
		s.Log.Info("saving puzzles index", "to", index)
		if err := util.SaveBinary(index, s.Puzzles); err != nil {
			s.Log.Warn("failed to save puzzles index", "err", err)
			return err
		}
	}

	return nil
}

func (s *State) SearchOpening() (name opening.Name) {
	util.Assert(s.Log != nil, "app log must be set")
	util.Assert(s.Openings != nil, "openings must be loaded")

	positions := s.Config.Game.Positions()
	i := len(positions) - 1
	var candidate opening.Name
	var pos util.Position
	for ; i >= 0; i-- {
		pos = util.PositionFromChess(positions[i])
		candidate = s.Openings.Search(pos)
		if !candidate.Empty() {
			break
		}
	}
	if i == 0 {
		s.Log.Warn("no opening found")
		return
	}

	s.Log.Info("found opening", "family", candidate.Family(), "variation", candidate.Variation())
	name = candidate

	return
}

func (s *State) SearchPuzzles(openingName opening.Name) (puzzles []puzzle.Data) {
	util.Assert(s.Log != nil, "app log must be set")
	util.Assert(s.Puzzles != nil, "puzzles must be loaded")

	puzzles = s.Puzzles.Search(openingName.Tag(), s.Config.Turn, s.Config.MinMoves, s.Config.MaxMoves)
	if len(puzzles) == 0 {
		s.Log.Warn("no puzzles found")
		return nil
	}
	s.Log.Info("found puzzles", "count", len(puzzles))

	// todo: filter out puzzles by moves after the opening

	return
}
