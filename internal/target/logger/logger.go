package logger

import (
	"fmt"
	"log"
	"os"
)

func New(targetName string, strategy string) (*log.Logger, error) {
	name := fmt.Sprintf("%s - %s", targetName, strategy)
	var logpath = fmt.Sprintf("logs/%s.log", name)

	var file, err = os.Create(logpath)
	if err != nil {
		return nil, err
	}
	logger := log.New(file, "", log.LstdFlags|log.Lshortfile)
	log.Println("created LogFile : " + logpath)

	return logger, nil
}
