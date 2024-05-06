package main

import (
	"archive/zip"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type extractParams struct {
	SearchRoot      string
	StartDate       string
	EndDate         string
	parsedStartDate time.Time
	parsedEndDate   time.Time
}

func (p *extractParams) ParseArgs() error {
	var err error
	if p.StartDate != "" {
		if p.parsedStartDate, err = time.Parse(time.RFC3339, p.StartDate); err != nil {
			return fmt.Errorf("invalid start date: %w", err)
		}
	}
	if p.EndDate != "" {
		if p.parsedEndDate, err = time.Parse(time.RFC3339, p.EndDate); err != nil {
			return fmt.Errorf("invalid end date: %w", err)
		}
	}
	_, err = os.Stat(p.SearchRoot)
	if err != nil {
		return fmt.Errorf("could not inspect source path: %w", err)
	}

	return nil

}

func Must[T any](v T, e error) T {
	if e != nil {
		panic(e)
	}
	return v
}

func main() {
	var opts = &extractParams{}
	flag.StringVar(&opts.SearchRoot, "source", "", "the root path to search")
	flag.StringVar(&opts.StartDate, "start", "", "start date to include")
	flag.StringVar(&opts.EndDate, "end", "", "end date to limit search to")
	flag.Parse()
	if err := opts.ParseArgs(); err != nil {
		fmt.Println(err.Error())
		flag.PrintDefaults()
		os.Exit(1)
	}

	// files := Must(getFilesToProcess(opts.SearchRoot, int64(opts.parsedStartDate.Year()), int64(opts.parsedEndDate.Year())))
	// visits := Must(getFilteredPlacesVisit(files, opts))
	cout := csv.NewWriter(os.Stdout)
	cout.Write([]string{"timestart", "timeend", "durationHours", "locationname", "locationaddress"})
	err := walkValidFiles(opts.SearchRoot, opts, func(v *placesVisit) error {
		return cout.Write([]string{
			v.Duration.Start.Format(time.DateTime),
			v.Duration.End.Format(time.DateTime),
			fmt.Sprintf("%f", v.Duration.End.Sub(v.Duration.Start).Hours()),
			v.Location.Name,
			v.Location.Address,
		})
	})
	if err != nil {
		panic(err)
	}
	cout.Flush()
	os.Exit(0)
}

func walkValidFiles(readpath string, opts *extractParams, visitor func(*placesVisit) error) error {
	info, err := os.Stat(readpath)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return walkFS(readpath, opts, visitor)
	}
	ext := path.Ext(info.Name())
	switch ext {
	case ".zip":
		return walkZip(readpath, opts, visitor)
	default:
		return fmt.Errorf("unsupported file extention: %s", ext)

	}
}

func walkFS(searchroot string, opts *extractParams, visitor func(*placesVisit) error) error {

	var (
		startyear = int64(opts.parsedStartDate.Year())
		endyear   = int64(opts.parsedEndDate.Year())
	)
	err := filepath.Walk(searchroot, func(pth string, info fs.FileInfo, err error) error {

		parts := strings.Split(path.Base(info.Name()), "_")
		yr, err := strconv.ParseInt(parts[0], 10, 16)
		//	fmt.Println(yr)
		if err == nil && !info.IsDir() && startyear <= yr && yr <= int64(endyear) {
			fin, err := os.Open(pth)
			if err != nil {
				return err
			}
			arr, err := getFilteredPlacesVisit(fin, opts)
			fin.Close()
			if err != nil {
				return err
			}
			for _, item := range arr {
				if err := visitor(item); err != nil {
					return err
				}
			}

		}
		return nil
	})

	return err
}

func walkZip(zipPath string, opts *extractParams, visitor func(*placesVisit) error) error {
	rdr, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	var (
		startyear = int64(opts.parsedStartDate.Year())
		endyear   = int64(opts.parsedEndDate.Year())
	)
	defer func() { _ = rdr.Close() }()
	for _, f := range rdr.File {
		parts := strings.Split(path.Base(f.Name), "_")
		yr, err := strconv.ParseInt(parts[0], 10, 16)
		if err == nil && startyear <= yr && yr <= int64(endyear) {
			fin, err := f.Open()
			if err != nil {
				return err
			}
			arr, err := getFilteredPlacesVisit(fin, opts)
			fin.Close()
			if err != nil {
				return err
			}
			for _, item := range arr {
				if err := visitor(item); err != nil {
					return err
				}
			}

		}
	}
	return nil
}

func getFilteredPlacesVisit(rdr io.Reader, opts *extractParams) ([]*placesVisit, error) {
	var visits []*placesVisit

	jdec := json.NewDecoder(rdr)
	var tl timeline
	if err := jdec.Decode(&tl); err != nil {
		panic(err)
	}
	for _, v_iter := range tl.TimelineObjects {
		if v_iter.Visit != nil {
			v := v_iter.Visit
			visitTime := v.Duration.Start
			if visitTime.After(opts.parsedStartDate) && visitTime.Before(opts.parsedEndDate) {
				visits = append(visits, v)
			}
		}
	}

	return visits, nil
}

func getFilesToProcess(searchroot string, startyear, endyear int64) ([]string, error) {
	var foundfiles []string
	err := filepath.Walk(searchroot, func(pth string, info fs.FileInfo, err error) error {

		parts := strings.Split(path.Base(info.Name()), "_")
		yr, err := strconv.ParseInt(parts[0], 10, 16)
		//	fmt.Println(yr)
		if err == nil && !info.IsDir() && startyear <= yr && yr <= int64(endyear) {
			fpath := pth
			foundfiles = append(foundfiles, fpath)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return foundfiles, nil
}

type timeline struct {
	TimelineObjects []*timelineObject `json:"timelineObjects"`
}

type timelineObject struct {
	Visit *placesVisit `json:"placeVisit"`
}

type placesVisit struct {
	Location *placeLocation `json:"location"`
	Duration *placeDuration `json:"duration"`
}

type placeLocation struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

type placeDuration struct {
	Start time.Time `json:"startTimestamp"`
	End   time.Time `json:"endTimestamp"`
}
