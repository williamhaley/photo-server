package analyzer

import (
	"fmt"
	"github.com/dsoprea/go-exif/v3"
	"github.com/dsoprea/go-exif/v3/common"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"time"
)

// AnalysisInfo contains the result from analyzing a photo.
type AnalysisInfo struct {
	Date  *time.Time
	Path  string
	Error error
}

// Analyze takes in a path to a photo and will send the result to the Analyzer's
// results channel.
func Analyze(path string, resultsChan chan *AnalysisInfo) {
	analysisInfo := &AnalysisInfo{
		Path: path,
	}
	date, err := getDateForPhoto(path)
	if err != nil {
		analysisInfo.Error = err
	} else {
		analysisInfo.Date = date
	}

	resultsChan <- analysisInfo
}

func getDateForPhoto(path string) (*time.Time, error) {
	date, err := getDateFromExif(path)
	if err != nil && err != io.EOF {
		log.WithError(err).Warnf("failed to get date from exif %q", path)
	}
	if date != nil {
		return date, nil
	}

	date, err = getDateFromFile(path)
	if err != nil {
		// Should never happen.
		log.WithError(err).Fatalf("failed to get date from file %q", path)
		return nil, err
	}
	if date != nil {
		return date, nil
	}

	return nil, fmt.Errorf("no date found for %s", path)
}

func findAnyExifTag(index exif.IfdIndex) string {
	var tagEntry *exif.IfdTagEntry
outer:
	for _, value := range index.Ifds {
		for _, tagName := range []string{"DateTimeOriginal", "DateTime"} {
			tagEntries, err := value.FindTagWithName(tagName)
			if err != nil && err.Error() != "tag not found" && err.Error() != "tag is not known" {
				for _, value := range index.Ifds {
					value.PrintTagTree(true)
				}
			}
			if len(tagEntries) > 1 {
				log.Fatalf("found multiple %q tag entries, which should be impossible", tagName)
			}

			if len(tagEntries) == 0 {
				continue
			}

			tagEntry = tagEntries[0]
			break outer
		}
	}

	// Didn't find anything.
	if tagEntry == nil {
		return ""
	}

	valueRaw, err := tagEntry.Value()
	if err != nil {
		log.WithError(err).Fatal("error parsing value from tag entry")
	}

	return valueRaw.(string)
}

func getDateFromExif(path string) (*time.Time, error) {
	rawExif, err := exif.SearchFileAndExtractExif(path)
	if err != nil {
		if err.Error() == "no exif data" {
			return nil, err
		}
		log.WithError(err).Fatalf("error searching for exif data in file")
	}
	im, err := exifcommon.NewIfdMappingWithStandard()
	if err != nil {
		log.WithError(err).Fatalf("error creating mapping")
	}
	tagIndex := exif.NewTagIndex()
	_, index, err := exif.Collect(im, tagIndex, rawExif)
	if err != nil {
		log.WithError(err).Fatalf("error collecting exif structure")
	}

	dateString := findAnyExifTag(index)

	for _, format := range []string{"2006:01:02 15:04:05", "2006:01:02 15:04: 5", "2006:01:02 15:04", "2006:01:02"} {
		date, err := time.Parse(format, dateString)
		if err != nil {
			continue
		}
		return &date, nil
	}

	return nil, fmt.Errorf("no exif date found for %s", path)
}

func getDateFromFile(path string) (*time.Time, error) {
	stats, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	date := stats.ModTime()
	return &date, nil
}
