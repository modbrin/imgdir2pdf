# imgdir2pdf

Small utility for converting series of images into single pdf.

## Download

Get it in releases. Windows and Linux versions are provided.

## How to use
```shell script
imgdir2pdf path/to/images/dir
```

All images of supported formats (png, jpg, gif) will be merged into pdf.

Resulting pdf is saved in same folder with images and matches folder's base name.


## How to build
```shell script
go build imgdir2pdf
```

## Dependencies
> github.com/jung-kurt/gofpdf

## Future considerations
* Add argument for output dir
* Add cropping utility with convenient interface
* Add more options for modifying images, e.g. rotating, size fitting
* Add progress bar
* Add OCR features