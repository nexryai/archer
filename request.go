package archer

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func filterIPv4(addrs []net.IP) []string {
	var ipv4Addrs []string
	for _, ip := range addrs {
		if strings.Contains(ip.String(), ".") {
			ipv4Addrs = append(ipv4Addrs, ip.String())
		}
	}
	return ipv4Addrs
}

// サイズ制限付きダウンローダー
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

type SecurityError struct {
	Message string
}

func (se *SecurityError) Error() string {
	return se.Message
}

type SecureRequest struct {
	Request *http.Request
	TimeOut int64
	MaxSize int64
}

func (sr *SecureRequest) Send() (*http.Response, error) {
	targetUrl := sr.Request.URL

	// sageなURLかチェック
	if !IsSafeUrl(targetUrl.String()) {
		return nil, errors.New("unsafe URL, download aborted")
	}

	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		// DualStack: true, // this is deprecated as of go 1.16
	}
	// or create your own transport, there's an example on godoc.
	http.DefaultTransport.(*http.Transport).DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		var connectTo net.IP

		i := net.ParseIP(targetUrl.Hostname())
		if i != nil {
			connectTo = i
		} else {
			fmt.Println(targetUrl.Hostname())
			ips, resolveErr := net.ResolveIPAddr("ip", targetUrl.Hostname())
			if resolveErr != nil {
				return nil, errors.New("failed to resolve hostname")
			}

			connectTo = ips.IP
		}

		if isPrivateAddress(connectTo.String()) {
			return nil, errors.New("private address, download aborted")
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
		//Transport: transport,
	}

	// リクエストを送信
	resp, err := client.Do(sr.Request)
	if err != nil {
		return nil, err
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
