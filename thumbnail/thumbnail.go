package thumbnail

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/williamhaley/goepeg"
	"github.com/williamhaley/gothumb"
	"github.com/williamhaley/photo-server/datasource"
	"github.com/williamhaley/photo-server/model"
	"io"
	"os"
	"path/filepath"
	"time"
)

// Manager tracks some state needed for generating thumbnails. The DB
// for looking up photos and the directory in which to generate photos.
type Manager struct {
	db                      *datasource.Database
	photosDirectoryRootPath string
	thumbnailsDirectoryPath string
}

// NewManager creates a new thumbnail manager.
func NewManager(db *datasource.Database, photosDirectoryRootPath, thumbnailsDirectoryPath string) *Manager {
	return &Manager{
		db:                      db,
		photosDirectoryRootPath: photosDirectoryRootPath,
		thumbnailsDirectoryPath: thumbnailsDirectoryPath,
	}
}

// Generate creates a thumbnail for a given photo. The thumbnail may or may not
// be overwritten depending on the argument. The generated (or existing) file is
// returned along with a bool indicating whether or not a thumbnail was created.
func (m *Manager) Generate(photo *model.Photo, overwrite bool) (*os.File, bool, error) {
	uuid := photo.UUID
	sourceImagePath := filepath.Join(m.photosDirectoryRootPath, photo.Path)
	created := false

	// Should mean we need to get to ~4096 photos before any directories need
	// duplicates. This allows for a reasonably (fingers crossed) wide
	// distribution of files. Not really necessary, but a nice mainteanance
	// convenience.
	partition := string(uuid[0:3])
	thumbnailDirectoryPath := filepath.Join(m.thumbnailsDirectoryPath, partition)
	if _, err := os.Stat(thumbnailDirectoryPath); os.IsNotExist(err) {
		if err := os.Mkdir(thumbnailDirectoryPath, 0755); err != nil {
			// This may look weird, but is possible with multiple workers. One
			// worker could create the dir between when we decided it didn't
			// exist and when we tried to create it.
			if !os.IsNotExist(err) {
				log.WithError(err).Errorf("error creating thumbnail directory for %q", uuid)
				return nil, false, err
			}
		}
	}
	thumbnailPath := filepath.Join(thumbnailDirectoryPath, fmt.Sprintf("%s.jpg", uuid))

	if _, err := os.Stat(thumbnailPath); overwrite || os.IsNotExist(err) {
		quality := 100
		maxSize := 200

		sourceImage, err := os.Open(sourceImagePath)
		if err != nil {
			log.WithError(err).Errorf("error opening source image %q", thumbnailPath)
			return nil, false, err
		}
		thumbnailImage, err := gothumb.Thumbnail(sourceImage, maxSize, quality, goepeg.ScaleTypeFitMax)
		if err != nil {
			log.WithError(err).Errorf("error generating thumbnail %q", uuid)
			return nil, false, err
		}
		thumbnailImageFile, err := os.OpenFile(thumbnailPath, os.O_CREATE|os.O_WRONLY, 0664)
		if err != nil {
			log.WithError(err).Errorf("error allocating thumbnail file %q", uuid)
			return nil, false, err
		}
		_, err = io.Copy(thumbnailImageFile, thumbnailImage)
		if err != nil {
			log.WithError(err).Errorf("error writing thumbnail file %q", uuid)
			return nil, false, err
		}

		created = true
	}
	file, err := os.Open(thumbnailPath)
	if err != nil {
		log.WithError(err).Errorf("error loading thumbnail %q", uuid)
		return nil, false, err
	}

	return file, created, nil
}

// GenerateAll uses the db to find all photos and create a thumbnail. The
// arguments allow skipping or overwriting existing thumbnails.
func (m *Manager) GenerateAll(overwriteExisting bool, workers int) {
	start := time.Now()
	batchStart := time.Now()
	count := 0
	skipped := 0
	batchSize := 1000
	limit := 10
	thumbnailChan := make(chan *model.Photo)

	for i := 0; i < workers; i++ {
		go func() {
			for photo := range thumbnailChan {
				file, created, err := m.Generate(photo, overwriteExisting)
				if err != nil {
					log.WithError(err).Fatal("error generating thumbnail")
				}
				if created {
					count++
					if count%batchSize == 0 {
						batchElapsedSeconds := time.Now().Sub(batchStart).Seconds()
						rate := float64(batchSize) / batchElapsedSeconds
						log.Infof("[Progress] generated %d, skipped %d, processed %d (%f/s)", count, skipped, count+skipped, rate)
						batchStart = time.Now()
					}
				} else {
					skipped++
				}
				file.Close()
			}
		}()
	}

	hasMore := true
	offset := 0
	for hasMore {
		photos, err := m.db.AllPaginated(limit, offset)
		if err != nil {
			log.WithError(err).Fatalf("error loading photos at offset %d", offset)
		}
		for _, photo := range photos {
			thumbnailChan <- photo
		}
		if len(photos) < limit {
			hasMore = false
			break
		}
		offset += limit
	}

	close(thumbnailChan)

	log.Infof("[Finished] generated %d, skipped %d, processed %d, %v seconds", count, skipped, count+skipped, time.Now().Sub(start).Seconds())
}
