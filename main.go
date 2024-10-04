package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

const uploadPath = "./uploads"

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, "Upload File")

	//upload file form
	fmt.Fprintf(w, `
	<form enctype="multipart/form-data" action="/upload" method="post">
		<input type="file" name="uploadfile" />
		<input type="submit" value="Upload File" />
	</form>

	<form action="/files" method="get">
		<button type="submit">Show all the files</button>
	</form>
	`)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	file, fileHeader, err := r.FormFile("uploadfile")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	filePath := filepath.Join(uploadPath, fileHeader.Filename)
	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}

	//fmt.Fprintf(w, "File uploaded successfully: %s ..... Redirecting to home page", fileHeader.Filename)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func showFiles(w http.ResponseWriter, r *http.Request) {
	files, err := os.ReadDir(uploadPath)
	if err != nil {
		http.Error(w, "Error reading files", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<h1>Files</h1> 
	
	<form action="/" method="get">
		<button type="home">Home</button>
	</form>
	`)
	for _, file := range files {
		fmt.Fprintf(w, `<p><a href="/download?file=%s">%s</p><a>`, file.Name(), file.Name())
	}
}

func downloadFile(w http.ResponseWriter, r *http.Request) {
	fileName := r.URL.Query().Get("file")

	filePath := filepath.Join(uploadPath, fileName)
	if fileName == "" {
		http.Error(w, "File not specified", http.StatusBadRequest)
		return
	}
	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	w.Header().Set("Content-Type", "application/octet-stream")

	// Set cache control headers to prevent caching
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	http.ServeFile(w, r, filePath)
}

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/upload", uploadHandler)
	http.HandleFunc("/files", showFiles)
	http.HandleFunc("/download", downloadFile)

	fmt.Printf("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
