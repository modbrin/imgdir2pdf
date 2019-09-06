package main

import (
	"encoding/binary"
	"fmt"
	"github.com/jung-kurt/gofpdf"
	"image"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	helpString = "\nusage: imgdir2pdf DIR\n" +
		"Convert all images in given directory to single pdf.\n" +
		"Order is defined by sorting their names.\n" +
		"\nSupported files: png, jpg, jpeg, gif (first frame only)\n" +
		"Resulting PDF matches DIR's base name and is saved in DIR.\n"
	a4Width  = 210
	a4Height = 297
)

var imageFormats = [...]string{"png", "jpg", "jpeg", "gif"}

// Print program help message
func printHelp() {
	fmt.Println(helpString)
}

type strCheck func(string, string) bool

// Check if any of 'other' conform to property 'f' with 'target'
func any(target string, other []string, f strCheck) bool {
	for _, e := range other {
		if f(target, e) {
			return true
		}
	}
	return false
}

// Get list of all files with extensions from 'fileExtension' in dirpath.
// Resulting paths are absolute
func lsdir(dirpath string, fileExtension []string) []string {
	var result []string
	files, err := ioutil.ReadDir(dirpath)
	if err != nil {
		panic(err)
	}
	for _, elem := range files {
		curfile := elem.Name()
		if !elem.IsDir() && any(curfile, fileExtension, strings.HasSuffix) {
			result = append(result, elem.Name())
		}
	}
	sort.Slice(
		result,
		func(i, j int) bool {
			return sortName(result[i]) < sortName(result[j])
		},
	)
	for i, elem := range result {
		result[i], err = filepath.Abs(filepath.Join(dirpath, elem))
		if err != nil {
			panic(err)
		}
	}
	return result
}

// adapted from https://stackoverflow.com/questions/51359930/sorting-strings-with-numbers-in-filenames-with-golang
// sortName returns a filename sort key with
// non-negative integer suffixes in numeric order.
// For example, amt, amt0, amt2, amt10, amt099, amt100, ...
func sortName(filename string) string {
	ext := filepath.Ext(filename)
	name := filename[:len(filename)-len(ext)]
	// split numeric suffix
	i := len(name) - 1
	for ; i >= 0; i-- {
		if '0' > name[i] || name[i] > '9' {
			break
		}
	}
	i++
	// string numeric suffix to uint64 bytes
	// empty string is zero, so integers are plus one
	b64 := make([]byte, 64/8)
	s64 := name[i:]
	if len(s64) > 0 {
		u64, err := strconv.ParseUint(s64, 10, 64)
		if err == nil {
			binary.BigEndian.PutUint64(b64, u64+1)
		}
	}
	// prefix + numeric-suffix + ext
	return name[:i] + string(b64) + ext
}

// Image Processing

// Get dimenstions of given image
func getImageSize(imagepath string) (w, h float64) {
	file, err := os.Open(imagepath)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	imgconf, _, err := image.DecodeConfig(file)
	if err != nil {
		panic(err)
	}
	return float64(imgconf.Width), float64(imgconf.Height)
}

// Add image to pdf
func addImagePage(document *gofpdf.Fpdf, imagepath string) {
	imageW, imageH := getImageSize(imagepath)
	resW, resH := optimalPageSize(a4Width, a4Height, imageW, imageH)
	ext := strings.ToUpper(path.Ext(imagepath)[1:])
	document.AddPageFormat("P", gofpdf.SizeType{Wd: resW, Ht: resH})
	document.ImageOptions(imagepath, 0, 0, resW, resH, false, gofpdf.ImageOptions{ImageType: ext, ReadDpi: true}, 0, "")
}

// Initialize new pdf file with custom size in mm
func createDocument(w, h float64) *gofpdf.Fpdf {
	return gofpdf.NewCustom(&gofpdf.InitType{UnitStr: "mm", Size: gofpdf.SizeType{Wd: w, Ht: h}})
}

// very simplistic size determination algorithm
// makes document size based to be
// similar to template height, e.g. A4 size
func optimalPageSize(templateW, templateH, givenW, givenH float64) (w, h float64) {
	scaler := templateW / givenW
	w, h = templateW, scaler*givenH
	return w, h
}

// Add images from paths into single pdf
func processImages(paths []string, saveAs string) {
	if len(paths) < 1 {
		panic("No suitable files in given directory.")
	}
	firstW, firstH := getImageSize(paths[0])
	pdf := createDocument(optimalPageSize(a4Width, a4Height, firstW, firstH))
	for _, elem := range paths {
		addImagePage(pdf, elem)
	}
	err := pdf.OutputFileAndClose(saveAs)
	if err != nil {
		fmt.Printf("Error writing pdf: %v", err)
	}
}

// Construct absolute path of resulting pdf as
// base folder of 'basepath'
// i.e. /some/folder/ will turn into /abs/path/some/folder/folder.pdf
func getOutFilename(basepath string) string {
	resultPath, err := filepath.Abs(basepath)
	if err != nil {
		panic(err)
	}
	basename := filepath.Base(resultPath)
	return filepath.Join(resultPath, fmt.Sprintf("%s.pdf", basename))
}

// Main logic of program
func main() {
	if len(os.Args) <= 1 {
		printHelp()
		return
	}
	dir := os.Args[1]
	processImages(lsdir(dir, imageFormats[:]), getOutFilename(dir))
}
