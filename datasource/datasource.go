package datasource

import (
	"fmt"
	"os"
	"path"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3" // Import to initialize driver
	log "github.com/sirupsen/logrus"
	"github.com/williamhaley/photo-server/model"
)

// Database is the general concept wrapping the organization of photos.
type Database struct {
	db   *sqlx.DB
	path string
}

// New allocates a new instance of the datasource.
func New(dataDirectory string) *Database {
	path := path.Join(dataDirectory, "database.db")

	var db *sqlx.DB

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if _, err := os.Create(path); err != nil {
			log.WithError(err).Fatalf("failed to allocate db %q", path)
		}
		db = MustOpen(path)
		if err := DestructiveReset(db); err != nil {
			log.WithError(err).Fatalf("failed to set up db %q", path)
		}
		log.Info("database created")
	} else if err != nil {
		log.WithError(err).Fatalf("failed to get db status %q", path)
	} else {
		db = MustOpen(path)
		log.Info("database opened")
	}

	return &Database{
		db: db,
	}
}

func DestructiveReset(db *sqlx.DB) error {
	_, err := db.Exec(`
		DROP TABLE IF EXISTS photos;
		CREATE TABLE photos (
			uuid VARCHAR(32) PRIMARY KEY,
			path VARCHAR(512) NOT NULL,
			name VARCHAR(64) NOT NULL,
			date DATETIME NOT NULL,
			year INTEGER NOT NULL,
			month INTEGER NOT NULL
		);
		CREATE INDEX year_index ON photos(year);
		CREATE INDEX month_index ON photos(month);
	`)

	// TODO WFH Index on year, others.

	if err != nil {
		return err
	}
	return nil
}

// MustOpen opens the DB for API access.
func MustOpen(path string) *sqlx.DB {
	db, err := sqlx.Connect("sqlite3", path)
	if err != nil {
		log.WithError(err).Fatalf("failed to open db %q", path)
	}

	return db
}

// DateBucketsForIds returns date buckets for each YYYY-MM provided. The results
// are guaranteed to be sorted in the same order as the request ids.
func (d *Database) DateBucketsForIds(ids ...string) ([]*model.YearMonthBucket, error) {
	query := squirrel.
		Select("year", "month", "uuid", `strftime("%Y-%m-%dT%H:%M:%S:%f", date) || "~" || name || "~" || uuid AS cursor`).
		From("photos")
	or := squirrel.Or{}
	for _, id := range ids {
		or = append(or, squirrel.Like{"date": id})
	}
	query.Where(or).OrderBy("cursor desc")

	sql, args, err := query.ToSql()
	if err != nil {
		log.WithError(err).Error("failed to build query for date buckets")
		return nil, err
	}

	var results []*model.Photo = make([]*model.Photo, 0)
	err = d.db.Select(&results, sql, args...)
	if err != nil {
		log.WithError(err).Error("failed to query total counts")
		return nil, err
	}

	asMap := make(map[string]*model.YearMonthBucket, len(results))
	for _, result := range results {
		key := fmt.Sprintf("%d-%d", result.Year, result.Month)
		if _, ok := asMap[key]; !ok {
			asMap[key] = &model.YearMonthBucket{
				Year:       result.Year,
				Month:      result.Month,
				Date:       key,
				PhotoUuids: []string{},
			}
		}
		asMap[key].PhotoUuids = append(asMap[key].PhotoUuids, result.UUID)
	}

	sorted := make([]*model.YearMonthBucket, len(ids))
	for index, date := range ids {
		sorted[index] = asMap[date]
	}

	return sorted, nil
}

func (d *Database) SkeletonMetaData() ([]*model.YearMonthBucket, error) {
	query := squirrel.Select("year", "month", "count(*) as total_count").From("photos").GroupBy("year", "month").OrderBy("year desc", "month desc")
	sql, args, err := query.ToSql()
	if err != nil {
		log.WithError(err).Error("failed to build query for total counts")
		return nil, err
	}

	var dateBuckets []*model.YearMonthBucket = make([]*model.YearMonthBucket, 0)
	err = d.db.Select(&dateBuckets, sql, args...)
	if err != nil {
		log.WithError(err).Error("failed to query total counts")
		return nil, err
	}

	return dateBuckets, nil
}

// PhotosCount returns the count of all photos for a given yearh and month.
func (d *Database) PhotosCount(year, month int) (int, error) {
	var count int
	err := d.db.Get(&count, "SELECT COUNT(*) FROM photos WHERE year = ? AND month = ?", year, month)
	if err != nil {
		log.WithError(err).Error("failed to query photos")
		return 0, err
	}
	return count, nil
}

// AllPhotos returns all the photos within a given range and whether or not there
// are more photos after the query.
func (d *Database) AllPhotos(year, month, limit int, after string) ([]*model.Photo, bool, error) {
	log.Debugf("[datasource.AllPhotos] year:%d month:%d limit:%d after:%q", year, month, limit, after)

	var photos []*model.Photo = make([]*model.Photo, 0)
	err := d.db.Select(&photos, `
		SELECT uuid, name, date, strftime("%Y-%m-%dT%H:%M:%S:%f", date) || "~" || name || "~" || uuid AS cursor
		FROM photos
		WHERE year = ? AND month = ? AND (? = '' OR cursor < ?)
		ORDER BY cursor DESC
		LIMIT ?
	`, year, month, after, after, limit+1)
	if err != nil {
		log.WithError(err).Error("failed to query photos")
		return nil, false, err
	}

	// When we query above we actually add 1 to the limit. If 1 more than our
	// limit came back, there's more data.
	hasMore := len(photos) > limit

	endIndex := limit
	if !hasMore {
		endIndex = len(photos)
	}

	if len(photos) > 0 {
		log.Debugf("[datasource.AllPhotos] photos slice from %d to %d", 0, endIndex)
		photos = photos[0:endIndex]
	}

	return photos, hasMore, nil
}

// AddPhoto inserts a record into the database with photo information.
func (d *Database) AddPhoto(photo *model.Photo) error {
	_, err := d.db.NamedExec(`
		INSERT INTO photos
			(uuid, path, name, date, year, month)
		VALUES
			(:uuid, :path, :name, :date, :year, :month)
	`, photo)
	if err != nil {
		log.WithError(err).Errorf("failed to insert photo %q", photo.Path)
		return err
	}
	return nil
}

// GetPhoto returns a specific photo for a given uuid.
func (d *Database) GetPhoto(uuid string) (*model.Photo, error) {
	var photo model.Photo
	err := d.db.Get(&photo, "SELECT uuid, path, date FROM photos WHERE uuid=?", uuid)
	if err != nil {
		log.WithError(err).Errorf("failed to scan photo for uuid %q", uuid)
		return nil, err
	}
	return &photo, nil
}

func (d *Database) AllPaginated(limit, offset int) ([]*model.Photo, error) {
	var photos []*model.Photo = make([]*model.Photo, 0)
	err := d.db.Select(&photos, "SELECT path, uuid FROM photos ORDER BY path DESC LIMIT ? OFFSET ?", limit, offset)
	if err != nil {
		log.WithError(err).Error("failed to load photos")
		return nil, err
	}

	return photos, err
}
