package management

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

func testNewRequest(repo *Repository, ctx context.Context, method string, path string, payload any) (*http.Request, error) {
	var body *bytes.Buffer

	body = &bytes.Buffer{}

	if payload != nil {
		if err := json.NewEncoder(body).Encode(payload); err != nil {
			return nil, err
		}
	}

	fullURL := repo.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, fullURL, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set(internalServiceHeader, repo.serviceName)
	req.Header.Set(internalTokenHeader, repo.token)

	if payload != nil && body.Len() > 0 {
		req.Header.Set(contentTypeHeader, contentTypeJSON)
	}

	return req, nil
}
