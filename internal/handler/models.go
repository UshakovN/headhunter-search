package handler

import (
  "fmt"
  "strings"
)

type ParsedMessage struct {
  AreaCode       string
  ExperienceCode string
  Keywords       string
}

func parseTextMessage(text string) (*ParsedMessage, error) {
  const (
    sep   = "//"
    count = 3
  )
  if text == "" {
    return nil, fmt.Errorf("text message is empty string")
  }
  parts := strings.Split(text, sep)

  if c := len(parts); c != count {
    return nil, fmt.Errorf("text message has a wrong parts count")
  }
  areaCode := strings.TrimSpace(parts[0])

  if _, ok := mVacancyAreas[areaCode]; !ok {
    return nil, fmt.Errorf("specified wrong vacancy area")
  }
  experienceCode := strings.TrimSpace(parts[1])

  experience, ok := mVacancyExperiences[experienceCode]
  if !ok {
    return nil, fmt.Errorf("specified wrong vacancy experience")
  }
  experienceCode = experience

  keywords := strings.TrimSpace(parts[2])

  return &ParsedMessage{
    AreaCode:       areaCode,
    ExperienceCode: experienceCode,
    Keywords:       keywords,
  }, nil
}

var (
  mVacancyAreas = map[string]struct{}{
    "1": {},
    "2": {},
    "3": {},
  }
  mVacancyExperiences = map[string]string{
    "1": "between1And3",
    "2": "between3And6",
    "3": "noExperience",
  }
)
