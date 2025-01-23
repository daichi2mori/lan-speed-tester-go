package main

import (
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

const dataSizeMB = 10                   // ダウンロード用のデータサイズ（MB）
const maxUploadSize = 100 * 1024 * 1024 // アップロードの最大サイズ(100MB)

func main() {
	// ダウンロード用エンドポイント
	http.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		data := make([]byte, dataSizeMB*1024*1024) // ダミーデータ作成

		// ランダムデータを作成
		_, err := rand.Read(data)
		if err != nil {
			fmt.Println("Error generating random data:", err)
			return
		}

		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)

		// データ送信
		_, err = w.Write(data)
		if err != nil {
			fmt.Println("Error sending data:", err)
			return
		}
		fmt.Printf("Sent %d MB of data to client\n", dataSizeMB)
	})

	// アップロード用エンドポイント
	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		// Content-Lengthヘッダーを取得
		contentLengthStr := r.Header.Get("Content-Length")
		if contentLengthStr == "" {
			http.Error(w, "Content-Length header is missing", http.StatusBadRequest)
			return
		}
		contentLength, err := strconv.ParseInt(contentLengthStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid Content-Length header", http.StatusBadRequest)
			return
		}

		if contentLength > maxUploadSize {
			http.Error(w, "File size exceeds the maximum limit", http.StatusRequestEntityTooLarge)
			return
		}

		limitedReader := io.LimitReader(r.Body, contentLength)

		data, err := io.ReadAll(limitedReader)
		if err != nil {
			http.Error(w, "Failed to read data", http.StatusInternalServerError)
			return
		}

		fmt.Printf("Received %d bytes from client\n", len(data))
		w.WriteHeader(http.StatusOK)
	})

	// サーバーの起動
	fmt.Println("Starting server on port :8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Server failed:", err)
	}
}
