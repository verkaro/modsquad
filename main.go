// modsquad: batch export tracker modules to audio formats using xmp, flac, and lame
package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
)

var (
	outDir    string
	format    string
	recursive bool
	// track current temp file for cleanup on interrupt
	currentTempFile string
)

func init() {
	flag.StringVar(&outDir, "out", "out", "Output directory")
	flag.StringVar(&format, "format", "mp3", "Output format: wav, flac, mp3")
	flag.BoolVar(&recursive, "recursive", false, "Recurse into directories")
}

func main() {
	flag.Parse()
	inputs := flag.Args()
	if len(inputs) == 0 {
		log.Fatal("Usage: modsquad [options] <file or dir>...")
	}

	// Verify required tools are in PATH
	tools := []string{"xmp"}
	switch format {
	case "flac":
		tools = append(tools, "flac")
	case "mp3":
		tools = append(tools, "lame")
	case "wav":
		// no extra tool
	default:
		log.Fatalf("Unknown format: %s", format)
	}
	for _, tool := range tools {
		if _, err := exec.LookPath(tool); err != nil {
			log.Fatalf("Required tool '%s' not found in PATH", tool)
		}
	}

	// setup interrupt handler to clean temp file
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		log.Println("Interrupted, cleaning up...")
		if currentTempFile != "" {
			os.Remove(currentTempFile)
		}
		os.Exit(1)
	}()

	// create base output directory
	if err := os.MkdirAll(outDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	for _, in := range inputs {
		info, err := os.Stat(in)
		if err != nil {
			log.Printf("Skipping %s: %v", in, err)
			continue
		}
		if info.IsDir() {
			if recursive {
				root := filepath.Base(in)
				filepath.WalkDir(in, func(path string, d fs.DirEntry, err error) error {
					if err != nil {
						log.Printf("Error reading %s: %v", path, err)
						return nil
					}
					if d.IsDir() {
						return nil
					}
					rel, err := filepath.Rel(in, path)
					if err != nil {
						rel = filepath.Base(path)
					}
					rel = filepath.Join(root, rel)
					processFile(path, rel)
					return nil
				})
			} else {
				log.Printf("Skipping directory %s (use -recursive)", in)
			}
		} else {
			processFile(in, filepath.Base(in))
		}
	}
}

// processFile exports a single module at inputPath, preserving relPath under outDir
func processFile(inputPath, relPath string) {
	base := relPath[:len(relPath)-len(filepath.Ext(relPath))]
	outRel := base + "." + format
	outPath := filepath.Join(outDir, outRel)

	// skip if output already exists
	if _, err := os.Stat(outPath); err == nil {
		log.Printf("Skipping %s: output already exists at %s", inputPath, outPath)
		return
	} else if !os.IsNotExist(err) {
		log.Printf("Error checking %s: %v", outPath, err)
	}

	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		log.Printf("Failed to create dir for %s: %v", outPath, err)
		return
	}

	// create temporary WAV file
	tmpWav, err := os.CreateTemp("", "modsquad-*.wav")
	if err != nil {
		log.Printf("Failed to create temp wav file: %v", err)
		return
	}
	tmpWav.Close()
	// track for cleanup
	currentTempFile = tmpWav.Name()
	defer func() {
		os.Remove(tmpWav.Name())
		currentTempFile = ""
	}()

	// export with xmp
	cmd := exec.Command("xmp", "-o", tmpWav.Name(), inputPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Printf("xmp failed for %s: %v", inputPath, err)
		return
	}

	// convert to desired format
	switch format {
	case "wav":
		if err := os.Rename(tmpWav.Name(), outPath); err != nil {
			log.Printf("Failed to move wav file: %v", err)
		}
	case "flac":
		cmd := exec.Command("flac", "-o", outPath, tmpWav.Name())
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Printf("flac failed for %s: %v", inputPath, err)
		}
	case "mp3":
		// use verbose MP3 encoding without silent flag to show progress
		cmd := exec.Command("lame", "--vbr-new", "-V", "6", tmpWav.Name(), outPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Printf("lame failed for %s: %v", inputPath, err)
		}
	}

	fmt.Printf("Processed %s -> %s\n", inputPath, outPath)
}

