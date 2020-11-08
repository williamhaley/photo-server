package debug

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	_ "net/http/pprof" // Standard pprof import
	"runtime"
)

// Serve for pprof access
func Serve(port string) {
	go func() {
		log.Println(http.ListenAndServe(fmt.Sprintf("localhost:%s", port), nil))
	}()
}

// PrintMemUsage prints the current memory usage
func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	log.Infof("Alloc = %v MiB", bToMb(m.Alloc))
	log.Infof("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	log.Infof("\tSys = %v MiB", bToMb(m.Sys))
	log.Infof("\tHeapInuse = %v MiB", bToMb(m.HeapInuse))
	log.Infof("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
