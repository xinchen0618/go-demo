package gox

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"go.uber.org/zap"
)

// Restful 发起Restful API请求
//	@param method string
//	@param rawUrl string
//	@param params map[string]interface{} url参数或entity参数
//	@param headers map[string]string
//	@return body map[string]interface{}
//	@return httpCode int
//	@return err error
func Restful(method, rawUrl string, params map[string]interface{}, headers map[string]string) (body map[string]interface{}, httpCode int, err error) {
	method = strings.ToUpper(method)

	// 参数
	var entityParams io.Reader
	if len(params) > 0 {
		if InSlice(method, []string{"GET", "DELETE"}) { // 参数在url中
			urlParams := url.Values{}
			Url, err := url.Parse(rawUrl)
			if err != nil {
				zap.L().Error(err.Error())
				return nil, 0, err
			}
			for k, v := range params {
				urlParams.Set(k, fmt.Sprint(v))
			}
			Url.RawQuery = urlParams.Encode()
			rawUrl = Url.String()

		} else { // 参数在body中
			paramBytes, err := json.Marshal(params)
			if err != nil {
				zap.L().Error(err.Error())
				return nil, 0, err
			}
			entityParams = bytes.NewBuffer(paramBytes)
		}
	}

	req, err := http.NewRequest(method, rawUrl, entityParams)
	if err != nil {
		zap.L().Error(err.Error())
		return nil, 0, err
	}

	// Header
	if InSlice(method, []string{"POST", "PUT"}) {
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
		return nil, 0, err
	}
	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		zap.L().Error(err.Error())
		return nil, 0, err
	}
	body = map[string]interface{}{}
	if len(bodyBytes) > 0 {
		if err := json.Unmarshal(bodyBytes, &body); err != nil {
			zap.L().Error(err.Error())
			return nil, 0, err
		}
	}

	return body, resp.StatusCode, nil
}
