package downloader

import (
	"fmt"
	"time"
	"bufio"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

type NOAAModel struct {
	model                         string
	openDataDeliveryOffsetMinutes int
	intervalHours                 int
	urlFormat                     string
	res                           string
	maxStep                       map[int]int
	breakPoint                    int
}

var noaaModels = map[string]NOAAModel{
	"gfs": {
		model:                         "gfs",
		openDataDeliveryOffsetMinutes: 360,
		intervalHours:                 6,
		urlFormat:                     "https://noaa-gfs-bdp-pds.s3.amazonaws.com/gfs.%s/%s/atmos/gfs.t%sz.pgrb2.%s.f%s",
		res:                           "0p25",
		maxStep: map[int]int{
			0:  384,
			6:  384,
			12: 384,
			18: 384,
		},
		breakPoint: 120,
	},
}

type NOAADownloader struct {
	modelName    string
	param        string
	height       string
	outputFolder string
	tmpFolder    string
	maxStep      int
	modelDetails NOAAModel
	httpClient   *http.Client
}

type NOAADownloaderOptions struct {
	ModelName    string
	Param        string
	Height       string
	OutputFolder string
	MaxStep      int
	ModelDetails NOAAModel
}

type IndexData map[string]map[string]struct {
	Start int
	End   int
}

func NewNOAADownloader(options NOAADownloaderOptions) *NOAADownloader {

	tmpFolder := "/tmp/gribdl/noaa"

	if _, err := os.Stat(tmpFolder); os.IsNotExist(err) {
		os.MkdirAll(tmpFolder, 0755)
	}

	return &NOAADownloader{
		modelName:    options.ModelName,
		param:        options.Param,
		height:       options.Height,
		outputFolder: options.OutputFolder,
		tmpFolder:    tmpFolder,
		maxStep:      options.MaxStep,
		modelDetails: options.ModelDetails,
		httpClient:   &http.Client{Timeout: 60 * time.Second},
	}
}

func (wdp *NOAADownloader) getGribFileUrl(step int, timestamp time.Time) string {

	run := timestamp.Format("20060102")
	stepStr := fmt.Sprintf("%03d", step)
	intervalGroupStr := fmt.Sprintf("%02d", timestamp.Hour())

	return fmt.Sprintf(wdp.modelDetails.urlFormat, run, intervalGroupStr, intervalGroupStr, wdp.modelDetails.res, stepStr)
}

func (wdp *NOAADownloader) getMostRecentModelTimestamp() time.Time {
	offset := time.Duration(-wdp.modelDetails.openDataDeliveryOffsetMinutes) * time.Minute
	return time.Now().UTC().Add(offset).Truncate(time.Duration(wdp.modelDetails.intervalHours) * time.Hour)
}

func (wdp *NOAADownloader) getIndexFile(url string) (IndexData, error) {
	resp, err := http.Get(fmt.Sprintf("%s.idx", url))
	if err != nil {
		return nil, fmt.Errorf("error fetching index: %v", err)
	}
	defer resp.Body.Close()

	result := make(IndexData)
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")

		if len(parts) < 5 {
			continue
		}

		fileType := parts[3]
		height := parts[4]

		start, err := strconv.Atoi(parts[1])
		if err != nil {
			continue
		}
		var end int
		if scanner.Scan() {
			nextLine := scanner.Text()
			nextParts := strings.Split(nextLine, ":")
			if len(nextParts) > 1 {
				end, err = strconv.Atoi(nextParts[1])
				if err != nil {
					continue
				}
			}
		}

		// Store the data
		if _, ok := result[fileType]; !ok {
			result[fileType] = make(map[string]struct {
				Start int
				End   int
			})
		}
		result[fileType][height] = struct {
			Start int
			End   int
		}{start, end}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading index file: %v", err)
	}

	return result, nil
}

func (wdp *NOAADownloader) downloadAndProcessFile(url string, index IndexData, param string, retries int) error {
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", index[param][wdp.height].Start, index[param][wdp.height].End))

	resp, err := wdp.httpClient.Do(req)

	if err != nil {
		if retries > 0 {
			log.Println("[DL] Retrying...")
			return wdp.downloadAndProcessFile(url, index, param, retries-1)
		}
		return fmt.Errorf("[DL] getting url: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPartialContent {
		if retries > 0 {
			log.Println("[DL] Retrying...")
			return wdp.downloadAndProcessFile(url, index, param, retries-1)
		}
	}

	filePath := filepath.Join(wdp.outputFolder, filepath.Base(url))
	outputFile, err := os.Create(filePath + "_" + param + "_" + wdp.height + ".grib2")
	if err != nil {
		return fmt.Errorf("[DL] creating file: %w", err)
	}
	defer outputFile.Close()

	if _, err = io.Copy(outputFile, resp.Body); err != nil {
		if retries > 0 {
			log.Println("[DL] Retrying...")
			return wdp.downloadAndProcessFile(url, index, param, retries-1)
		}
		return fmt.Errorf("[DL] copying file: %w", err)
	}

	return nil
}

func StartNOAADownloader(options NOAADownloaderOptions) map[int][]byte {
	modelDetails, exists := noaaModels[options.ModelName]
	if !exists {
		log.Println("[MAIN] Model not found. Available models are:")
		for key := range noaaModels {
			log.Println("-", key)
		}
		return nil
	}

	options.ModelDetails = modelDetails

	wdp := NewNOAADownloader(options)

	timestamp := wdp.getMostRecentModelTimestamp()

	log.Printf("[MAIN] Processing %s model for parameter %s up to %d steps starting from %s\n", wdp.modelName, wdp.param, wdp.maxStep, timestamp)
	params := strings.Split(wdp.param, ",")

	var wg sync.WaitGroup
	errors := make(chan error, wdp.maxStep*len(params))

	if wdp.maxStep > options.ModelDetails.maxStep[timestamp.Hour()] {
		wdp.maxStep = options.ModelDetails.maxStep[timestamp.Hour()]
	}

	firstLoop := wdp.maxStep

	if wdp.maxStep >= wdp.modelDetails.breakPoint {
		firstLoop = wdp.modelDetails.breakPoint
	}

	for step := 0; step < firstLoop; step++ {
		wg.Add(1)
		go func(step int, params []string) {
			defer wg.Done()

			url := wdp.getGribFileUrl(step, timestamp)

			index, err := wdp.getIndexFile(url)
			if err != nil {
				log.Println(err)
				return
			}

			for _, param := range params {

				err = wdp.downloadAndProcessFile(url, index, param, 5)
				if err != nil {
					errors <- err
					return
				}
			}
		}(step, params)
	}

	for step := wdp.modelDetails.breakPoint; step <= wdp.maxStep; step += 3 {
		wg.Add(1)
		go func(step int, params []string) {
			defer wg.Done()

			url := wdp.getGribFileUrl(step, timestamp)

			index, err := wdp.getIndexFile(url)
			if err != nil {
				log.Println(err)
				return
			}

			for _, param := range params {

				err = wdp.downloadAndProcessFile(url, index, param, 5)
				if err != nil {
					errors <- err
					return
				}
			}
		}(step, params)
	}

	wg.Wait()
	return nil
}
