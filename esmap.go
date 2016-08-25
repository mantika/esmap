package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	if waitUntilESReady() {
		sendMappings()
		sendTemplates()
	} else {
		log.Println("ES is not ready. Won't send mappings")
	}
}

func sendTemplates() {
	log.Println("Will send all template files under /templates")
	filepath.Walk("/templates", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println(err)
			return err
		}
		if path == "/templates" && info.IsDir() {
			return nil
		}
		relPath := strings.TrimPrefix(path, "/templates/")
		if !info.IsDir() {
			file := relPath
			fileName := filepath.Base(relPath)
			chunks := strings.Split(fileName, ".")
			name := chunks[0]
			log.Printf("Sending [%s] as template name [%s]\n", file, name)
			fileContent, readFileErr := ioutil.ReadFile(path)
			if readFileErr != nil {
				log.Printf("Couldn't read file [%s]. %s\n", path, readFileErr.Error())
				return nil
			}
			req, reqErr := http.NewRequest("PUT", fmt.Sprintf("http://%s/_template/%s", os.Getenv("ESHOST"), name), bytes.NewReader(fileContent))
			if reqErr != nil {
				log.Printf(reqErr.Error())
				return nil
			}
			_, putErr := http.DefaultClient.Do(req)
			if putErr != nil {
				log.Printf("Couldn't post file [%s] as template [%s]. %s\n", file, name, putErr.Error())
				return nil
			}
		}
		return nil
	})
}

func sendMappings() {
	log.Println("Will send all mapping files under /mappings")
	filepath.Walk("/mappings", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println(err)
			return err
		}
		if path == "/mappings" && info.IsDir() {
			return nil
		}
		relPath := strings.TrimPrefix(path, "/mappings/")
		if info.IsDir() {
			log.Printf("Found index [%s]\n", relPath)
		}
		if !info.IsDir() {
			chunks := strings.Split(relPath, "/")
			if len(chunks) != 2 {
				log.Printf("Ignoring [%s] as it doesn't follow the <index>/<mapping file> standard\n", relPath)
				return nil
			}
			index := chunks[0]
			file := chunks[1]
			log.Printf("Sending [%s] to index [%s]\n", file, index)
			fileContent, readFileErr := ioutil.ReadFile(path)
			if readFileErr != nil {
				log.Printf("Couldn't read file [%s]. %s\n", path, readFileErr.Error())
				return nil
			}
			_, postErr := http.Post(fmt.Sprintf("http://%s/%s", os.Getenv("ESHOST"), index), "application/json", bytes.NewReader(fileContent))
			if postErr != nil {
				log.Printf("Couldn't post file [%s] to index [%s]. %s\n", file, index, postErr.Error())
				return nil
			}
		}
		return nil
	})
}

func waitUntilESReady() bool {
	ready := false
	log.Printf("Waiting until ES (%s) is available.\n", os.Getenv("ESHOST"))
	for i := 0; i < 60; i++ {
		resp, err := http.Get(fmt.Sprintf("http://%s/_cluster/health", os.Getenv("ESHOST")))
		if err != nil {
			log.Println("Failed to check status. ", err)
			retry()
			continue
		}
		if resp.StatusCode != 200 {
			log.Printf("Failed to check status. Got %d status code.\n", resp.StatusCode)
			retry()
			continue
		}
		b, readErr := ioutil.ReadAll(resp.Body)
		if readErr != nil {
			log.Println("Failed to cluster health response. ", readErr)
			retry()
			continue
		}

		var health map[string]interface{}
		jsonErr := json.Unmarshal(b, &health)
		if jsonErr != nil {
			log.Println("Failed to read cluster health json response. ", jsonErr)
			retry()
			continue
		}

		status := health["status"].(string)
		if status == "red" {
			log.Printf("Cluster status is [%s]\n", status)
			retry()
			continue
		}
		ready = true
		break
	}

	return ready
}

func retry() {
	log.Println("Retrying...")
	time.Sleep(1 * time.Second)
}
