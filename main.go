package main

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pdfcpu/pdfcpu/pkg/api"
)

const (
	uploadDir = "uploads"
	outputDir = "temp_outputs"
	zipName   = "split_files.zip"
)

func main() {
	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()
	r.SetTrustedProxies(nil)
	r.LoadHTMLGlob("templates/*")

	// Ensure upload directory exists
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		panic(fmt.Sprintf("Failed to create upload directory: %s", err))
	}

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	r.POST("/split", handleSplitPDF)
	r.Run(":8000")
}

// handleSplitPDF handles the PDF splitting process
func handleSplitPDF(c *gin.Context) {
	file, err := c.FormFile("pdf")
	if err != nil {
		c.String(http.StatusBadRequest, "Failed to upload file: %s", err.Error())
		return
	}

	names := c.PostForm("names")
	nameList := strings.Split(strings.TrimSpace(names), "\n")

	src := filepath.Join(uploadDir, file.Filename)
	if err := c.SaveUploadedFile(file, src); err != nil {
		c.String(http.StatusInternalServerError, "Failed to save file: %s", err.Error())
		return
	}

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		c.String(http.StatusInternalServerError, "Failed to create output directory: %s", err.Error())
		return
	}

	if err := api.SplitFile(src, outputDir, 1, nil); err != nil {
		c.String(http.StatusInternalServerError, "Failed to split PDF: %s", err.Error())
		return
	}

	files, err := filepath.Glob(filepath.Join(outputDir, "*.pdf"))
	if err != nil || len(files) == 0 {
		c.String(http.StatusInternalServerError, "No files found after splitting.")
		return
	}

	sortFilesByPageNumber(&files)

	for i, f := range files {
		newName := fmt.Sprintf("part%d.pdf", i+1)
		if i < len(nameList) {
			newName = fmt.Sprintf("%s.pdf", strings.TrimSpace(nameList[i]))
		}
		if err := os.Rename(f, filepath.Join(outputDir, newName)); err != nil {
			c.String(http.StatusInternalServerError, "Failed to rename file: %s", err.Error())
			return
		}
	}

	files, _ = filepath.Glob(filepath.Join(outputDir, "*.pdf"))

	zipFilePath := zipName
	zipFile, err := os.Create(zipFilePath)
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to create ZIP file: %s", err.Error())
		return
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	if err := addFilesToZip(zipWriter, files); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	if err := zipWriter.Close(); err != nil {
		c.String(http.StatusInternalServerError, "Failed to close ZIP: %s", err.Error())
		return
	}

	// Clean up output directory
	if err := os.RemoveAll(outputDir); err != nil {
		c.String(http.StatusInternalServerError, "Failed to remove output directory: %s", err.Error())
		return
	}

	defer os.Remove(zipFilePath)
	c.File(zipFilePath)
}

// addFilesToZip adds files to the ZIP archive
func addFilesToZip(zipWriter *zip.Writer, files []string) error {
	for _, f := range files {
		fileToZip, err := os.Open(f)
		if err != nil {
			return fmt.Errorf("failed to open file: %s", err)
		}

		zipEntry, err := zipWriter.Create(filepath.Base(f))
		if err != nil {
			fileToZip.Close()
			return fmt.Errorf("failed to add file to ZIP: %s", err)
		}
		_, err = io.Copy(zipEntry, fileToZip)
		fileToZip.Close()
		if err != nil {
			return fmt.Errorf("failed to copy file to ZIP: %s", err)
		}
	}
	return nil
}

// sortFilesByPageNumber sorts files by page number extracted from their names
func sortFilesByPageNumber(files *[]string) {
	re := regexp.MustCompile(`(\d+)\.pdf$`)

	sort.Slice(*files, func(i, j int) bool {
		numI := extractPageNumber((*files)[i], re)
		numJ := extractPageNumber((*files)[j], re)
		return numI < numJ
	})
}

// extractPageNumber extracts the page number from a filename using a regex
func extractPageNumber(filename string, re *regexp.Regexp) int {
	matches := re.FindStringSubmatch(filename)
	if len(matches) < 2 {
		return 0
	}
	num, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0
	}
	return num
}
