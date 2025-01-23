package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"sort"
	"sync"
	"time"
)

const (
	downloadURL     = "http://localhost:8080/download"
	uploadURL       = "http://localhost:8080/upload"
	dataSizeMB      = 10 // データサイズ（MB）
	numMeasurements = 5  // 測定回数
	threads         = 4  // 並列ダウンロードのスレッド数
)

// 並列ダウンロード速度測定
func parallelDownload(url string, threads int) float64 {
	var wg sync.WaitGroup
	start := time.Now()

	// チャネルを利用して一斉に開始
	startSignal := make(chan struct{})
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-startSignal // goroutineはチャネルから値を受信するまでここで待機する
			resp, err := http.Get(url)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
			if resp.StatusCode != http.StatusOK {
				fmt.Printf("Error: HTTP Status %d\n", resp.StatusCode)
				return
			}
			defer resp.Body.Close()

			io.Copy(io.Discard, resp.Body) // データを捨てる
		}()
	}

	// チャネルが閉じられることで、待機してるgoroutineが一斉開始する
	close(startSignal)
	wg.Wait()
	duration := time.Since(start).Seconds()
	totalData := float64(dataSizeMB*1024*1024*8) * float64(threads) // データ量（ビット）
	return totalData / (duration * 1024 * 1024)                     // Mbpsで返す
}

// 並列アップロード速度測定
func parallelUpload(url string, threads int) float64 {
	var wg sync.WaitGroup
	start := time.Now()

	data := bytes.Repeat([]byte("A"), dataSizeMB*1024*1024) // アップロード用のデータ

	// チャネルを利用して一斉に開始
	startSignal := make(chan struct{})
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-startSignal // goroutineはチャネルから値を受信するまでここで待機する
			resp, err := http.Post(url, "application/octet-stream", bytes.NewReader(data))
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
			if resp.StatusCode != http.StatusOK {
				fmt.Printf("Error: HTTP Status %d\n", resp.StatusCode)
				return
			}
			defer resp.Body.Close()
		}()
	}

	// チャネルが閉じられることで、待機してるgoroutineが一斉開始する
	close(startSignal)
	wg.Wait()
	duration := time.Since(start).Seconds()
	totalData := float64(dataSizeMB*1024*1024*8) * float64(threads) // データ量（ビット）
	return totalData / (duration * 1024 * 1024)                     // Mbpsで返す
}

// 測定結果を分析
func analyzeSpeeds(speeds []float64) (float64, float64) {
	average := calculateAverage(speeds)
	median := calculateMedian(speeds)
	return average, median
}

func calculateAverage(speeds []float64) float64 {
	if len(speeds) == 0 {
		return 0
	}

	var total float64
	for _, speed := range speeds {
		total += speed
	}
	return total / float64(len(speeds))
}

func calculateMedian(speeds []float64) float64 {
	if len(speeds) == 0 {
		return 0
	}

	sort.Float64s(speeds)
	mid := len(speeds) / 2
	if len(speeds)%2 == 0 {
		return (speeds[mid-1] + speeds[mid]) / 2
	}
	return speeds[mid]
}

func displayResults(testType string, speeds []float64, avg, median float64) {
	fmt.Printf("\n===== %s Speed Test Results =====\n", testType)
	for i, speed := range speeds {
		fmt.Printf("Measurement %d: %.2f Mbps\n", i+1, speed)
	}
	fmt.Printf("\nAverage Speed: %.2f Mbps\n", avg)
	fmt.Printf("Median Speed: %.2f Mbps\n", median)
	fmt.Println("===================================")
}

func main() {
	// ダウンロード速度測定
	fmt.Println("Measuring download speed...")
	downloadSpeeds := make([]float64, numMeasurements)
	for i := 0; i < numMeasurements; i++ {
		downloadSpeeds[i] = parallelDownload(downloadURL, threads)
		fmt.Printf("Measurement %d: %.2f Mbps\n", i+1, downloadSpeeds[i])
	}
	downloadAvg, downloadMedian := analyzeSpeeds(downloadSpeeds)
	displayResults("Download", downloadSpeeds, downloadAvg, downloadMedian)

	// アップロード速度測定
	fmt.Println("\nMeasuring upload speed...")
	uploadSpeeds := make([]float64, numMeasurements)
	for i := 0; i < numMeasurements; i++ {
		uploadSpeeds[i] = parallelUpload(uploadURL, threads)
		fmt.Printf("Measurement %d: %.2f Mbps\n", i+1, uploadSpeeds[i])
	}
	uploadAvg, uploadMedian := analyzeSpeeds(uploadSpeeds)
	displayResults("Upload", uploadSpeeds, uploadAvg, uploadMedian)
}
