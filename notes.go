package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func createNote(rootDir, note, tags string) {
	now := time.Now()
	today := now.Format("2006-01-02")
	dir := filepath.Join(rootDir, today)

	ls, err := ioutil.ReadDir(dir)
	if err != nil && !os.IsNotExist(err) {
		log.Fatalf("Failed to get %v files list: %v", dir, err)
	}
	nextNote := fmt.Sprintf("%02d", len(ls))

	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalf("Failed to create today's dir: %v", err)
	}
	notePath := filepath.Join(dir, nextNote)
	f, err := os.Create(notePath)
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()
	f.WriteString(now.Format("2006-01-02 15:04:05|"))
	f.WriteString(note)
	f.Write([]byte("|tags: "))
	f.WriteString(tags)
	f.Write([]byte("\n"))

	for _, tag := range strings.Split(tags, ",") {
		if tag == "" {
			continue
		}
		tagDir := filepath.Join(rootDir, "tags", tag)
		if err := os.MkdirAll(tagDir, 0755); err != nil {
			log.Fatalf("Failed to create tag %v dir: %v", tag, err)
		}
		err := os.Symlink(
			filepath.Join("..", "..", today, nextNote),
			filepath.Join(tagDir, today+"-"+nextNote))
		if err != nil {
			log.Fatalf("Failed to link note to tag %v dir: %v", tag, err)
		}
	}
}

func readFiles(dir string) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to list %v files: %v", dir, err)
	}
	for i := len(files) - 1; i >= 0; i-- {
		data, err := ioutil.ReadFile(filepath.Join(dir, files[i].Name()))
		if err != nil {
			return fmt.Errorf("failed to read %v: %v", files[i].Name(), err)
		}
		fmt.Print(string(data))
	}
	return nil
}
func showNotes(rootDir, fromDay, tags string) error {
	if tags != "" {
		for _, tag := range strings.Split(tags, ",") {
			if err := readFiles(filepath.Join(rootDir, "tags", tag)); err != nil {
				return err
			}
		}
		return nil
	}

	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	alldirs, err := ioutil.ReadDir(rootDir)
	if err != nil {
		return fmt.Errorf("failed to list all dirs: %v", err)
	}
	lastDay := today.Add(-time.Hour * 24)
	if fromDay != "" {
		lastDay, err = time.ParseInLocation("2006-01-02", fromDay, time.Local)
		if err != nil {
			return fmt.Errorf("failed to parse fromDay %v: %v", fromDay, err)
		}
	}

	for i := len(alldirs) - 1; i >= 0; i-- {
		if alldirs[i].Name() == "tags" {
			continue
		}
		t, err := time.ParseInLocation("2006-01-02", alldirs[i].Name(), time.Local)
		if err != nil {
			return fmt.Errorf("failed to parse date from dir %v: %v", alldirs[i].Name(), err)
		}
		if fromDay != "" && t.Before(lastDay) {
			break
		} else if fromDay == "" && t.After(lastDay) && !t.Equal(today) {
			continue
		}
		if err := readFiles(filepath.Join(rootDir, alldirs[i].Name())); err != nil {
			return fmt.Errorf("failed to read files from %v: %v", alldirs[i].Name(), err)
		}
		if fromDay == "" && !t.Equal(today) {
			break
		}
	}
	return nil
}

func main() {
	var (
		rootDir = flag.String("root-dir", "", "Root dir path")
		tags    = flag.String("t", "", "Note tags")
		fromDay = flag.String("from-day", "", "Show all notes since given day")
	)
	flag.Parse()
	if *rootDir == "" {
		hd, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Failed to get user home dir: %v", err)
		}
		*rootDir = filepath.Join(hd, ".notes")
	}

	if len(flag.Args()) == 0 {
		showNotes(*rootDir, *fromDay, *tags)
		return
	}

	createNote(*rootDir, strings.Join(flag.Args(), " "), *tags)
}
