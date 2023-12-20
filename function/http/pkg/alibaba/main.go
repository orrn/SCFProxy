package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"github.com/aliyun/fc-runtime-go-sdk/fc"
	"io"
	"net/http"
	"time"
)

// RequestData 结构体用于解析输入的JSON数据
type RequestData struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Body    string            `json:"body"`
	Headers map[string]string `json:"headers"`
}

// ResponseData 结构体用于构建返回的JSON数据
type ResponseData struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Content    string            `json:"content"`
}

func Handle(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	var requestData RequestData

	// 解析请求体
	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 解码请求正文
	decodedBody, _ := base64.StdEncoding.DecodeString(requestData.Body)

	// 创建自定义的http.Client，忽略证书验证
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

	req, err := http.NewRequest(requestData.Method, requestData.URL, io.NopCloser(bytes.NewBuffer(decodedBody)))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 添加请求头
	for key, value := range requestData.Headers {
		req.Header.Add(key, value)
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// 读取响应体
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 构建响应数据
	responseData := ResponseData{
		StatusCode: resp.StatusCode,
		Headers:    make(map[string]string),
		Content:    base64.StdEncoding.EncodeToString(respBody),
	}
	for key, values := range resp.Header {
		responseData.Headers[key] = values[0]
	}

	// 返回JSON格式响应
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(responseData)
	if err != nil {
		// 记录或处理错误
		//log.Printf("Error encoding response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func main() {
	fc.StartHttp(Handle)
}
