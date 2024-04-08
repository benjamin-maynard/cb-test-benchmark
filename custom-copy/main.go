package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

func main() {

	var bucketName, folderPath, localPath string
	flag.StringVar(&bucketName, "bucket-name", "", "The name of the GCS bucket")
	flag.StringVar(&folderPath, "bucket-prefix", "", "The path in the GCS bucket")
	flag.StringVar(&localPath, "local-path", "./", "The local path to write the files to")
	flag.Parse()

	if bucketName == "" || localPath == "" {
		log.Fatal("bucket-name and local-path must be set")
	}

	client, err := storage.NewClient(context.Background())
	if err != nil {
		log.Fatalf("error instantiating client: %v", err)
	}

	err = recursiveDownload(context.Background(), client, bucketName, folderPath, "test", 5)
	if err != nil {
		log.Fatalf("error downloading files: %v", err)
	}

	defer client.Close()

}

func recursiveDownload(ctx context.Context, client *storage.Client, bucketName, folderPath, localPath string, concurrencyLimit int) error {

	query := &storage.Query{Prefix: folderPath}
	it := client.Bucket(bucketName).Objects(ctx, query)

	sem := make(chan struct{}, concurrencyLimit)
	var wg sync.WaitGroup

	for {

		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("error iterating over objects: %v", err)
		}

		// Skip folders
		if strings.HasSuffix(attrs.Name, "/") {
			continue
		}

		// Increment WaitGroup counter
		wg.Add(1)

		// Acquire semaphore
		sem <- struct{}{}

		// Start a goroutine for each download
		go func(attrs *storage.ObjectAttrs) {

			defer wg.Done()
			defer func() { <-sem }()

			relativePath := strings.TrimPrefix(attrs.Name, folderPath)
			localFilePath := filepath.Join(localPath, relativePath)

			if err := os.MkdirAll(filepath.Dir(localFilePath), 0755); err != nil {
				log.Printf("creating local directory: %v", err)
				return
			}

			file, err := os.Create(localFilePath)
			if err != nil {
				log.Printf("error creating local file: %v", err)
				return
			}
			defer file.Close()

			rc, err := client.Bucket(bucketName).Object(attrs.Name).NewReader(ctx)
			if err != nil {
				log.Printf("error creating object reader: %v", err)
				return
			}
			defer rc.Close()

			if _, err := io.Copy(file, rc); err != nil {
				log.Printf("error copying data: %v", err)
				return
			}

			log.Printf("Downloaded file: %s/%s", attrs.Bucket, attrs.Name)

		}(attrs)

	}

	wg.Wait()
	return nil

}
