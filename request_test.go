package archer

import (
	"net/http"
	"testing"
)

func testSecureRequest(t *testing.T, url string, expectError bool) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Errorf("failed to create request: %v", err)
	}
	SecureRequestService := SecureRequest{
		Request: req,
		TimeOut: 10,
		MaxSize: 1024 * 1024 * 10,
	}

	_, err = SecureRequestService.Send()
	if err != nil && !expectError {
		t.Errorf("unexpected error: %v", err)
	} else if err == nil && expectError {
		t.Errorf("expected error but got nil")
	}
}

func TestSecureRequest_Send(t *testing.T) {
	testSecureRequest(t, "https://[::ffff:255.255.255.255]/hoge", true)
	testSecureRequest(t, "https://localhost/hoge", true)
	testSecureRequest(t, "https://unix:/var/run/super.sock", true)
	testSecureRequest(t, "https://nexryai@google.com", true)
	testSecureRequest(t, "http://google.com", true)
	testSecureRequest(t, "https://google.com", false)

	req, err := http.NewRequest("GET", "https://s3.sda1.net/nyan/contents/127ba6c2-b0db-40a5-8e1d-56d32a17c9cf.jpg", nil)
	if err != nil {
		t.Errorf("failed to create request: %v", err)
	}
	SecureRequestService := SecureRequest{
		Request: req,
		TimeOut: 10,
		MaxSize: 1024 * 1024 * 10,
	}

	resp, err := SecureRequestService.Send()
	if err != nil {
		t.Errorf("failed to send request: %v", err)
	} else if resp.StatusCode != 200 {
		t.Errorf("failed to send request: %v", resp.Status)
	}
}
