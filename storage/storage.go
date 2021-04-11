package storage

import (
	"github.com/pkg/errors"
	"log"
	"sync"
)

type Storage struct {
	files []WatchedFile
	sync.Mutex
	logger *log.Logger
}

func NewStorage(logger *log.Logger) Storage {
	return Storage{
		logger: logger,
	}
}

func (s *Storage) AddFileIfNotExist(fileToAdd WatchedFile) {
	s.Lock()
	defer s.Unlock()
	for _, f := range s.files {
		if fileToAdd.FullFilePath == f.FullFilePath {
			return
		}
	}
	s.files = append(s.files, fileToAdd)
	s.logger.Println("added file", fileToAdd.FullFilePath)
}

func (s *Storage) RemoveFile(fileToRemove WatchedFile) error {
	s.Lock()
	defer s.Unlock()

	var (
		i int
		f WatchedFile
	)
	for i, f = range s.files {
		if f.FullFilePath == fileToRemove.FullFilePath {
			s.files = append(s.files[:i], s.files[i+1:]...)
			s.logger.Printf("file %s was removed", fileToRemove.FullFilePath)
			return nil
		}
	}

	return errors.Errorf("%s can't be removed: file not found", fileToRemove.FullFilePath)
}

// List returns a copy of the slice to not to go crazy with mutexes
func (s *Storage) List() []WatchedFile {
	s.Lock()
	defer s.Unlock()
	return append([]WatchedFile(nil), s.files...)
}

func (s *Storage) Update(fileToUpdate WatchedFile) error {
	s.Lock()
	defer s.Unlock()

	for i := range s.files {
		if s.files[i].FullFilePath == fileToUpdate.FullFilePath {
			s.files[i] = fileToUpdate
			s.logger.Printf("file %s was updated", fileToUpdate.FullFilePath)
			return nil
		}
	}
	return errors.Errorf("%s can't be updated: file not found", fileToUpdate.FullFilePath)
}

type WatchedFile struct {
	FullFilePath string
	LastOffset   int64
	PreviousSize int64
	BeingWatched bool
}
