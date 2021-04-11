package rectail

import (
	"bufio"
	"context"
	"github.com/pkg/errors"
	"github.com/svetlyi/rectail/storage"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

type RecTail struct {
	// startWith - files or directories to start from, for example /tmp/
	startWith []string
	// regexpsToWatch - a set of regexps that are used to watch dirs
	// for example:
	// \/tmp\/somedir1\/.*
	// somedir2\/.*
	regexpsToWatch   []regexp.Regexp
	updates          chan<- string
	delayMillisecond int64
	maxOffset        int64
	logger           *log.Logger
	storage          storage.Storage
}

func NewRecTail(
	startWithDirs []string,
	regexpsToWatch []string,
	updates chan<- string,
	delayMillisecond int64,
	maxOffset int64,
	logger *log.Logger,
) (RecTail, error) {
	var err error
	for i := range startWithDirs {
		if startWithDirs[i], err = filepath.Abs(startWithDirs[i]); err != nil {
			return RecTail{}, errors.Wrapf(err, "could not get absolute path for %s", startWithDirs[i])
		}
	}
	logger.Println("starting with", strings.Join(startWithDirs, ","))
	var (
		regexps   = make([]regexp.Regexp, len(regexpsToWatch))
		curRegexp *regexp.Regexp
	)
	for i, reg := range regexpsToWatch {
		if curRegexp, err = regexp.Compile(reg); err != nil {
			return RecTail{}, errors.Wrapf(err, "could not compile regular expression %s", reg)
		}
		logger.Println("added regexp", reg)
		regexps[i] = *curRegexp
	}
	return RecTail{
		startWith:        startWithDirs,
		regexpsToWatch:   regexps,
		updates:          updates,
		delayMillisecond: delayMillisecond,
		maxOffset:        maxOffset,
		storage:          storage.NewStorage(logger),
		logger:           logger,
	}, nil
}

// Watch - starts watching for changes and putting them to the channel
func (rt *RecTail) Watch(ctx context.Context, lines chan<- FileUpdate) error {
	defer rt.stop()
	var (
		err         error
		updateErrCh = make(chan error)
		readErrCh   = make(chan error)
	)

	go rt.updateFileListToWatch(ctx, updateErrCh)
	go rt.readFiles(ctx, readErrCh, lines)

	select {
	case err = <-readErrCh:
		return err
	case err = <-updateErrCh:
		return err
	case <-ctx.Done():
		return nil
	}
}

func (rt *RecTail) stop() {
	rt.logger.Println("stopped watching")
	close(rt.updates)
}

func (rt *RecTail) updateFileListToWatch(ctx context.Context, errCh chan<- error) {
	var (
		stat       os.FileInfo
		err        error
		usedRegexp *regexp.Regexp
	)
	defer close(errCh)
	for {
		for pathIndex, entity := range rt.startWith {
			select {
			case <-ctx.Done():
				return
			default:
			}
			if len(rt.regexpsToWatch) > pathIndex {
				usedRegexp = &rt.regexpsToWatch[pathIndex]
			} else {
				usedRegexp = nil
			}

			if stat, err = os.Stat(entity); err != nil {
				errCh <- errors.Wrapf(err, "could not get stat for %s", entity)
				return
			}
			if !stat.IsDir() {
				rt.storage.AddFileIfNotExist(storage.WatchedFile{FullFilePath: entity})
				continue
			}
			err = filepath.Walk(entity, func(fullpath string, info fs.FileInfo, err error) error {
				if err != nil {
					return errors.Wrapf(err, "error walking starting with dir %s", entity)
				}
				if !info.IsDir() && (usedRegexp == nil || usedRegexp.MatchString(fullpath)) {
					rt.storage.AddFileIfNotExist(storage.WatchedFile{FullFilePath: fullpath})
				}
				return nil
			})
			if err != nil {
				errCh <- errors.Wrapf(err, "could not walk in %s", entity)
			}
		}
		time.Sleep(time.Millisecond * time.Duration(rt.delayMillisecond))
	}
}

func (rt *RecTail) readFiles(ctx context.Context, errCh chan<- error, linesCh chan<- FileUpdate) {
	for {
		for _, f := range rt.storage.List() {
			select {
			case <-ctx.Done():
				return
			default:
			}
			if err := rt.readLines(f, linesCh); err != nil {
				errCh <- errors.Wrapf(err, "could not read lines in %s", f.FullFilePath)
				return
			}
		}
		time.Sleep(time.Millisecond * time.Duration(rt.delayMillisecond))
	}
}

// readLines reads lines starting from where it stopped the last time. If
// it's a new file to watch, it reads the last (FileSize - MaxOffset) bytes
func (rt *RecTail) readLines(f storage.WatchedFile, linesCh chan<- FileUpdate) error {
	var (
		stat os.FileInfo
		err  error
	)
	file, err := os.Open(f.FullFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return rt.storage.RemoveFile(f)
		}
		return errors.Wrapf(err, "could not open file %s", f.FullFilePath)
	}
	defer func(file *os.File) {
		errClose := file.Close()
		if errClose != nil {
			rt.logger.Printf("error while closing file %s: %v", f.FullFilePath, errClose)
		}
	}(file)

	if stat, err = file.Stat(); err != nil {
		return errors.Wrap(err, "could not open file")
	} else if !f.BeingWatched && stat.Size() > rt.maxOffset {
		// if it is a new file, just read the last (stat.Size() - rt.maxOffset) bytes
		f.LastOffset = stat.Size() - rt.maxOffset
	}
	// file was not changed
	if f.PreviousSize == stat.Size() {
		return nil
	}

	if _, err = file.Seek(f.LastOffset, io.SeekStart); err != nil {
		return errors.Wrapf(err, "could set position in file with offset %d", f.LastOffset)
	}

	var (
		scanner      = bufio.NewScanner(file)
		linesCounter = 0
		fu           = FileUpdate{
			FileName:     file.Name(),
			FullFilePath: f.FullFilePath,
			Lines:        make([]string, 0),
		}
	)

	for scanner.Scan() {
		// if we are scanning the file for the first time (it is a new one),
		// skip the first line as it might be incomplete
		if f.LastOffset == 0 || linesCounter != 0 {
			fu.Lines = append(fu.Lines, scanner.Text())
		}
		linesCounter++
		f.LastOffset += int64(len(scanner.Bytes()))
	}
	if len(fu.Lines) > 0 {
		linesCh <- fu
	}
	f.PreviousSize = stat.Size()
	f.BeingWatched = true
	if err = rt.storage.Update(f); err != nil {
		return errors.Wrapf(err, "could not update information about file %s", f.FullFilePath)
	}

	return nil
}

type FileUpdate struct {
	FileName     string
	FullFilePath string
	Lines        []string
}
