package server

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"io"
	"net/http"
	"time"
)

// Handle 处理API网关请求
func Handle(ctx context.Context, event EventRequest) (*EventResponse, error) {
	// 解析经Base64编码的请求体
	decodedBody, err := base64.StdEncoding.DecodeString(event.Body)
	if err != nil {
		return &EventResponse{StatusCode: 400}, err
	}

	// 准备HTTP请求
	req, err := http.NewRequest(event.Method, event.Path, io.NopCloser(bytes.NewReader(decodedBody)))
	if err != nil {
		return &EventResponse{StatusCode: 400}, err
	}

	// 添加请求头
	for key, value := range event.Headers {
		req.Header.Add(key, value)
	}

	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return &EventResponse{StatusCode: 500}, err
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return &EventResponse{StatusCode: 500}, err
	}

	// Encode the response body in base64
	encodedBody := base64.StdEncoding.EncodeToString(respBody)

	// 处理响应头
	headers := make(map[string]string)
	for key, values := range resp.Header {
		headers[key] = values[0] // 取第一个值
	}

	// 构造响应
	return &EventResponse{
		StatusCode:      resp.StatusCode,
		Headers:         headers,
		Body:            encodedBody,
		IsBase64Encoded: false,
		//Body:            base64.StdEncoding.EncodeToString(respBody),
		//IsBase64Encoded: true,
	}, nil
}
