package archer

import (
	"testing"
)

func testIsPrivateIP(t *testing.T, ip string, expectedResult bool) {
	result := isPrivateAddress(ip)
	if result != expectedResult {
		t.Errorf("isPrivateAddress(%s) = %t, expected %t", ip, result, expectedResult)
	}
}

func TestIsPrivateIP(t *testing.T) {
	privateIPs := []string{
		// loopback address
		"127.0.0.1",
		"127.0.0.53",
		"0.0.0.0",
		"0.255.255.255",
		"::1",
		"localhost",
		// private address
		"192.168.1.1",
		"100.127.255.255",
		"172.16.0.1",
		"100.85.142.25",
		// multicast address
		"169.254.169.254",
		// link local unicast address
		"255.255.255.255",
		// IPv6
		"::",
		"fe80::",
		"ff00::",
		"::ffff:255.255.255.255",
		"::ffff:0:255.255.255.255",
		"100::ffff:ffff:ffff:ffff ",
		"64:ff9b::255.255.255.255",
		"64:ff9b:1:ffff:ffff:ffff:ffff:ffff",
		"ff00::",
		"ffff:ffff:ffff:ffff:ffff:ffff:ffff:ffff",
		"fd7a:115c:a1e0::48bf:643e",
	}

	publicIPs := []string{
		"1.1.1.1",
		"8.8.8.8",
		"2606:4700:4700::1111",
		"2606:4700:4700::1113",
	}

	for _, ip := range privateIPs {
		testIsPrivateIP(t, ip, true)
	}

	for _, ip := range publicIPs {
		testIsPrivateIP(t, ip, false)
	}
}

func testIsSafeUrlResult(t *testing.T, url string, expectedResult bool) {
	result := IsSafeUrl(url)
	if result != expectedResult {
		t.Errorf("isSafeUrl(%s) = %t, expected %t", url, result, expectedResult)
	}
}

func TestIsSafeUrl(t *testing.T) {
	unsafeUrls := []string{
		"https://fd7a:115c:a1e0::48bf:643e",
		"http://192.168.1.1",
		"http://127.0.0.1",
		"http://0.0.0.0",
		"http://localhost",
		"http://192.168.1.1:8080",
		"https://test.sda1.net:3000",
		"https://1.1.1.1:3000",
		"https://unix:/var/run/super.sock",
		"https://hogehost",
		"http://fugehost",
		"https://nexryai@google.com",
		"https://nexryai:veryunsafe@sda1.net",
		"https://::1/",
		"https://[::1]/",
		"https://100.85.142.25",
		"https://169.254.169.254/",
	}

	safeUrls := []string{
		"https://google.com",
		"https://sda1.net:443",
	}

	for _, url := range unsafeUrls {
		testIsSafeUrlResult(t, url, false)
	}

	for _, url := range safeUrls {
		testIsSafeUrlResult(t, url, true)
	}
}
