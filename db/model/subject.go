package model

import (
	"errors"
	"fmt"
	"net/http"
	"sort"

	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/franzwilhelm/uio-exam-helper/db"
	log "github.com/sirupsen/logrus"
)

type Subject struct {
	ID        string     `json:"id"`
	Study     string     `json:"study"`
	Faculty   string     `json:"faculty"`
	Resources []Resource `json:"resources"`
}

func (s Subject) Logger() *log.Entry {
	return log.WithField("subject", s.ID)
}

func (s Subject) BaseURL() string {
	return fmt.Sprintf("https://www.uio.no/studier/emner/%s/%s/%s", s.Faculty, s.Study, s.ID)
}

func (s Subject) ResourceURL() string {
	return fmt.Sprint(s.BaseURL(), "/oppgaver")
}

func (s Subject) IsValid() bool {
	url := s.BaseURL()
	resp, err := http.Get(url)
	s.Logger().Infof("Requested %s. Got %v, %v", url, resp.StatusCode, err)
	return err == nil && resp.StatusCode == 200
}

func (s Subject) DiscoverNewResources() (resources []Resource, err error) {
	l := s.Logger()
	doc, err := goquery.NewDocument(s.ResourceURL())
	if err != nil {
		return nil, err
	}

	// find all resource links
	doc.Find("#vrtx-main-user").Each(func(_ int, parent *goquery.Selection) {
		parent.Find("a").Each(func(_ int, item *goquery.Selection) {
			resource := Resource{SubjectID: s.ID}
			url, _ := item.Attr("href")
			if !strings.Contains(url, "pdf") {
				return
			}
			if r, err := ResourceByURL(url); err == nil {
				l.Warnf("Resource %s exists. Skipping db creation...", r.Name)
				return
			}

			if strings.HasPrefix(url, "/") {
				url = "https://uio.no" + url
			}
			resource.UseURL(url)
			resources = append(resources, resource)
		})
	})
	return resources, nil
}

func (s *Subject) FetchLatestResources() error {
	tx := db.Default.Begin()
	resources, err := s.DiscoverNewResources()
	if err != nil {
		return err
	}
	for i := range resources {
		err := db.Default.FirstOrCreate(&resources[i]).Error
		if err != nil {
			tx.Rollback()
			s.Resources = nil
			return err
		}
	}
	tx.Commit()
	return nil
}

func (s *Subject) DownloadResources() {
	l := s.Logger()
	l.Infof("Downloading resrouces")
	for _, r := range s.Resources {
		if !r.IsDownloaded() {
			r.Download()
		} else {
			l.Warnf("Skipping download. Resource %s already exists!", r.Name)
		}
	}
}

func (s *Subject) DeleteResources() {
	s.Logger().Warnf("Deleting resrouces")
	for _, r := range s.Resources {
		if r.IsDownloaded() {
			r.Delete()
		}
	}
}

func (s *Subject) GenerateWordTree() ([]Group, error) {
	var allWords []string
	for _, r := range s.Resources {
		words, err := r.GetWords()
		if err != nil {
			return nil, err
		}
		allWords = append(allWords, words...)
	}
	groups := groupWords(allWords)

	sort.Slice(groups[:], func(i, j int) bool {
		return groups[i].Count > groups[j].Count
	})

	return groups, nil
}
func (s *Subject) Refresh() error {
	return db.Default.Preload("Resources").Where("id = ?", s.ID).First(s).Error
}

func GetSubjectByID(id string) (s Subject, err error) {
	return s, db.Default.Where("id = ?", id).First(&s).Error
}

func NewSubject(id, faculty, study string) (Subject, error) {
	s := Subject{
		ID:      strings.ToUpper(id),
		Faculty: faculty,
		Study:   study,
	}
	if !s.IsValid() {
		return s, errors.New(fmt.Sprint("Not a valid UiO subject:", s.BaseURL()))
	}
	return s, db.Default.Create(&s).Error
}
