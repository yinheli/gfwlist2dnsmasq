package gfwlist

import (
    "bufio"
    "encoding/base64"
    "fmt"
    "io"
    "net/http"
    "os"
    "regexp"
    "strings"

    "golang.org/x/net/proxy"
)

const DefaultUrl = "https://raw.githubusercontent.com/gfwlist/gfwlist/master/gfwlist.txt"

type Domains []string

type gfwlist struct {
    domains Domains
}

func New() gfwlist {
    return gfwlist{}
}

// proxyClient to create http client with socks5 proxy
func (g *gfwlist) proxyClient(addr string) (*http.Client, error) {
    // setup a http client
    httpTransport := &http.Transport{
        Proxy: http.ProxyFromEnvironment,
    }

    // create a socks5 dialer
    dialer, err := proxy.SOCKS5("tcp", addr, nil, proxy.Direct)
    if err != nil {
        return nil, err
    }

    // set our socks5 as the dialer
    if contextDialer, ok := dialer.(proxy.ContextDialer); ok {
        httpTransport.DialContext = contextDialer.DialContext
    }

    return &http.Client{
        Transport: httpTransport,
    }, nil
}

func (g *gfwlist) ParseFromString(data string) (Domains, error) {
    reader := bufio.NewReader(strings.NewReader(data))
    return g.ParseFromReader(reader)
}

func (g *gfwlist) ParseFromFile(path string) (Domains, error) {
    fp, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer fp.Close()

    return g.ParseFromReader(bufio.NewReader(fp))
}

// ParseFromUrl to fetch gfwlist from specified url
func (g *gfwlist) ParseFromUrl(url string) (Domains, error) {
    var resp *http.Response
    var err error

    proxy := os.Getenv("SOCKS5_PROXY")
    if proxy != "" {
        client, err := g.proxyClient(proxy)
        if err != nil {
            return nil, err
        }
        resp, err = client.Get(url)
    } else {
        resp, err = http.Get(url)
    }

    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    reader := base64.NewDecoder(base64.StdEncoding, resp.Body)
    return g.ParseFromReader(bufio.NewReader(reader))
}

func (g *gfwlist) ParseFromReader(reader *bufio.Reader) (Domains, error) {
    lineSep := byte('\n')
    n := 0

    g.domains = make(Domains, 0)

    domainRegex := regexp.MustCompile(`^[a-zA-Z0-9\.\-]*$`)
    ipRegex := regexp.MustCompile(`^\d+\.\d+.\d+.\d+$`)

    for {
        n++
        line, err := reader.ReadString(lineSep)
        if err == io.EOF {
            break
        } else if err != nil {
            return nil, err
        }

        line = strings.TrimSpace(line)

        if len(line) == 0 ||
            strings.HasPrefix(line, "!") ||
            strings.HasPrefix(line, "/") ||
            strings.HasPrefix(line, "@@") ||
            strings.HasPrefix(line, "[") {
            continue
        }

        domain := ""
        if strings.HasPrefix(line, "||") {
            domain = line[2:]
        } else if strings.HasPrefix(line, "|") {
            continue
        } else {
            domain = line
        }

        if domain == "" {
            continue
        }

        if strings.Index(domain, "*") >= 0 {
            continue
        }

        if strings.Index(domain, ".") == -1 {
            continue
        }

        if !domainRegex.MatchString(domain) {
            continue
        }

        if ipRegex.MatchString(domain) {
            continue
        }

        if !strings.HasPrefix(domain, ".") {
            domain = fmt.Sprintf(".%s", domain)
        }

        g.domains = append(g.domains, domain)
    }

    return g.domains, nil
}
