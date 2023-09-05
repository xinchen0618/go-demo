// Package gox Golang 增强函数
package gox

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/goccy/go-json"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

// RESTful 发起 RESTful 请求
//
//	params GET/DELETE 请求为 url 参数, 并会进行 url 转义, 其他请求为 entity 参数.
//	返回 body 是 json.Unmarshal 的数据.
func RESTful(method, rawUrl string, params map[string]any, headers map[string]string) (body map[string]any, httpCode int, err error) {
	method = strings.ToUpper(method)

	// 参数
	var entityParams io.Reader
	if len(params) > 0 {
		if lo.Contains([]string{"GET", "DELETE"}, method) { // url 参数
			urlParams := url.Values{}
			for k, v := range params {
				urlParams.Set(k, fmt.Sprint(v))
			}
			Url, err := url.Parse(rawUrl)
			if err != nil {
				zap.L().Error(err.Error())
				return map[string]any{}, 0, err
			}
			Url.RawQuery = urlParams.Encode()
			rawUrl = Url.String()

		} else { // entity 参数
			paramBytes, err := json.Marshal(params)
			if err != nil {
				zap.L().Error(err.Error())
				return map[string]any{}, 0, err
			}
			entityParams = bytes.NewBuffer(paramBytes)
		}
	}

	req, err := http.NewRequest(method, rawUrl, entityParams)
	if err != nil {
		zap.L().Error(err.Error())
		return map[string]any{}, 0, err
	}

	// Header
	if lo.Contains([]string{"POST", "PUT"}, method) {
		req.Header.Set("Content-Type", "application/json")
	}
	if len(headers) > 0 {
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		zap.L().Error(err.Error())
		return map[string]any{}, 0, err
	}
	defer func(Body io.ReadCloser) {
		if err := Body.Close(); err != nil {
			zap.L().Error(err.Error())
		}
	}(resp.Body)

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		zap.L().Error(err.Error())
		return map[string]any{}, 0, err
	}
	body = map[string]any{}
	if len(bodyBytes) > 0 {
		if err := json.Unmarshal(bodyBytes, &body); err != nil {
			zap.L().Error(err.Error())
			return map[string]any{}, 0, err
		}
	}

	return body, resp.StatusCode, nil
}
