package indexer

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/williamhaley/photo-server/analyzer"
	"github.com/williamhaley/photo-server/datasource"
	"github.com/williamhaley/photo-server/model"
	"github.com/williamhaley/photo-server/thumbnail"
)

type Indexer struct {
	db                      *datasource.Database
	photosDirectoryRootPath string
	thumbnailManager        *thumbnail.Manager
	batchSize               int
	numWorkers              int
}

func New(db *datasource.Database, photosDirectoryRootPath string, thumbnailManager *thumbnail.Manager, numWorkers int) *Indexer {
	return &Indexer{
		db:                      db,
		photosDirectoryRootPath: photosDirectoryRootPath,
		thumbnailManager:        thumbnailManager,
		batchSize:               1000,
		numWorkers:              numWorkers,
	}
}

func (i *Indexer) Scan() {
	// https://blog.golang.org/pipelines
	analysisInfoChan := i.fileProcessor()
	thumbnailChan := i.analysisInfoProcessor(analysisInfoChan)
	progressChan := i.thumbnailProcessor(thumbnailChan)

	start := time.Now()
	batchStart := time.Now()

	var total int
	// Not running this in a gorouting. We want to actually block and wait for
	// this to finish. Alternatively, use a wait group or similar mechanism to
	// block for completion.
	for progress := range progressChan {
		if progress%i.batchSize == 0 {
			rate := float64(i.batchSize) / time.Now().Sub(batchStart).Seconds()
			log.Infof("[completed] %d (%f/s)", progress, rate)
			batchStart = time.Now()
		}
		total = progress
	}
	log.Infof("[done] scanned %d photos in %v seconds", total, time.Now().Sub(start))

	log.Info("done")
}

func (i *Indexer) fileProcessor() <-chan *analyzer.AnalysisInfo {
	out := make(chan *analyzer.AnalysisInfo)

	go func() {
		err := filepath.Walk(i.photosDirectoryRootPath, func(photoPath string, info os.FileInfo, err error) error {
			if err != nil {
				log.WithError(err).Fatal("unexpected error")
				return err
			}
			if info.IsDir() {
				return nil
			}

			switch filepath.Ext(photoPath) {
			case ".jpg", ".JPG", ".JPEG", ".jpeg":
				analyzer.Analyze(os.ExpandEnv(photoPath), out)
			}
			return nil
		})
		if err != nil && err.Error() != "EOF" {
			log.WithError(err).Fatalf("error walking directory")
			return
		}

		close(out)
	}()

	return out
}

func (i *Indexer) analysisInfoProcessor(in <-chan *analyzer.AnalysisInfo) <-chan *model.Photo {
	out := make(chan *model.Photo)
	waitGroup := sync.WaitGroup{}
	total := 0

	for j := 0; j < i.numWorkers; j++ {
		waitGroup.Add(1)
		go func() {
			for analysisInfo := range in {
				if analysisInfo.Error != nil {
					log.WithError(analysisInfo.Error).Error("error!")
					return
				}

				date := analysisInfo.Date
				// Use a relative path to make data more portable if a user wants
				// to re-home their server at any point.
				relativePath, err := filepath.Rel(i.photosDirectoryRootPath, analysisInfo.Path)
				if err != nil {
					log.WithError(err).Fatalf("failed to resolve photo relative path %q %q", analysisInfo.Path, i.photosDirectoryRootPath)
				}

				photo := model.NewPhoto(date, relativePath)
				err = i.db.AddPhoto(photo)
				if err != nil {
					log.WithError(err).Fatalf("failed to index photo %q", photo.Path)
				}
				total++
				if total%i.batchSize == 0 {
					log.Infof("[photos] %d processed", total)
				}
				if true {
					out <- photo
				}
			}
			waitGroup.Done()
		}()
	}

	go func() {
		waitGroup.Wait()
		close(out)
		log.Infof("[photos] indexed %d", total)
	}()

	return out
}

func (i *Indexer) thumbnailProcessor(in <-chan *model.Photo) <-chan int {
	out := make(chan int)
	waitGroup := sync.WaitGroup{}

	overwrite := false
	thumbnailsCreated := 0
	thumbnailsSkipped := 0

	for j := 0; j < i.numWorkers; j++ {
		waitGroup.Add(1)
		go func() {
			for photo := range in {
				file, created, err := i.thumbnailManager.Generate(photo, overwrite)
				if err != nil {
					log.WithError(err).Fatalf("failed to generate thumbnail during indexing %q", photo.Path)
				}
				if created {
					thumbnailsCreated++
				} else {
					thumbnailsSkipped++
				}
				total := thumbnailsCreated + thumbnailsSkipped
				if total%i.batchSize == 0 {
					log.Infof("[thumbnails] %d processed", total)
				}
				file.Close()
				out <- total
			}
			waitGroup.Done()
		}()
	}

	go func() {
		waitGroup.Wait()
		close(out)
		log.Infof("[thumbnails] created %d, skipped %d", thumbnailsCreated, thumbnailsSkipped)
	}()

	return out
}
