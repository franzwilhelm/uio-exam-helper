package model

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"code.sajari.com/docconv"
	log "github.com/sirupsen/logrus"
)

type Resource struct {
	Name      string `json:"name"`
	URL       string `json:"URL"`
	Filepath  string `json:"filepath"`
	SubjectID string `json:"subjectID"`
}

func (r *Resource) UseURL(url string) {
	r.URL = url
	tokens := strings.Split(url, "/")
	name := tokens[len(tokens)-1]
	if filepath.Ext(name) == "" {
		name += ".pdf"
	}
	r.Name = name
}

func (r *Resource) IsDownloaded() bool {
	_, err := os.Stat(r.Name)
	return err == nil
}

func (r *Resource) Download() error {
	if r.IsDownloaded() {
		log.Warn("Skipping download! Resource already exists: ", r.Name)
		return nil
	}

	file, err := os.Create(r.Name)
	defer file.Close()
	if err != nil {
		log.WithError(err).Error("Error while creating " + r.Name)
		return err
	}

	log.Infof("Downloading %s to %s", r.URL, r.Name)

	response, err := http.Get(r.URL)
	defer response.Body.Close()
	if err != nil {
		log.WithError(err).Error("Error while downloading" + r.URL)
		return err
	}

	_, err = io.Copy(file, response.Body)
	if err != nil {
		log.WithError(err).Error("Error while copying to" + file.Name())
		return err
	}
	return nil
}

func (r *Resource) Delete() error {
	log.Warn("Deleting file: ", r.Name)
	return os.Remove(r.Name)
}

func (r *Resource) GenerateWordTree() {
	res, err := docconv.ConvertPath(r.Name)
	if err != nil {
		log.Fatal(err)
	}
	lower := strings.ToLower(res.Body)
	lines := strings.Split(lower, "\n")
	words := words(lines)
	processed, err := deleteSpecials(words)
	if err != nil {
		log.Fatal(err)
	}
	countMap := generateCountMap(processed)
	max := getMaxCount(countMap)
	strs := make([][]string, max)
	for key, value := range countMap {
		strs[value-1] = append(strs[value-1], key)
	}
	for i, arr := range strs {
		if len(arr) > 0 {
			fmt.Printf("%v: %s\n", i+1, arr)
		}
	}
}

func words(lines []string) (words []string) {
	for _, line := range lines {
		words = append(words, strings.Split(line, " ")...)
	}
	return
}

func deleteSpecials(words []string) (processed []string, err error) {
	for _, line := range words {
		reg, err := regexp.Compile("[^a-zA-ZæøåÆØÅ]+")
		if err != nil {
			return nil, err
		}
		replaced := reg.ReplaceAllString(line, "")
		if len(replaced) > 2 {
			processed = append(processed, replaced)
		}
	}
	return
}

func generateCountMap(a []string) map[string]int {
	countMap := make(map[string]int)
	for _, s := range a {
		countMap[s] += 1
	}
	return countMap
}

func getMaxCount(countMap map[string]int) int {
	max := 0
	for _, value := range countMap {
		if value > max {
			max = value
		}
	}
	return max
}
