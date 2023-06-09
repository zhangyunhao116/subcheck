package main

import (
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/zhangyunhao116/skipset"
)

var delete = flag.Bool("d", false, "Delete all videos that don't have subtitles.")

func main() {
	flag.Parse()

	var err error
	curpath, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	allfiles := skipset.NewString()
	err = filepath.Walk(curpath, func(inputpath string, info os.FileInfo, err error) error {
		if err == nil && isVideo(inputpath) {
			allfiles.Add(inputpath)
			if info.Size() == 0 {
				panic("invalid size:" + inputpath)
			}
		}
		if err != nil {
			panic(err)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	nosubtitle := skipset.NewString()
	var wg sync.WaitGroup
	for i := 0; i < runtime.GOMAXPROCS(-1); i++ {
		wg.Add(1)
		go func() {
			allfiles.Range(func(value string) bool {
				if allfiles.Remove(value) {
					cmd := "ffprobe " + `"` + value + `"`
					command := exec.Command("bash", "-c", cmd)
					out, err := command.CombinedOutput()
					if err != nil {
						panic(err.Error() + "\n" + string(out))
					}
					ffprobeInfo := string(out)
					if !strings.Contains(ffprobeInfo, "Subtitle:") {
						nosubtitle.Add(value)
						println(value)
					}
				}
				return true
			})
			wg.Done()
		}()
	}
	wg.Wait()

	if nosubtitle.Len() != 0 {
		if *delete {
			nosubtitle.Range(func(value string) bool {
				command := exec.Command("bash", "-c", "rm "+`"`+value+`"`)
				out, err := command.CombinedOutput()
				if err != nil {
					panic("Delete error: " + err.Error() + "\n" + string(out))
				}
				println("Deleted ", value)
				return true
			})
		} else {
			println("Failed to fetch subtitle: ", nosubtitle.Len())
		}
	} else {
		println("Success!")
	}
}

func isVideo(path string) bool {
	var videos = []string{".mp4", ".mkv", ".webm"}
	for _, v := range videos {
		if strings.HasSuffix(path, v) {
			return true
		}
	}
	return false
}
