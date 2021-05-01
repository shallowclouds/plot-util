package ls

import (
	"strconv"
	"strings"
	"time"
)

const (
	TypeDir  = "directory"
	TypeFile = "file"
)

type File struct {
	Name    string
	Path    string
	Size    int64
	ModTime time.Time
	Type    string
}

type Parser func(string) ([]*File, error)

func NameCommand(path string) ([]string, Parser) {
	return []string{path}, func(raw string) ([]*File, error) {
		splits := strings.Split(raw, " ")
		files := make([]*File, 0, len(splits))
		for _, split := range splits {
			files = append(files, &File{
				Name:    split,
				Path:    "",
				Size:    0,
				ModTime: time.Time{},
				Type:    "",
			})
		}

		return files, nil
	}
}

func FullCommand(path string) ([]string, Parser) {
	return []string{"-l", path}, func(raw string) ([]*File, error) {
		lines := strings.Split(raw, "\n")
		files := make([]*File, 0, len(lines))
		for _, line := range lines {
			rawParts := strings.Split(line, " ")
			parts := make([]string, 0, len(rawParts))
			for _, part := range rawParts {
				if part != "" {
					parts = append(parts, part)
				}
			}

			if len(parts) != 9 {
				continue
			}

			size, _ := strconv.ParseInt(parts[4], 10, 64)
			file := &File{
				Name:    parts[8],
				Path:    "",
				Size:    size,
				ModTime: time.Time{},
				Type:    "",
			}
			files = append(files, file)
		}

		return files, nil
	}
}
