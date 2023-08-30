package fetcher

import (
	"context"
	"fmt"
	"main/pkg/http"
)

type Fetcher interface {
	Fetch(context.Context, *Request) (*Response, error)
}

type fetcher struct {
	ctx    context.Context
	proxy  string
	client *http.Client
}

func NewFetcher(ctx context.Context, proxy string) Fetcher {
	return &fetcher{
		ctx:    ctx,
		proxy:  proxy,
		client: http.NewClient(ctx),
	}
}

func (f *fetcher) Fetch(ctx context.Context, req *Request) (*Response, error) {
	query, err := req.Query()
	if err != nil {
		return nil, fmt.Errorf("cannot got query from vacancies request: %v", err)
	}
	buf, err := f.client.Get(vacanciesRequestURL,
		http.WithContext(ctx),
		http.WithQuery(query),
		http.WithPrefix(f.proxy),
	)
	if err != nil {
		return nil, fmt.Errorf("cannot get request to %s: %v", vacanciesRequestURL, err)
	}
	resp := &Response{}

	if err = http.UnmarshalResponse(buf, resp); err != nil {
		return nil, fmt.Errorf("cannot unmarshal vacancies response: %v", err)
	}
	return resp, nil
}
