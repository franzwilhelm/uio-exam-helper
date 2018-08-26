package model

import (
	"fmt"

	"strings"

	"github.com/PuerkitoBio/goquery"
	log "github.com/sirupsen/logrus"
)

type Subject struct {
	ID          string     `json:"id"`
	Study       string     `json:"study"`
	Faculty     string     `json:"faculty"`
	ResourceURL string     `json:"resourceURL"`
	Resources   []Resource `json:"resources"`
}

func (s *Subject) PreloadResources() {
	url := fmt.Sprintf("https://www.uio.no/studier/emner/%s/%s/%s/oppgaver/", s.Faculty, s.Study, s.ID)
	doc, err := goquery.NewDocument(url)
	if err != nil {
		log.Fatal(err)
	}
	doc.Find(".vrtx-title-link").Each(func(index int, item *goquery.Selection) {
		resource := Resource{SubjectID: s.ID}
		url, _ := item.Attr("href")
		resource.UseURL(url)
		s.Resources = append(s.Resources, resource)
	})
	s.ResourceURL = url
}

func NewSubject(id, faculty, study string) Subject {
	return Subject{
		ID:      strings.ToUpper(id),
		Faculty: faculty,
		Study:   study,
	}
}
