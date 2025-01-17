package m3u8

import (
	"bufio"
	"regexp"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	pgt "github.com/nicoxiang/geektime-downloader/internal/pkg/geektime"
	"github.com/nicoxiang/geektime-downloader/internal/pkg/logger"
)

var (
	client *resty.Client
	// regex pattern for extracting `key=value` parameters from a line
	linePattern = regexp.MustCompile(`([a-zA-Z-]+)=("[^"]+"|[^",]+)`)
)

func init() {
	client = resty.New().
		SetRetryCount(1).
		SetTimeout(10*time.Second).
		SetHeader(pgt.UserAgentHeaderName, pgt.UserAgentHeaderValue).
		SetLogger(logger.DiscardLogger{})
}

// Parse do m3u8 url GET request, and extract ts file names and decrypt key from that
func Parse(m3u8url string) (tsFileNames []string, keyURI string, err error) {
	m3u8Resp, err := client.R().SetDoNotParseResponse(true).Get(m3u8url)
	if err != nil {
		return nil, "", err
	}
	defer m3u8Resp.RawBody().Close()
	s := bufio.NewScanner(m3u8Resp.RawBody())
	var lines []string
	for s.Scan() {
		lines = append(lines, s.Text())
	}

	gotKeyURI := false

	for _, line := range lines {
		// geektime video ONLY has one EXT-X-KEY
		if strings.HasPrefix(line, "#EXT-X-KEY") && !gotKeyURI {
			// ONLY Method and URI, IV not present
			params := parseLineParameters(line)
			keyURI, gotKeyURI = params["URI"], true
		}
		if !strings.HasPrefix(line, "#") && strings.HasSuffix(line, ".ts") {
			tsFileNames = append(tsFileNames, line)
		}
	}

	return
}

// parseLineParameters extra parameters in string `line`
func parseLineParameters(line string) map[string]string {
	r := linePattern.FindAllStringSubmatch(line, -1)
	params := make(map[string]string)
	for _, arr := range r {
		params[arr[1]] = strings.Trim(arr[2], "\"")
	}
	return params
}
