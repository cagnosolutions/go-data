package tools

import (
	"embed"
)

//go:embed *_file.json
var fp embed.FS

func LoadFileData(name string) []byte {
	b, err := fp.ReadFile(name)
	if err != nil {
		panic(err)
	}
	return b
}
