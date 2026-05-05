package main

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestTextAPI(t *testing.T) {
	// Setup
	sharedText = ""

	t.Run("GET /api/text returns empty text initially", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/api/text", nil)
		if err != nil {
			t.Fatal(err)
		}
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(getTextHandler)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("wrong status code: got %v want %v", status, http.StatusOK)
		}
		if rr.Body.String() != "" {
			t.Errorf("unexpected body: got %v want empty string", rr.Body.String())
		}
	})

	t.Run("POST /api/text saves text", func(t *testing.T) {
		textToSave := "Hello, TDD!"
		req, err := http.NewRequest("POST", "/api/text", bytes.NewBufferString(textToSave))
		if err != nil {
			t.Fatal(err)
		}
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(postTextHandler)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("wrong status code: got %v want %v", status, http.StatusOK)
		}
		if sharedText != textToSave {
			t.Errorf("sharedText not updated: got %v want %v", sharedText, textToSave)
		}
	})
}

func TestFilesAPI(t *testing.T) {
	// Setup temporary upload directory
	tempDir, err := os.MkdirTemp("", "fileshare_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)
	uploadsDir = tempDir

	t.Run("GET /api/files initially empty", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/api/files", nil)
		if err != nil {
			t.Fatal(err)
		}
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(listFilesHandler)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("wrong status code: got %v want %v", status, http.StatusOK)
		}

		var files []FileInfo
		if err := json.Unmarshal(rr.Body.Bytes(), &files); err != nil {
			t.Fatal(err)
		}
		if len(files) != 0 {
			t.Errorf("expected 0 files, got %v", len(files))
		}
	})

	t.Run("POST /api/upload uploads file", func(t *testing.T) {
		var b bytes.Buffer
		w := multipart.NewWriter(&b)
		fw, err := w.CreateFormFile("file", "test.txt")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.WriteString(fw, "test file content"); err != nil {
			t.Fatal(err)
		}
		w.Close()

		req, err := http.NewRequest("POST", "/api/upload", &b)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", w.FormDataContentType())

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(uploadFileHandler)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("wrong status code: got %v want %v", status, http.StatusOK)
		}

		// Check if file exists
		content, err := os.ReadFile(filepath.Join(uploadsDir, "test.txt"))
		if err != nil {
			t.Fatal("file was not saved")
		}
		if string(content) != "test file content" {
			t.Errorf("wrong file content: got %v want %v", string(content), "test file content")
		}
	})

	t.Run("GET /api/download/test.txt downloads file", func(t *testing.T) {
		mux := setupRouter()
		req, err := http.NewRequest("GET", "/api/download/test.txt", nil)
		if err != nil {
			t.Fatal(err)
		}
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("wrong status code: got %v want %v", status, http.StatusOK)
		}
		if rr.Body.String() != "test file content" {
			t.Errorf("wrong content: got %v", rr.Body.String())
		}
	})

	t.Run("DELETE /api/files/test.txt deletes file", func(t *testing.T) {
		mux := setupRouter()
		req, err := http.NewRequest("DELETE", "/api/files/test.txt", nil)
		if err != nil {
			t.Fatal(err)
		}
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("wrong status code: got %v want %v", status, http.StatusOK)
		}

		if _, err := os.Stat(filepath.Join(uploadsDir, "test.txt")); !os.IsNotExist(err) {
			t.Errorf("file was not deleted")
		}
	})
}
