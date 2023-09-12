package logger

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	errs "github.com/pkg/errors"
)

func New(targetName string, strategy string) (*log.Logger, error) {
	name := fmt.Sprintf("%s - %s", targetName, strategy)
	var err error
	if _, err := os.Stat("../../logs"); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir("../../logs", os.ModePerm)
		if err != nil {
			return nil, errs.Wrap(err, "err while creating dir")
		}
	}

	path := filepath.Join("../../logs", name)

	newFilePath := filepath.FromSlash(path)
	file, err := os.Create(newFilePath)
	if err != nil {
		return nil, errs.Wrap(err, "err while creating file")
	}

	logger := log.New(file, "", log.LstdFlags|log.Lshortfile)

	return logger, nil
}
