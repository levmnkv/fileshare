package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	sharedText string
	textMutex  sync.RWMutex
	uploadsDir = getEnv("UPLOAD_DIR", "/mnt/docker-data")
)

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

type FileInfo struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
	Time string `json:"time"`
}

func getTextHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	textMutex.RLock()
	defer textMutex.RUnlock()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(sharedText))
}

func postTextHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()
	textMutex.Lock()
	sharedText = string(bodyBytes)
	textMutex.Unlock()
	w.WriteHeader(http.StatusOK)
}

func listFilesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	entries, err := os.ReadDir(uploadsDir)
	if err != nil {
		http.Error(w, "Failed to read directory", http.StatusInternalServerError)
		return
	}

	var files []FileInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		files = append(files, FileInfo{
			Name: entry.Name(),
			Size: info.Size(),
			Time: info.ModTime().Format(time.RFC3339),
		})
	}

	if files == nil {
		files = make([]FileInfo, 0)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

func uploadFileHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.ParseMultipartForm(10 << 20) // 10 MB limit in memory
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	destPath := filepath.Join(uploadsDir, filepath.Base(handler.Filename))
	destFile, err := os.Create(destPath)
	if err != nil {
		http.Error(w, "Error creating file", http.StatusInternalServerError)
		return
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, file); err != nil {
		http.Error(w, "Error saving file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "File uploaded successfully")
}

func setupRouter() *http.ServeMux {
	mux := http.NewServeMux()
	
	mux.HandleFunc("/api/text", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			getTextHandler(w, r)
		} else if r.Method == http.MethodPost {
			postTextHandler(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	
	mux.HandleFunc("/api/files", listFilesHandler)
	mux.HandleFunc("/api/upload", uploadFileHandler)
	
	mux.HandleFunc("/api/download/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		filename := strings.TrimPrefix(r.URL.Path, "/api/download/")
		if filename == "" {
			http.Error(w, "Filename required", http.StatusBadRequest)
			return
		}
		filePath := filepath.Join(uploadsDir, filepath.Base(filename))
		http.ServeFile(w, r, filePath)
	})

	mux.HandleFunc("/api/files/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		filename := strings.TrimPrefix(r.URL.Path, "/api/files/")
		if filename == "" {
			http.Error(w, "Filename required", http.StatusBadRequest)
			return
		}
		filePath := filepath.Join(uploadsDir, filepath.Base(filename))
		if err := os.Remove(filePath); err != nil {
			http.Error(w, "Failed to delete file", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	mux.Handle("/", http.FileServer(http.Dir("./static")))

	return mux
}

func main() {
	if err := os.MkdirAll(uploadsDir, os.ModePerm); err != nil {
		log.Printf("Warning: Failed to create uploads directory: %v", err)
	}

	mux := setupRouter()
	port := "8080"
	log.Printf("Starting fileshare MVP on :%s...", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
