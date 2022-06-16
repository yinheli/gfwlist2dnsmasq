package main

import (
    "bufio"
    "flag"
    "fmt"
    "os"
    "text/template"
    "time"

    gfwlist "github.com/yinheli/gfwlist2dnsmasq"
)

const templateString = `#{{.Now}}{{$host := .DNSHost}}{{$port := .DNSPort}}{{$ipset := .IpSet}}
{{range $domain := .Domains}}
server=/{{$domain}}/{{$host}}#{{$port}}{{if (ne $ipset "")}}
ipset=/{{$domain}}/{{$ipset}}{{end}}
{{end}}
#{{.Now}}`

var (
    ipset      = flag.String("n", "gfwlist", "the ipset list name which you want.")
    dnsHost    = flag.String("h", "127.0.0.1", "upstream dns host")
    dnsPort    = flag.Uint("p", 5353, "upstream dns host port")
    outputFile = flag.String("o", "/etc/dnsmasq.d/gfwlist.conf", "output dnsmasq configure file path")
)

func init() {
    flag.Parse()
}

func main() {
    list := gfwlist.New()

    gfwListUrl := os.Getenv("GFW_LIST")
    if gfwListUrl == "" {
        gfwListUrl = gfwlist.DefaultUrl
    }

    fmt.Printf("the gfwlist location is %s\n", gfwListUrl)
    domains, err := list.ParseFromUrl(gfwListUrl)
    if err != nil {
        panic(err)
    }

    data := struct {
        Domains        gfwlist.Domains
        IpSet, DNSHost string
        DNSPort        uint
        Now            time.Time
    }{
        Now:     time.Now(),
        Domains: domains,
        IpSet:   *ipset,
        DNSHost: *dnsHost,
        DNSPort: *dnsPort,
    }

    ur, err := template.New("").Parse(templateString)

    if err != nil {
        panic(err)
    }

    fp, err := os.Create(*outputFile)
    if err != nil {
        panic(err)
    }
    defer fp.Close()

    writer := bufio.NewWriter(fp)
    defer writer.Flush()

    err = ur.Execute(writer, data)
    if err != nil {
        panic(err)
    }

    fmt.Printf("dnsmasq configure has been write to %s", *outputFile)
}
