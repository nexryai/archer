## archer
Simple http requester with basic anti-SSRF for go

### usage
```
req, err := http.NewRequest("GET", "https://google.com", nil)
if err != nil {
	log.Fatal(err)
}
SecureRequestService := archer.SecureRequest{
	Request: req,
	TimeOut: 10,
	MaxSize: 1024 * 1024 * 10,
}

resp, err := SecureRequestService.Send()
if err != nil {
    log.Fatal(err)
}

fmt.Println(resp.StatusCode)
```
### Security
URLs like the following would be safely rejected

```
# private address
"http://192.168.1.1"

# loopback address
"http://127.0.0.1"
"http://0.0.0.0"
"http://localhost"

# non-http port
"https://test.sda1.net:3000"

# non-https URL
"http://test.sda1.net"

# uncommon address
"https://unix:/var/run/super.sock"
"https://hogehost"

# URL with authentication information
"https://nexryai@google.com"
"https://nexryai:veryunsafe@sda1.net"

# Address that is not a global IP address
"https://::1/"
"https://[::1]/"
"https://100.85.142.25"
"https://169.254.169.254/"
```