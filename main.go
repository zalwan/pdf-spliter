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

	os.MkdirAll(uploadDir, os.ModePerm)

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	r.POST("/split", handleSplitPDF)

	r.Run(":8080")
}

func handleSplitPDF(c *gin.Context) {
	file, err := c.FormFile("pdf")
	if err != nil {
		c.String(http.StatusBadRequest, "Gagal mengunggah file: %s", err.Error())
		return
	}

	names := c.PostForm("names")
	nameList := strings.Split(strings.TrimSpace(names), "\n")

	src := filepath.Join(uploadDir, file.Filename)
	if err := c.SaveUploadedFile(file, src); err != nil {
		c.String(http.StatusInternalServerError, "Gagal menyimpan file: %s", err.Error())
		return
	}

	os.MkdirAll(outputDir, os.ModePerm)

	if err := api.SplitFile(src, outputDir, 1, nil); err != nil {
		c.String(http.StatusInternalServerError, "Gagal memisahkan PDF: %s", err.Error())
		return
	}

	files, err := filepath.Glob(filepath.Join(outputDir, "*.pdf"))
	if err != nil || len(files) == 0 {
		c.String(http.StatusInternalServerError, "Tidak ada file yang ditemukan setelah split.")
		return
	}

	sortFilesByPageNumber(&files)

	for i, f := range files {
		newName := fmt.Sprintf("part%d.pdf", i+1)
		if i < len(nameList) {
			newName = fmt.Sprintf("%s.pdf", strings.TrimSpace(nameList[i]))
		}
		os.Rename(f, filepath.Join(outputDir, newName))
	}

	files, _ = filepath.Glob(filepath.Join(outputDir, "*.pdf"))

	zipFilePath := zipName
	zipFile, err := os.Create(zipFilePath)
	if err != nil {
		c.String(http.StatusInternalServerError, "Gagal membuat file ZIP: %s", err.Error())
		return
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)

	for _, f := range files {
		fileToZip, err := os.Open(f)
		if err != nil {
			c.String(http.StatusInternalServerError, "Gagal membuka file: %s", err.Error())
			return
		}

		zipEntry, err := zipWriter.Create(filepath.Base(f))
		if err != nil {
			fileToZip.Close()
			c.String(http.StatusInternalServerError, "Gagal menambah file ke ZIP: %s", err.Error())
			return
		}
		_, err = io.Copy(zipEntry, fileToZip)
		fileToZip.Close()
		if err != nil {
			c.String(http.StatusInternalServerError, "Gagal menyalin file ke ZIP: %s", err.Error())
			return
		}
	}

	if err := zipWriter.Close(); err != nil {
		c.String(http.StatusInternalServerError, "Gagal menutup ZIP: %s", err.Error())
		return
	}

	os.RemoveAll(outputDir)

	defer os.Remove(zipFilePath)
	c.File(zipFilePath)
}

func sortFilesByPageNumber(files *[]string) {
	re := regexp.MustCompile(`(\d+)\.pdf$`)

	sort.Slice(*files, func(i, j int) bool {
		numI := extractPageNumber((*files)[i], re)
		numJ := extractPageNumber((*files)[j], re)
		return numI < numJ
	})
}

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
