package lichess

import "regexp"

const (
	SiteURL        = "https://lichess.org/"
	PuzzleURL      = SiteURL + "training/"
	ExportGamesURL = SiteURL + "games/export/_ids"
)

var gameRe = regexp.MustCompile(`lichess\.org/([a-zA-Z0-9]+)`)

func Puzzle(id string) string {
	return PuzzleURL + id
}

func GameID(url string) string {
	matches := gameRe.FindStringSubmatch(url)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func Game(id string) string {
	return SiteURL + id
}
