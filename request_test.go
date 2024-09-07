package archer

import (
	"errors"
	"net/http"
	"testing"
)

func testSecureRequest(t *testing.T, url string, expectError error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Errorf("failed to create request: %v", err)
	}
	SecureRequestService := SecureRequest{
		Request:     req,
		TimeoutSecs: 10,
		MaxSize:     1024 * 1024 * 10,
	}

	_, err = SecureRequestService.Send()
	if !errors.Is(err, expectError) {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestSecureRequest_Send(t *testing.T) {
	testSecureRequest(t, "https://[::ffff:255.255.255.255]/hoge", ErrUnsafeUrlDetected)
	testSecureRequest(t, "https://localhost/hoge", ErrUnsafeUrlDetected)
	testSecureRequest(t, "https://unix:/var/run/super.sock", ErrUnsafeUrlDetected)
	testSecureRequest(t, "https://nexryai@google.com", ErrUnsafeUrlDetected)
	testSecureRequest(t, "http://google.com", ErrUnsafeUrlDetected)
	testSecureRequest(t, "https://local1.sda1.net", ErrPrivateAddressDetected)
	testSecureRequest(t, "https://google.com", nil)
}
