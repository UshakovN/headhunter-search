package http

import (
  "context"
  "encoding/json"
  "fmt"
  "io"
  "main/pkg/utils"
  "net/http"
  "net/url"
  "time"
)

type Client struct {
  ctx    context.Context
  client *http.Client
}

func NewClient(ctx context.Context) *Client {
  return &Client{
    ctx:    ctx,
    client: &http.Client{},
  }
}

type Headers map[string]string

type Query map[string][]string

type Option func(*option)

type option struct {
  ctx     context.Context
  headers Headers
  query   Query
}

func (c *Client) Get(requestURL string, options ...Option) ([]byte, error) {
  const (
    retryCount = 10
    retryWait  = 3 * time.Second
  )
  var (
    buf []byte
    err error
  )
  if err = utils.DoWithRetries(retryCount, retryWait, func() error {
    buf, err = c.get(requestURL, options...)
    return err

  }); err != nil {
    return nil, err
  }
  return buf, nil
}

func (c *Client) get(requestURL string, options ...Option) ([]byte, error) {
  o := newOptions(options...)

  if ctx := o.ctx; ctx == nil {
    o.ctx = context.Background()
  }
  req, err := http.NewRequestWithContext(o.ctx, http.MethodGet, requestURL, nil)
  if err != nil {
    return nil, fmt.Errorf("cannot create http request with context for %s: %v", requestURL, err)
  }
  if h := o.headers; len(h) != 0 {
    req.Header = h.toHttpHeaders()
  }
  if q := o.query; len(q) != 0 {
    req.URL.RawQuery = url.Values(q).Encode()
  }
  resp, err := c.client.Do(req)
  if err != nil {
    return nil, fmt.Errorf("cannot do get request to %s: %v", requestURL, err)
  }
  if code := resp.StatusCode; code != http.StatusOK {
    return nil, fmt.Errorf("%w: got wrong status code from %s: %d", utils.ErrDoRetry, requestURL, code)
  }
  buf, err := io.ReadAll(resp.Body)
  if err != nil {
    return nil, fmt.Errorf("cannot read response body from %s: %v", requestURL, err)
  }
  return buf, nil
}

func WithContext(ctx context.Context) Option {
  return func(o *option) {
    o.ctx = ctx
  }
}

func WithQuery(query Query) Option {
  return func(o *option) {
    o.query = query
  }
}

func WithHeaders(headers Headers) Option {
  return func(o *option) {
    o.headers = headers
  }
}

func newOptions(options ...Option) *option {
  o := &option{}

  for _, option := range options {
    option(o)
  }
  return o
}

func (h Headers) toHttpHeaders() http.Header {
  headers := http.Header{}

  for key, val := range h {
    headers[key] = []string{val}
  }
  return headers
}

func UnmarshalResponse(buf []byte, resp any) error {
  if err := json.Unmarshal(buf, resp); err != nil {
    return fmt.Errorf("cannot unmarshal response json bytes to struct: %v", err)
  }
  return nil
}
