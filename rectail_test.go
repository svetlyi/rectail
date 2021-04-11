package rectail_test

import (
	"context"
	"github.com/svetlyi/rectail"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

var logOutput = ioutil.Discard

func TestWatch(t *testing.T) {
	logger := log.New(logOutput, "", 0)

	var (
		fileUpdate    rectail.FileUpdate
		expectedLines = [][]string{
			{"first"},
			{"second"},
			{"third"},
			{
				"џѕѓ bar foo bar foo bar foo bar foo bar foo bar foo bar foo bar foo bar џѕѓ bar",
				"џѕѓ1 bar1 foo1 bar1 foo1 bar1 foo1 bar1 foo1 bar1 foo1 bar1 foo1 bar1 foo1 bar1 foo1 bar1 џѕѓ1 bar1",
				"џѕѓ2 bar2 foo2 bar2 foo2 bar2 foo2 bar2 foo2 bar2 foo2 bar2 foo2 bar2 foo2 bar2 foo2 bar2 џѕѓ2 bar2",
				"џѕѓ3 bar3 foo3 bar3 foo3 bar3 foo3 bar3 foo3 bar3 foo3 bar3 foo3 bar3 foo3 bar3 foo3 bar3 џѕѓ3 bar3",
				"џѕѓ4 bar4 foo4 bar4 foo4 bar4 foo4 bar4 foo4 bar4 foo4 bar4 foo4 bar4 foo4 bar4 foo4 bar4 џѕѓ4 bar4",
			},
		}
		textToWrite = []string{
			"first\n",
			"second\n",
			"third",
			`
џѕѓ bar foo bar foo bar foo bar foo bar foo bar foo bar foo bar foo bar џѕѓ bar
џѕѓ1 bar1 foo1 bar1 foo1 bar1 foo1 bar1 foo1 bar1 foo1 bar1 foo1 bar1 foo1 bar1 foo1 bar1 џѕѓ1 bar1
џѕѓ2 bar2 foo2 bar2 foo2 bar2 foo2 bar2 foo2 bar2 foo2 bar2 foo2 bar2 foo2 bar2 foo2 bar2 џѕѓ2 bar2
џѕѓ3 bar3 foo3 bar3 foo3 bar3 foo3 bar3 foo3 bar3 foo3 bar3 foo3 bar3 foo3 bar3 foo3 bar3 џѕѓ3 bar3
џѕѓ4 bar4 foo4 bar4 foo4 bar4 foo4 bar4 foo4 bar4 foo4 bar4 foo4 bar4 foo4 bar4 foo4 bar4 џѕѓ4 bar4`,
		}
		fileUpdates = make(chan rectail.FileUpdate)
		updates     = make(chan string)
		err         error
		f           *os.File
	)
	f, err = os.Create("test_data/test.log")
	if err != nil {
		t.Fatal(err)
	}

	rt, err := rectail.NewRecTail(
		[]string{"test_data"},
		[]string{"\\.log"},
		updates,
		600,
		100,
		logger,
	)
	if err != nil {
		t.Fatalf("could not initialize rectail: %v", err)
	}
	rectailCtx, rectailCtxCancel := context.WithCancel(context.Background())
	defer rectailCtxCancel()

	go func() {
		if err = rt.Watch(rectailCtx, fileUpdates); err != nil {
			logger.Fatalln(err)
		}
	}()
	go func() {
		for update := range updates {
			logger.Println(update)
		}
	}()
	for i := range textToWrite {
		if _, err = f.WriteString(textToWrite[i]); err != nil {
			t.Fatal(err)
		}
		fileUpdate = <-fileUpdates
		if len(fileUpdate.Lines) != len(expectedLines[i]) {
			t.Errorf("expected %d lines with update; got: %d", len(expectedLines[i]), len(fileUpdate.Lines))
		}
		for lineIndex := range fileUpdate.Lines {
			if fileUpdate.Lines[lineIndex] != expectedLines[i][lineIndex] {
				t.Errorf("expected: %s; got: %s", expectedLines[i][lineIndex], fileUpdate.Lines[lineIndex])
			}
		}
	}
}

func TestFirstRead(t *testing.T) {
	logger := log.New(logOutput, "", 0)

	var (
		fileUpdate    rectail.FileUpdate
		expectedLines = []string{
			"џѕѓ bar foo bar foo bar foo bar foo bar foo bar foo bar foo bar foo bar џѕѓ bar",
			"џѕѓ1 bar1 foo1 bar1 foo1 bar1 foo1 bar1 foo1 bar1 foo1 bar1 foo1 bar1 foo1 bar1 foo1 bar1 џѕѓ1 bar1",
			"џѕѓ2 bar2 foo2 bar2 foo2 bar2 foo2 bar2 foo2 bar2 foo2 bar2 foo2 bar2 foo2 bar2 foo2 bar2 џѕѓ2 bar2",
			"џѕѓ3 bar3 foo3 bar3 foo3 bar3 foo3 bar3 foo3 bar3 foo3 bar3 foo3 bar3 foo3 bar3 foo3 bar3 џѕѓ3 bar3",
			"џѕѓ4 bar4 foo4 bar4 foo4 bar4 foo4 bar4 foo4 bar4 foo4 bar4 foo4 bar4 foo4 bar4 foo4 bar4 џѕѓ4 bar4",
		}
		textToWrite = `џѕѓ bar foo bar foo bar foo bar foo bar foo bar foo bar foo bar foo bar џѕѓ bar
џѕѓ1 bar1 foo1 bar1 foo1 bar1 foo1 bar1 foo1 bar1 foo1 bar1 foo1 bar1 foo1 bar1 foo1 bar1 џѕѓ1 bar1
џѕѓ2 bar2 foo2 bar2 foo2 bar2 foo2 bar2 foo2 bar2 foo2 bar2 foo2 bar2 foo2 bar2 foo2 bar2 џѕѓ2 bar2
џѕѓ3 bar3 foo3 bar3 foo3 bar3 foo3 bar3 foo3 bar3 foo3 bar3 foo3 bar3 foo3 bar3 foo3 bar3 џѕѓ3 bar3
џѕѓ4 bar4 foo4 bar4 foo4 bar4 foo4 bar4 foo4 bar4 foo4 bar4 foo4 bar4 foo4 bar4 foo4 bar4 џѕѓ4 bar4`
		fileUpdates = make(chan rectail.FileUpdate)
		updates     = make(chan string)
		err         error
		f           *os.File
	)
	f, err = os.Create("test_data/test.log")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.WriteString(textToWrite); err != nil {
		t.Fatal(err)
	}

	rt, err := rectail.NewRecTail(
		[]string{"test_data"},
		[]string{"\\.log"},
		updates,
		600,
		600,
		logger,
	)
	if err != nil {
		t.Fatalf("could not initialize rectail: %v", err)
	}
	rectailCtx, rectailCtxCancel := context.WithCancel(context.Background())
	defer rectailCtxCancel()

	go func() {
		if err = rt.Watch(rectailCtx, fileUpdates); err != nil {
			logger.Fatalln(err)
		}
	}()
	go func() {
		for update := range updates {
			logger.Println(update)
		}
	}()
	fileUpdate = <-fileUpdates
	if len(fileUpdate.Lines) != len(expectedLines) {
		t.Errorf("expected %d lines with update; got: %d", len(expectedLines), len(fileUpdate.Lines))
	}
	for lineIndex := range fileUpdate.Lines {
		if fileUpdate.Lines[lineIndex] != expectedLines[lineIndex] {
			t.Errorf("expected: %s; got: %s", expectedLines[lineIndex], fileUpdate.Lines[lineIndex])
		}
	}
}
