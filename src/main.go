package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"hstin-de/gribdl/downloader"
)

const (
	defaultParam   = "t_2m"
	defaultMaxStep = 10
	defaultOutput  = "output"
	defaultHeight  = "surface"
)

var (
	supportedDWDModels  = []string{"icon", "icon-d2", "icon-eu"}
	supportedNOAAModels = []string{"gfs"}
)

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func printHelp() {
	fmt.Println("\nUsage: gribdl [command] [options]")
	fmt.Println("\nCommands:")
	fmt.Printf("  dwd [model<%s>]\n", strings.Join(supportedDWDModels, "|"))
	fmt.Printf("  noaa [model<%s>]\n\n", strings.Join(supportedNOAAModels, "|"))
	fmt.Println("Options:")
	fmt.Println("  --param string")
	fmt.Println("        Parameter name (default \"t_2m\")")
	fmt.Println("  --maxStep int")
	fmt.Println("        Max download steps (default 10)")
	fmt.Println("  --output string")
	fmt.Println("        Output folder (default \"output\")")
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	help := flag.Bool("help", false, "Show help")
	flag.Parse()

	if *help {
		printHelp()
		os.Exit(0)
	}

	firstTwoArgs := os.Args[1:3]

	if len(firstTwoArgs) < 2 {
		fmt.Println("Error: Missing command or model")
		printHelp()
		os.Exit(1)
	}

	os.Args = os.Args[2:]

	command := firstTwoArgs[0]
	model := firstTwoArgs[1]

	outputFolder := flag.String("output", defaultOutput, "Output folder")
	param := flag.String("param", defaultParam, "Parameter name")
	maxStep := flag.Int("maxStep", defaultMaxStep, "Max download steps")
	height := flag.String("height", defaultHeight, "Height (only for NOAA)")
	flag.Parse()

	if _, err := os.Stat(*outputFolder); os.IsNotExist(err) {
		os.MkdirAll(*outputFolder, 0755)
	}

	switch command {
	case "dwd":
		if !contains(supportedDWDModels, model) {
			fmt.Println("Error: Unsupported model:", model)
			printHelp()
			os.Exit(1)
		}
		downloader.StartDWDDownloader(downloader.DWDOpenDataDownloaderOptions{
			ModelName:    model,
			Param:        *param,
			OutputFolder: *outputFolder,
			MaxStep:      *maxStep,
			Regrid:       true,
		})

	case "noaa":
		if !contains(supportedNOAAModels, model) {
			fmt.Println("Error: Unsupported model", model)
			printHelp()
			os.Exit(1)
		}

		downloader.StartNOAADownloader(downloader.NOAADownloaderOptions{
			ModelName:    model,
			Param:        *param,
			Height:       *height,
			OutputFolder: *outputFolder,
			MaxStep:      *maxStep,
		})

	default:
		fmt.Println("Error: Unrecognized command, expected 'dwd' or 'noaa'")
		os.Exit(1)
	}
}
