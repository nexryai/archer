package archer

import (
	"net"
	"net/url"
	"strings"
)

func isPrivateAddress(address string) bool {
	ip := net.ParseIP(address)
	// パースできないなら安全ではない
	if ip == nil {
		return true
	}

	// 抜け穴になりそうなので6to4アドレスは拒否
	if strings.HasPrefix(address, "::ffff:0:") {
		return true
	} else if strings.HasPrefix(address, "::ffff:") {
		return true
	}

	if ip != nil && (ip.IsLoopback() ||
		ip.IsPrivate() ||
		ip.IsMulticast() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLoopback() ||
		ip.IsUnspecified() ||
		!ip.IsGlobalUnicast()) {
		return true
	}

	// netパッケージでなんか判定できないやつ (https://ipinfo.io/bogon)
	privateCIDRs := []string{
		"0.0.0.0/8",
		"100.64.0.0/10", // Tailscaleとかで使うやつ（IsPrivateで判定できないのバグな気がする）
		"64:ff9b::/96",
		"64:ff9b:1::/48",
		"2001:10::/28",
		"2001:db8::/32",
		"::/96",
	}

	for _, privateCIDR := range privateCIDRs {
		_, privateNet, err := net.ParseCIDR(privateCIDR)

		if err == nil && privateNet.Contains(ip) {
			return true
		}
	}

	// その他の条件が満たされない場合はパブリックアドレスとみなす
	return false
}

func IsSafeUrl(requestedUrl string) bool {
	// URLをパース
	parsedURL, err := url.Parse(requestedUrl)
	if err != nil {
		return false
	}

	// https以外は拒否
	if parsedURL.Scheme != "https" {
		return false
	}

	// UnixソケットとIPv6アドレス指定を拒否
	if strings.Contains(parsedURL.Hostname(), ":") {
		return false
	}

	// おかしいね
	if !strings.Contains(parsedURL.Hostname(), ".") {
		return false
	}

	// 認証情報を含むのは拒否
	if parsedURL.User != nil {
		return false
	}

	// ポート番号を取得
	port := parsedURL.Port()

	if port != "" && port != "80" && port != "443" {
		// 宛先が80と443以外ならブロック
		return false
	}

	// hostname検証
	hostname := parsedURL.Hostname()

	ip := net.ParseIP(hostname)
	if ip != nil {
		//IPアドレスが指定されている場合、それがプライベートアドレスならブロック
		if isPrivateAddress(ip.String()) {
			return false
		}
	}

	return true
}
