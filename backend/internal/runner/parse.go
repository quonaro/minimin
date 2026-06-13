package runner

import (
	"regexp"
	"strconv"
	"strings"
)

var listResponseRe = regexp.MustCompile(`There are (\d+) of a max(?: of)? (\d+) players online[:.]?(.*)`)

// ParseListResponse extracts online count, max players, and player names from an RCON "list" response.
func ParseListResponse(resp string) (int, int, []string) {
	matches := listResponseRe.FindStringSubmatch(resp)
	if len(matches) < 4 {
		return 0, 0, []string{}
	}
	online, _ := strconv.Atoi(matches[1])
	maxPlayers, _ := strconv.Atoi(matches[2])
	names := strings.Split(matches[3], ",")
	players := make([]string, 0, len(names))
	for _, n := range names {
		if trimmed := strings.TrimSpace(n); trimmed != "" {
			players = append(players, trimmed)
		}
	}
	return online, maxPlayers, players
}

var tpsFloatRe = regexp.MustCompile(`\d+\.\d+`)

// ParseTPSOutput extracts the TPS float from an RCON "tps" response.
func ParseTPSOutput(resp string) *float64 {
	if !strings.Contains(strings.ToLower(resp), "tps") {
		return nil
	}
	m := tpsFloatRe.FindString(resp)
	if m == "" {
		return nil
	}
	v, err := strconv.ParseFloat(m, 64)
	if err != nil {
		return nil
	}
	return &v
}
