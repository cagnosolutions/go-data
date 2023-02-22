package buffer

import (
	"errors"

	"github.com/cagnosolutions/go-data/pkg/engine/storage"
)

var (
	ErrBadConfig = errors.New("config is incorrect")
	ErrNoConfig  = errors.New("config missing")
)

type Config struct {
	PageCount uint16
	Replacer
	storage.Storer
}

func checkConfig(conf *Config) error {
	if conf == nil {
		return ErrNoConfig
	}
	if conf.Replacer == nil {
		return ErrBadConfig
	}
	if conf.Storer == nil {
		return ErrBadConfig
	}
	if conf.PageCount < 1 {
		// set a min page count of 4 for now
		conf.PageCount = 4
	}
	if conf.PageCount > 4096 {
		// set a max page count of 4096 for now
		conf.PageCount = 4096
	}
	return nil
}
