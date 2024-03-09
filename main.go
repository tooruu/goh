package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/tooruu/goh/pkg/handlers"

	"github.com/hashicorp/go-retryablehttp"
)

const DEFAULT_LINKS_FILE = "links.txt"

var (
	reportedUnsupportedSites = make([]string, 0, 10)
	logNew                   = log.New(os.Stdout, "[ADDED]", 0)
	logUpdate                = log.New(os.Stdout, "[UPDATE]", 0)
	logUnknown               = log.New(os.Stderr, "[UNKNOWN]", 0)
	logError                 = log.New(os.Stderr, "[ERROR]", 0)
	httpClient               = retryablehttp.NewClient()
	handlerMap               = map[string]func(*http.Response) (time.Time, error){
		"f95zone.to":  handlers.F95zoneHandler,
		"kemono.su":   handlers.KemonoHandler,
		"twitter.com": handlers.TwitterHandler,
		// "chan.sankakucomplex.com": handlers.SankakuHandler,
	}
	wg sync.WaitGroup
)

func parseLine(line string) (url string, urlHost string, lastChecked time.Time, err error) {
	url, lastCheckedStr, timeIsSet := strings.Cut(line, " ")

	urlParts := strings.Split(url, "/")
	if len(urlParts) < 3 {
		err = fmt.Errorf("invalid URL: %s", url)
		return
	}
	urlHost = strings.TrimPrefix(urlParts[2], "www.")

	if timeIsSet {
		lastChecked, err = time.Parse(time.DateTime, lastCheckedStr)
	}
	return
}

func processLine(line string, output io.Writer) {
	defer wg.Done()
	defer func() { fmt.Fprintln(output, strings.TrimSpace(line)) }()
	link, site, lastChecked, err := parseLine(line)
	if err != nil {
		logError.Printf("(%v) %s\n", err, link)
		return
	}

	handler, ok := handlerMap[site]
	if !ok {
		if !slices.Contains(reportedUnsupportedSites, site) {
			reportedUnsupportedSites = append(reportedUnsupportedSites, site)
		}
		logUnknown.Printf("(%s) %s\n", site, link)
		return
	}
	resp, err := httpClient.Get(link)
	if err != nil {
		logError.Printf("(%v) %s\n", err, link)
		return
	}
	latestUpd, err := handler(resp)
	if err != nil {
		logError.Printf("(%v) %s\n", err, link)
		return
	}
	handleUpdates(link, lastChecked, latestUpd)
	line = fmt.Sprintln(link, latestUpd.Format(time.DateTime))
}

func handleUpdates(res string, old time.Time, new time.Time) {
	if old.IsZero() {
		logNew.Printf("(%s) %s\n", new.Format(time.DateTime), res)
	} else if new.Compare(old) == 1 {
		logUpdate.Printf("(%s->%s) %s\n", old.Format(time.DateTime), new.Format(time.DateTime), res)
	}
}

func processFile(r io.Reader, w io.Writer) error {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		wg.Add(1)
		go processLine(sc.Text(), w)
	}
	if err := sc.Err(); err != nil {
		return err
	}
	wg.Wait()
	return nil
}

func getIntendedLinksFile() (fname string, err error) {
	if len(os.Args) > 1 {
		fname = os.Args[1]
	} else {
		fname = DEFAULT_LINKS_FILE
	}
	stat, err := os.Lstat(fname)
	if err != nil {
		return
	}
	if stat.IsDir() {
		err = fmt.Errorf("%s is a directory", fname)
	}
	return
}

func main() {
	log.SetFlags(0)
	linksFname, err := getIntendedLinksFile()
	ensure(err)
	source, err := os.Open(linksFname)
	ensure(err)
	temp, err := os.CreateTemp(filepath.Dir(linksFname), "links")
	ensure(err)
	httpClient.Logger = nil
	if err = processFile(source, temp); err != nil {
		source.Close()
		temp.Close()
		log.Fatalf("[FATAL](%v)\n", err)
	}
	ensure(source.Close())
	ensure(temp.Close())
	ensure(os.Rename(temp.Name(), linksFname))
}

func ensure(e error) {
	if e != nil {
		log.Fatalf("[FATAL](%v)\n", e)
	}
}
