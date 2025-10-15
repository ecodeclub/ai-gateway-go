package mvp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// TestLocalProxyServer 启动一个本地代理服务器，用于转发请求到 OpenAI API
func TestLocalProxyServer(t *testing.T) {
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		t.Fatal("未设置 API_KEY 环境变量")
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	port := os.Getenv("PROXY_PORT")
	if port == "" {
		port = "8000"
	}

	mux := http.NewServeMux()

	// ============ 通用代理处理器 ============
	proxyHandler := func(w http.ResponseWriter, r *http.Request) {
		// 启用 CORS
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// 读取请求体
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("读取请求体失败: %v", err), http.StatusBadRequest)
			return
		}
		r.Body.Close()

		// 构建目标 URL
		path := strings.TrimPrefix(r.URL.Path, "/api")
		targetURL := baseURL + path
		if r.URL.RawQuery != "" {
			targetURL += "?" + r.URL.RawQuery
		}

		log.Printf("🔄 代理请求: %s %s -> %s", r.Method, r.URL.Path, targetURL)
		if len(bodyBytes) > 0 && len(bodyBytes) < 2000 {
			log.Printf("📤 请求体: %s", string(bodyBytes))
		}

		// 创建新请求
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, r.Method, targetURL, bytes.NewReader(bodyBytes))
		if err != nil {
			http.Error(w, fmt.Sprintf("创建请求失败: %v", err), http.StatusInternalServerError)
			return
		}

		// 复制请求头
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("Content-Type", r.Header.Get("Content-Type"))
		if req.Header.Get("Content-Type") == "" {
			req.Header.Set("Content-Type", "application/json")
		}

		// 发送请求
		client := &http.Client{Timeout: 60 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, fmt.Sprintf("请求失败: %v", err), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// 读取响应
		respBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("读取响应失败: %v", err), http.StatusInternalServerError)
			return
		}

		log.Printf("✅ 响应状态: %d, 大小: %d 字节", resp.StatusCode, len(respBytes))
		if len(respBytes) < 2000 {
			log.Printf("📥 响应体: %s", string(respBytes))
		}

		// 复制响应头
		for k, v := range resp.Header {
			for _, vv := range v {
				w.Header().Add(k, vv)
			}
		}
		w.WriteHeader(resp.StatusCode)
		w.Write(respBytes)
	}

	// ============ 注册路由 ============
	// Chat Completions API
	mux.HandleFunc("/api/chat/completions", proxyHandler)

	// Responses API
	mux.HandleFunc("/api/responses", proxyHandler)

	// Conversations API
	mux.HandleFunc("/api/conversations", proxyHandler)
	mux.HandleFunc("/api/conversations/", proxyHandler) // 处理带 ID 的路由

	// Audio API
	mux.HandleFunc("/api/audio/transcriptions", proxyHandler)
	mux.HandleFunc("/api/audio/translations", proxyHandler)
	mux.HandleFunc("/api/audio/speech", proxyHandler)

	// Files API（Conversations 可能需要）
	mux.HandleFunc("/api/files", proxyHandler)
	mux.HandleFunc("/api/files/", proxyHandler)

	// 健康检查
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":   "ok",
			"base_url": baseURL,
			"time":     time.Now().Format(time.RFC3339),
		})
	})

	// 根路径
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>OpenAI API Proxy</title>
    <style>
        body { font-family: system-ui; max-width: 800px; margin: 50px auto; padding: 20px; }
        h1 { color: #10a37f; }
        .endpoint { background: #f7f7f7; padding: 10px; margin: 10px 0; border-radius: 5px; }
        code { background: #e0e0e0; padding: 2px 6px; border-radius: 3px; }
    </style>
</head>
<body>
    <h1>🚀 OpenAI API 代理服务器</h1>
    <p>正在运行，基础 URL: <code>%s</code></p>
    
    <h2>📌 可用端点：</h2>
    <div class="endpoint"><strong>Chat Completions:</strong> <code>POST /api/chat/completions</code></div>
    <div class="endpoint"><strong>Responses:</strong> <code>POST /api/responses</code></div>
    <div class="endpoint"><strong>Conversations (创建):</strong> <code>POST /api/conversations</code></div>
    <div class="endpoint"><strong>Conversations (获取):</strong> <code>GET /api/conversations/{id}</code></div>
    <div class="endpoint"><strong>Conversations (更新):</strong> <code>PATCH /api/conversations/{id}</code></div>
    <div class="endpoint"><strong>Conversations (删除):</strong> <code>DELETE /api/conversations/{id}</code></div>
    <div class="endpoint"><strong>Audio Transcriptions:</strong> <code>POST /api/audio/transcriptions</code></div>
    <div class="endpoint"><strong>Audio Translations:</strong> <code>POST /api/audio/translations</code></div>
    <div class="endpoint"><strong>Audio Speech:</strong> <code>POST /api/audio/speech</code></div>
    <div class="endpoint"><strong>健康检查:</strong> <code>GET /health</code></div>
    
    <h2>🌐 前端调用示例：</h2>
    <pre>
const response = await fetch('http://localhost:%s/api/chat/completions', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    model: 'gpt-4o-mini',
    messages: [{ role: 'user', content: 'Hello!' }]
  })
});
const data = await response.json();
    </pre>
</body>
</html>`, baseURL, port)
	})

	addr := ":" + port
	log.Printf("🚀 本地代理服务器启动于 http://localhost%s", addr)
	log.Printf("📡 转发目标: %s", baseURL)
	log.Printf("✅ 支持的端点:")
	log.Printf("   - POST /api/chat/completions")
	log.Printf("   - POST /api/responses")
	log.Printf("   - POST /api/conversations")
	log.Printf("   - GET  /api/conversations/{id}")
	log.Printf("   - PATCH /api/conversations/{id}")
	log.Printf("   - DELETE /api/conversations/{id}")
	log.Printf("   - POST /api/audio/transcriptions")
	log.Printf("   - POST /api/audio/translations")
	log.Printf("   - POST /api/audio/speech")
	log.Printf("   - GET  /health")
	log.Printf("💡 提示: 按 Ctrl+C 停止服务器")

	if err := http.ListenAndServe(addr, mux); err != nil {
		t.Fatalf("服务器启动失败: %v", err)
	}
}

// TestRemoteProxyServer 启动一个远程代理服务器，用于转发请求到 OpenAI API
func TestRemoteProxyServer(t *testing.T) {
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		t.Fatal("未设置 API_KEY 环境变量")
	}

	internalToken := os.Getenv("INTERNAL_TOKEN")
	if internalToken == "" {
		t.Fatal("未设置 INTERNAL_TOKEN 环境变量")
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		t.Fatal("未设置 BASE_URL 环境变量")
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	origin := os.Getenv("ORIGIN")
	if origin == "" {
		t.Fatal("未设置 ORIGIN 环境变量")
	}

	port := os.Getenv("PROXY_PORT")
	if port == "" {
		port = "8000"
	}

	mux := http.NewServeMux()

	// ============ 通用代理处理器 ============
	proxyHandler := func(w http.ResponseWriter, r *http.Request) {
		// 启用 CORS
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// 读取请求体
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("读取请求体失败: %v", err), http.StatusBadRequest)
			return
		}
		r.Body.Close()

		// 构建目标 URL
		path := strings.TrimPrefix(r.URL.Path, "/api")
		targetURL := baseURL + path
		if r.URL.RawQuery != "" {
			targetURL += "?" + r.URL.RawQuery
		}

		log.Printf("🔄 代理请求: %s %s -> %s", r.Method, r.URL.Path, targetURL)
		if len(bodyBytes) > 0 && len(bodyBytes) < 2000 {
			log.Printf("📤 请求体: %s", string(bodyBytes))
		}

		// 创建新请求
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(ctx, r.Method, targetURL, bytes.NewReader(bodyBytes))
		if err != nil {
			http.Error(w, fmt.Sprintf("创建请求失败: %v", err), http.StatusInternalServerError)
			return
		}

		// 复制请求头
		req.Header.Set("Authorization", "Bearer "+apiKey)
		req.Header.Set("X-Internal-Token", internalToken)
		req.Header.Set("Origin", origin)
		req.Header.Set("Content-Type", r.Header.Get("Content-Type"))
		if req.Header.Get("Content-Type") == "" {
			req.Header.Set("Content-Type", "application/json")
		}

		// 发送请求
		client := &http.Client{Timeout: 60 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, fmt.Sprintf("请求失败: %v", err), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// 读取响应
		respBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("读取响应失败: %v", err), http.StatusInternalServerError)
			return
		}

		log.Printf("✅ 响应状态: %d, 大小: %d 字节", resp.StatusCode, len(respBytes))
		if len(respBytes) < 2000 {
			log.Printf("📥 响应体: %s", string(respBytes))
		}

		// ✅ 复制响应头（跳过 CORS 头，避免重复）
		skipHeaders := map[string]bool{
			"Access-Control-Allow-Origin":      true,
			"Access-Control-Allow-Methods":     true,
			"Access-Control-Allow-Headers":     true,
			"Access-Control-Allow-Credentials": true,
			"Access-Control-Expose-Headers":    true,
			"Access-Control-Max-Age":           true,
		}

		for k, v := range resp.Header {
			if skipHeaders[k] {
				continue
			}
			for _, vv := range v {
				w.Header().Add(k, vv)
			}
		}

		w.WriteHeader(resp.StatusCode)
		w.Write(respBytes)
	}

	// ============ 注册路由 ============
	// Chat Completions API
	mux.HandleFunc("/api/chat/completions", proxyHandler)

	// Responses API
	mux.HandleFunc("/api/responses", proxyHandler)

	// Conversations API
	mux.HandleFunc("/api/conversations", proxyHandler)
	mux.HandleFunc("/api/conversations/", proxyHandler) // 处理带 ID 的路由

	// Audio API
	mux.HandleFunc("/api/audio/transcriptions", proxyHandler)
	mux.HandleFunc("/api/audio/translations", proxyHandler)
	mux.HandleFunc("/api/audio/speech", proxyHandler)

	// Files API（Conversations 可能需要）
	mux.HandleFunc("/api/files", proxyHandler)
	mux.HandleFunc("/api/files/", proxyHandler)

	// 健康检查
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":   "ok",
			"base_url": baseURL,
			"time":     time.Now().Format(time.RFC3339),
		})
	})

	// 根路径
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>OpenAI API Proxy</title>
    <style>
        body { font-family: system-ui; max-width: 800px; margin: 50px auto; padding: 20px; }
        h1 { color: #10a37f; }
        .endpoint { background: #f7f7f7; padding: 10px; margin: 10px 0; border-radius: 5px; }
        code { background: #e0e0e0; padding: 2px 6px; border-radius: 3px; }
    </style>
</head>
<body>
    <h1>🚀 OpenAI API 代理服务器</h1>
    <p>正在运行，基础 URL: <code>%s</code></p>
    
    <h2>📌 可用端点：</h2>
    <div class="endpoint"><strong>Chat Completions:</strong> <code>POST /api/chat/completions</code></div>
    <div class="endpoint"><strong>Responses:</strong> <code>POST /api/responses</code></div>
    <div class="endpoint"><strong>Conversations (创建):</strong> <code>POST /api/conversations</code></div>
    <div class="endpoint"><strong>Conversations (获取):</strong> <code>GET /api/conversations/{id}</code></div>
    <div class="endpoint"><strong>Conversations (更新):</strong> <code>PATCH /api/conversations/{id}</code></div>
    <div class="endpoint"><strong>Conversations (删除):</strong> <code>DELETE /api/conversations/{id}</code></div>
    <div class="endpoint"><strong>Audio Transcriptions:</strong> <code>POST /api/audio/transcriptions</code></div>
    <div class="endpoint"><strong>Audio Translations:</strong> <code>POST /api/audio/translations</code></div>
    <div class="endpoint"><strong>Audio Speech:</strong> <code>POST /api/audio/speech</code></div>
    <div class="endpoint"><strong>健康检查:</strong> <code>GET /health</code></div>
    
    <h2>🌐 前端调用示例：</h2>
    <pre>
const response = await fetch('http://localhost:%s/api/chat/completions', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    model: 'gpt-4o-mini',
    messages: [{ role: 'user', content: 'Hello!' }]
  })
});
const data = await response.json();
    </pre>
</body>
</html>`, baseURL, port)
	})

	addr := ":" + port
	log.Printf("🚀 远程代理服务器启动于 http://localhost%s", addr)
	log.Printf("📡 转发目标: %s", baseURL)
	log.Printf("✅ 支持的端点:")
	log.Printf("   - POST /api/chat/completions")
	log.Printf("   - POST /api/responses")
	log.Printf("   - POST /api/conversations")
	log.Printf("   - GET  /api/conversations/{id}")
	log.Printf("   - PATCH /api/conversations/{id}")
	log.Printf("   - DELETE /api/conversations/{id}")
	log.Printf("   - POST /api/audio/transcriptions")
	log.Printf("   - POST /api/audio/translations")
	log.Printf("   - POST /api/audio/speech")
	log.Printf("   - GET  /health")
	log.Printf("💡 提示: 按 Ctrl+C 停止服务器")

	if err := http.ListenAndServe(addr, mux); err != nil {
		t.Fatalf("服务器启动失败: %v", err)
	}
}
