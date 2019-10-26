package fakehttp

import (
	"time"
)

const (
	txMethod = "POST"
	rxMethod = "GET"

	txFlag = "CDbHYQzabuNgrtgwSrC05w=="
	rxFlag = "CEjzOhPubJItdJA72O+6ETV+peA="

	targetUrl = "/"

	tokenCookieA = "cna"
	tokenCookieB = "_tb_token_"
	tokenCookieC = "_cna"

	userAgent = "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/47.0.2526.80 Safari/537.36 QQBrowser/9.3.6874.400"
	headerServer = "nginx"

	timeout = 10 * time.Second
	tokenTTL = 20 * time.Second
	tokenClean = 10 * time.Second
)



