package wasp

import (
	"fmt"
	"net/url"
	"path"
	"regexp"
	"strings"

	"go.uber.org/zap"
)

type Playlist struct {
	m3u8Playlist []string
}

func (p *Playlist) UpdatePlaylist(base string, data []byte) {
	if len(p.m3u8Playlist) == 0 {
		lines := strings.Split(string(data), "\n")[5:]

		p.m3u8Playlist = setBasePath(base, lines)
		return
	}

	// headers := getHeaderLines(data)
	newlines := getLastLines(data)
	toAdd := setBasePath(base, newlines)

	// zap.S().Infof("adding new lines to playlist: '%s'", newlines)
	// TODO: check file exists first
	// p.m3u8Playlist = append(headers, p.m3u8Playlist[4:]...)
	p.m3u8Playlist = append(p.m3u8Playlist, toAdd...)
}

func (p *Playlist) Print() []byte {
	headers := []string{
		"#EXTM3U",
		"#EXT-X-VERSION:3",
		fmt.Sprintf("#EXT-X-MEDIA-SEQUENCE:%d", len(p.m3u8Playlist)),
		"#EXT-X-ALLOW-CACHE:YES",
		fmt.Sprintf("#EXT-X-TARGETDURATION:%d", len(p.m3u8Playlist)),
	}
	output := append(headers, p.m3u8Playlist...)
	return []byte(strings.Join(output, "\n"))
}

func getHeaderLines(data []byte) []string {
	lines := strings.Split(string(data), "\n")
	return lines[:4]
}

func getLastLines(data []byte) []string {
	lines := strings.Split(string(data), "\n")
	return lines[len(lines)-3 : len(lines)-1]
}

func setBasePath(base string, lines []string) []string {
	newlines := []string{}
	baseUrl, err := url.Parse(base)
	if err != nil {
		zap.S().Errorf("problem parsing base lines: %s", err)
		return []string{}
	}
	for _, l := range lines {
		match, err := regexp.Match("output[0-9]+.ts", []byte(l))
		if err != nil {
			zap.S().Errorf("problem checking playlist lines: %s", err)
			return []string{}
		}
		if match {
			u := baseUrl
			u.Path = path.Join(u.Path, l)
			l = u.String()
		}
		newlines = append(newlines, l)
	}

	return newlines
}
