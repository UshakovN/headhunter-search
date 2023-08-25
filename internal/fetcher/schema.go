package fetcher

import (
	"encoding/json"
	"fmt"
	"main/pkg/http"
)

const vacanciesRequestURL = "https://api.hh.ru/vacancies"

type Request struct {
	Page        int    `json:"page,omitempty"`
	PerPage     int    `json:"per_page,omitempty"`
	Text        string `json:"text,omitempty"`
	SearchField string `json:"search_field,omitempty"`
	Experience  string `json:"experience,omitempty"`
	Employment  string `json:"employment,omitempty"`
	Area        string `json:"area,omitempty"`
	Period      int    `json:"period,omitempty"`
	DateFrom    string `json:"date_from,omitempty"`
	DateTo      string `json:"date_to,omitempty"`
}

func NewVacanciesRequest(text, area, experience string, period int) *Request {
	const (
		page        = 0
		perPage     = 100
		searchField = "name"
	)
	return &Request{

		Page:        page,
		PerPage:     perPage,
		SearchField: searchField,

		Text:       text,
		Area:       area,
		Experience: experience,
		Period:     period,
	}
}

func (r *Request) Query() (http.Query, error) {
	buf, err := json.Marshal(r)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal vacancies request to json: %v", err)
	}
	m := map[string]any{}

	if err = json.Unmarshal(buf, &m); err != nil {
		return nil, fmt.Errorf("cannot unmarshal vacancies response json to http query: %v", err)
	}
	q := http.Query{}

	for key, val := range m {
		q.Put(key, val)
	}
	return q, nil
}

type Response struct {
	Items []*VacancyResponseItem `json:"items"`
}

type VacancyResponseItem struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Area struct {
		Id   string `json:"id"`
		Name string `json:"name"`
		Url  string `json:"url"`
	} `json:"area"`
	Salary struct {
		From     int    `json:"from"`
		To       int    `json:"to"`
		Currency string `json:"currency"`
		Gross    bool   `json:"gross"`
	} `json:"salary"`
	Type struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	} `json:"type"`
	Address struct {
		City     string `json:"city"`
		Street   string `json:"street"`
		Building string `json:"building"`
		Raw      string `json:"raw"`
		Metro    struct {
			StationName string `json:"station_name"`
			LineName    string `json:"line_name"`
		} `json:"metro"`
		MetroStations []struct {
			StationName string `json:"station_name"`
			LineName    string `json:"line_name"`
		} `json:"metro_stations"`
		Id string `json:"id"`
	} `json:"address"`
	PublishedAt       string `json:"published_at"`
	CreatedAt         string `json:"created_at"`
	Archived          bool   `json:"archived"`
	ApplyAlternateUrl string `json:"apply_alternate_url"`
	Url               string `json:"url"`
	AlternateUrl      string `json:"alternate_url"`
	Employer          struct {
		Id           string `json:"id"`
		Name         string `json:"name"`
		Url          string `json:"url"`
		AlternateUrl string `json:"alternate_url"`
		LogoUrls     struct {
			Original string `json:"original"`
		} `json:"logo_urls"`
		VacanciesUrl         string `json:"vacancies_url"`
		AccreditedItEmployer bool   `json:"accredited_it_employer"`
		Trusted              bool   `json:"trusted"`
	} `json:"employer"`
	Snippet struct {
		Requirement    string `json:"requirement"`
		Responsibility string `json:"responsibility"`
	} `json:"snippet"`
	ProfessionalRoles []struct {
		Name string `json:"name"`
	} `json:"professional_roles"`
	Experience struct {
		Name string `json:"name"`
	} `json:"experience"`
	Employment struct {
		Name string `json:"name"`
	} `json:"employment"`
}
