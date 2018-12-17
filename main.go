package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

var (
	directory string
	dryRun    bool
	verbose   bool
)

func vlogf(s string, vv ...interface{}) {
	if !verbose {
		return
	}
	log.Printf(s, vv...)
}

func vlogln(vv ...interface{}) {
	if !verbose {
		return
	}
	log.Println(vv...)
}

func main() {
	flag.StringVar(&directory, "dir", "", "The directory to manage")
	flag.BoolVar(&dryRun, "dry-run", false, "don't actually make the changes, just say what would be done")
	flag.BoolVar(&verbose, "verbose", false, "verbose logging")

	flag.Parse()

	if directory == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	dir, err := os.Open(directory)
	if err != nil {
		log.Fatalln(err)
	}

	d, err := dir.Stat()
	if err != nil {
		log.Fatalf("could not stat directory: %v\n", err)
	}
	if !d.IsDir() {
		log.Fatalf(`"%v" is not a directory\n`, directory)
	}

	dirlist, err := dir.Readdir(0)
	if err != nil {
		log.Fatalf("could not read directory: %v\n", err)
	}

	for _, f := range dirlist {

		// We leave directories alone
		if f.IsDir() {
			vlogf(`"%s" is a directory, skipping\n`, f.Name())
			continue
		}

		// If it's been modified in the last 7 days, leave it alone
		if f.ModTime().After(time.Now().Add(-7 * 24 * time.Hour)) {
			vlogf(`"%v" has been modified within the last 7 days, skipping\n`, f.Name())
			continue
		}

		// We want to put the files in a directory of the month it was last modified in the formal {year}-{month}
		// if it doesn't exist, we create the destination directory
		destDirName := fmt.Sprintf("%v/%v-%v", directory, f.ModTime().Year(), f.ModTime().Month())
		if !dryRun {
			if _, err := os.Stat(destDirName); os.IsNotExist(err) {
				if err := os.Mkdir(destDirName, os.ModeDir|0777); err != nil {
					log.Fatal(err.Error())
				}
			}
		}

		logStr := fmt.Sprintf(`Moving "%v" to "%v"`, f.Name(), destDirName)
		// If dryrun, then log what we would have done and continue to the next file
		if dryRun {
			log.Println(logStr)
			continue
		}

		// Actually move the file to the destination directory
		vlogln(logStr)
		if err := os.Rename(directory+"/"+f.Name(), destDirName+"/"+f.Name()); err != nil {
			log.Fatal(err.Error())
		}
	}

}
