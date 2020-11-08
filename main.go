package main

import (
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/williamhaley/photo-server/datasource"
	"github.com/williamhaley/photo-server/indexer"
	"github.com/williamhaley/photo-server/server"
	"github.com/williamhaley/photo-server/thumbnail"
	"os"
)

var errorInvalidThumbnailDirectory = fmt.Errorf("'-thumbnails-directory path' must reference a valid directory")
var errorInvalidDatabasePath = fmt.Errorf("'-database path' must be defined")

func init() {
	log.SetLevel(log.DebugLevel)
}

func main() {
	if len(os.Args) < 2 {
		helpAndExit()
	}

	switch os.Args[1] {
	case "index":
		indexCommand := flag.NewFlagSet("index", flag.ExitOnError)
		photosDirectoryRootPath := indexCommand.String("photos-directory", ".", "Directory to search for indexing photos")
		generateThumbnails := indexCommand.Bool("thumbnails", true, "Whether or not to generate thumbnails while indexing")
		thumbnailsDirectoryPath := indexCommand.String("thumbnails-directory", "", "Directory to use for thumbnail")
		dbPath := indexCommand.String("database", "", "Path to database file")
		numWorkers := indexCommand.Int("workers", 1, "Number of workers for index processing")

		indexCommand.Parse(os.Args[2:])

		err := index(os.ExpandEnv(*dbPath), os.ExpandEnv(*photosDirectoryRootPath), *generateThumbnails, os.ExpandEnv(*thumbnailsDirectoryPath), *numWorkers)
		if err != nil {
			fmt.Println(err)
			fmt.Println()
			indexCommand.PrintDefaults()
		}
	case "thumbnails":
		thumbnailsCommand := flag.NewFlagSet("thumbnails", flag.ExitOnError)
		photosDirectoryRootPath := thumbnailsCommand.String("photos-directory", ".", "Root directory for all photos")
		thumbnailsDirectoryPath := thumbnailsCommand.String("thumbnails-directory", "", "Directory to use for thumbnail")
		overwriteExisting := thumbnailsCommand.Bool("overwrite-existing", false, "Whether or not to clobber existing thumbnails")
		dbPath := thumbnailsCommand.String("database", "", "Path to database file")
		numWorkers := thumbnailsCommand.Int("workers", 1, "Number of workers for generating thumbnails")

		thumbnailsCommand.Parse(os.Args[2:])

		err := thumbnails(os.ExpandEnv(*dbPath), os.ExpandEnv(*photosDirectoryRootPath), os.ExpandEnv(*thumbnailsDirectoryPath), *overwriteExisting, *numWorkers)
		if err != nil {
			fmt.Println(err)
			fmt.Println()
			thumbnailsCommand.PrintDefaults()
		}
	case "serve":
		serveCommand := flag.NewFlagSet("serve", flag.ExitOnError)
		photosDirectoryRootPath := serveCommand.String("photos-directory", ".", "Root directory for all photos")
		thumbnailsDirectoryPath := serveCommand.String("thumbnails-directory", "", "Directory to use for thumbnail")
		httpPort := serveCommand.String("http-port", "8080", "Port to server the app over HTTP")
		dbPath := serveCommand.String("database", "", "Path to database file")

		serveCommand.Parse(os.Args[2:])

		err := serve(os.ExpandEnv(*dbPath), os.ExpandEnv(*photosDirectoryRootPath), os.ExpandEnv(*thumbnailsDirectoryPath), *httpPort)
		if err != nil {
			fmt.Println(err)
			fmt.Println()
			serveCommand.PrintDefaults()
		}
	default:
		helpAndExit()
	}
}

func helpAndExit() {
	fmt.Println("expected 'index', 'serve', or 'thumbnails' subcommands")
	os.Exit(1)
}

func index(dbPath, photosDirectoryRootPath string, generateThumbnails bool, thumbnailsDirectoryPath string, numWorkers int) error {
	if dbPath == "" {
		return errorInvalidDatabasePath
	}
	db := datasource.New(datasource.MustCreate(dbPath))

	var thumbnailManager *thumbnail.Manager
	if generateThumbnails {
		if err := validateThumbnailConfig(thumbnailsDirectoryPath); err != nil {
			return err
		}
		thumbnailManager = thumbnail.NewManager(db, photosDirectoryRootPath, thumbnailsDirectoryPath)
	}

	log.Infof("index photos in %q", photosDirectoryRootPath)

	indexer := indexer.New(db, photosDirectoryRootPath, thumbnailManager, numWorkers)
	indexer.Scan()

	return nil
}

func thumbnails(dbPath, photosDirectoryRootPath, thumbnailsDirectoryPath string, overwriteExisting bool, numWorkers int) error {
	if err := validateThumbnailConfig(thumbnailsDirectoryPath); err != nil {
		return err
	}
	if dbPath == "" {
		return errorInvalidDatabasePath
	}
	db := datasource.New(datasource.MustOpen(dbPath))

	log.Infof("generating thumbnails with %d worker(s)", numWorkers)

	thumbnailManager := thumbnail.NewManager(db, photosDirectoryRootPath, thumbnailsDirectoryPath)
	thumbnailManager.GenerateAll(overwriteExisting, numWorkers)

	return nil
}

func serve(dbPath, photosDirectoryRootPath, thumbnailsDirectoryPath, httpPort string) error {
	if err := validateThumbnailConfig(thumbnailsDirectoryPath); err != nil {
		return err
	}
	if dbPath == "" {
		return errorInvalidDatabasePath
	}
	db := datasource.New(datasource.MustOpen(dbPath))

	log.Infof("starting http server on port %q", httpPort)

	thumbnailManager := thumbnail.NewManager(db, photosDirectoryRootPath, thumbnailsDirectoryPath)

	server := server.New(db, photosDirectoryRootPath, thumbnailManager, httpPort)
	return server.Start()
}

func validateThumbnailConfig(thumbnailsDirectoryPath string) error {
	thumbnailsDirectory, err := os.Stat(os.ExpandEnv(thumbnailsDirectoryPath))
	if err != nil {
		return errorInvalidThumbnailDirectory
	}
	if thumbnailsDirectoryPath == "" || !thumbnailsDirectory.IsDir() {
		return errorInvalidThumbnailDirectory
	}

	return nil
}
