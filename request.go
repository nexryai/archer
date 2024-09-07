package archer

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"
)

var (
	ErrPrivateAddressDetected = errors.New("private address detected. aborting request")
	ErrUnsafeUrlDetected      = errors.New("unsafe URL detected. request is blocked")
	ErrBlockedByDNS           = errors.New("blocked by DNS service")
)

// サイズ制限付きリーダー
type limitedReader struct {
	rc io.ReadCloser
	n  int64
}

func (lr *limitedReader) Read(p []byte) (int, error) {
	if lr.n <= 0 {
		return 0, io.EOF
	}
	if int64(len(p)) > lr.n {
		p = p[:lr.n]
	}
	n, err := lr.rc.Read(p)
	lr.n -= int64(n)
	return n, err
}

func (lr *limitedReader) Close() error {
	return lr.rc.Close()
}

// SecureRequest is a struct that holds a request, timeout, and max size.
type SecureRequest struct {
	Request     *http.Request
	TimeoutSecs int64
	MaxSize     int64
}

// Send sends the request and returns the response. Before sending the request, it checks if the URL is safe from SSRF attacks.
func (sr *SecureRequest) Send() (*http.Response, error) {
	targetUrl := sr.Request.URL

	// safeなURLかチェック
	if !IsSafeUrl(targetUrl.String()) {
		return nil, ErrUnsafeUrlDetected
	}

	dialer := &net.Dialer{
		Timeout:   time.Duration(sr.TimeoutSecs) * time.Second,
		KeepAlive: 30 * time.Second,
	}

	// Create custom transport that checks if the IP address is private
	http.DefaultTransport.(*http.Transport).DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		var connectTo net.IP

		i := net.ParseIP(targetUrl.Hostname())
		if i != nil {
			connectTo = i
		} else {
			ips, resolveErr := net.ResolveIPAddr("ip", targetUrl.Hostname())
			if resolveErr != nil {
				return nil, errors.New("failed to resolve hostname")
			}

			connectTo = ips.IP
		}

		if isPrivateAddress(connectTo.String()) {
			return nil, ErrPrivateAddressDetected
		}

		_, port, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}

		addr = net.JoinHostPort(connectTo.String(), port)

		return dialer.DialContext(ctx, network, addr)
	}

	// リクエストを作成
	client := &http.Client{
		Timeout: time.Duration(sr.TimeoutSecs) * time.Second,
	}

	// リクエストを送信
	resp, err := client.Do(sr.Request)
	if err != nil {
		return nil, err
	}

	if resp.Header.Get("Blocked-By") == "NextDNS" {
		return nil, ErrBlockedByDNS
	}

	// ファイルサイズが制限を超えているかチェック
	contentLength := resp.Header.Get("Content-Length")
	if contentLength != "" {
		length, err := strconv.ParseInt(contentLength, 10, 64)
		if err != nil {
			return nil, err
		}
		if length > sr.MaxSize {
			return nil, errors.New("file size exceeds the limit")
		}
	}

	resp.Body = &limitedReader{rc: resp.Body, n: sr.MaxSize}
	return resp, nil
}
