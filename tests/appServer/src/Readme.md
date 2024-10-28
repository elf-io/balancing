# 测试协议

## 运行服务器和代理

### 运行所有服务器
```bash
go run ./http_server.go -port=8080
go run ./udp_server.go -port=8080
go run ./proxy_server.go -port=8090
```

## 测试 HTTP

### 直接访问 HTTP 服务器
使用 `curl` 发送 POST 请求到 HTTP 服务器，并传递 JSON 数据：
```bash
curl -X POST http://127.0.0.1:8080 -d '{"name": "tom"}' | jq .
```

### 通过代理访问 HTTP 服务器
使用 `curl` 发送 POST 请求到代理服务器，代理服务器将请求转发到 HTTP 服务器：
```bash
curl -X POST http://127.0.0.1:8090 -d '{"BackendUrl":"http://127.0.0.1:8080","Timeout":5,"ForwardType":"http", "EchoData":"Hello, HTTP!"}' | jq .
```

## 测试 UDP

### 直接访问 UDP 服务器
使用 `netcat` 发送数据到 UDP 服务器：
```bash
echo '{"name": "tom"}' | nc -u -w1 localhost 8080 | jq .
```

### 通过代理访问 UDP 服务器
使用 `curl` 发送 POST 请求到代理服务器，代理服务器将请求转发到 UDP 服务器：
```bash
curl -X POST http://127.0.0.1:8090 -d '{"BackendUrl":"127.0.0.1:8080","Timeout":5,"ForwardType":"udp", "EchoData":"Hello, UDP!"}' | jq .
```

## 注意事项

- 确保服务器和代理在运行时监听的端口与命令中指定的端口一致。
- 使用 `jq` 可以格式化 JSON 响应，便于阅读。
- 在测试 UDP 时，`netcat` 是一个常用的工具，可以用于发送和接收 UDP 数据包。
