package main

import (
	"bufio"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

var (
	gfwListUrl        = "https://raw.githubusercontent.com/gfwlist/gfwlist/master/gfwlist.txt"
	download          = flag.Bool("d", false, fmt.Sprintf("download latest gfwlist from: %s . if true will load gfwlist file from current dir.", gfwListUrl))
	output            = flag.String("o", "/etc/dnsmasq.d/gfwlist.conf", "output the convent result to file location.")
	withIpset         = flag.Bool("i", true, "convent with ipset, dnsmasq will add the dns result ip to ipset automaticly.")
	ipsetName         = flag.String("n", "gfwlist", "the ipset list name which you want.")
	clearOldIpsetList = flag.Bool("c", true, "try clear old ipset list if exists.")
	restartDnsmasq    = flag.Bool("r", true, "after convert try to restart dnsmasq service.")
	dnsHost           = flag.String("h", "127.0.0.1", "upstream dns host")
	dnsPort           = flag.Int("p", 5353, "upstream dns host port")

	usage = func() {
		fmt.Fprint(os.Stdout, "\nThe gfwlist to dnsmasq converter by yinheli @version 1.0.1 \n\n")
		fmt.Fprintf(os.Stdout, "Usage of %s\n", os.Args[0])
		flag.PrintDefaults()
	}

	lg = log.New(os.Stdout, "[gfw2dnsmasq] ", log.Ldate|log.Ltime)
)

func main() {
	flag.Usage = usage
	flag.Parse()

	if *download {
		lg.Println("start download file")
		downloadGfwlist()
	}

	lg.Println("start parse file")
	parse()

	if *clearOldIpsetList {
		lg.Println("clear ipset", *ipsetName)
		e1 := exec.Command("ipset", []string{"destroy", *ipsetName}...).Run()
		e2 := exec.Command("ipset", []string{"flush", *ipsetName}...).Run()

		if e1 != nil && e2 != nil {
			lg.Println("create ipset", *ipsetName)
			if err := exec.Command("ipset", []string{"create", *ipsetName, "hash:net"}...).Run(); err != nil {
				lg.Println("create ipset fail", err)
			}
		}

		// always add google dns to gfwlist
		exec.Command("ipset", []string{"add", *ipsetName, "8.8.8.8"}...).Run()
		exec.Command("ipset", []string{"add", *ipsetName, "8.8.4.4"}...).Run()
	}

	if *restartDnsmasq {
		lg.Println("restart dsnmasq service")
		if err := exec.Command("service", []string{"dnsmasq", "restart"}...).Run(); err != nil {
			lg.Println("restart dnsmasq service fail, error:", err)
		}
	}

}

func downloadGfwlist() {
	resp, err := http.Get(gfwListUrl)
	if err != nil {
		lg.Fatalln("down load gfwlist file faild", err)
	}
	defer resp.Body.Close()

	gfwlist, _ := os.Create("gfwlist.txt")
	io.Copy(gfwlist, base64.NewDecoder(base64.StdEncoding, resp.Body))
	gfwlist.Close()
}

func parse() {
	gfwlist, err := os.Open("gfwlist.txt")
	if err != nil {
		lg.Fatalln(err)
	}
	defer gfwlist.Close()

	outputFile, err := os.Create(*output)
	if err != nil {
		panic(err)
	}
	defer outputFile.Close()

	writer := bufio.NewWriter(outputFile)
	defer writer.Flush()
	reader := bufio.NewReader(gfwlist)
	lineSep := byte('\n')
	n := 0
	domainList := make([]string, 0)
	domainRegex := regexp.MustCompile(`^[a-zA-Z0-9\.\-]*$`)
	ipRegex := regexp.MustCompile(`^\d+\.\d+.\d+.\d+$`)
	writer.WriteString(fmt.Sprintf("# gfwlist \n# @%v\n\n", time.Now()))
	for {
		n++
		line, err := reader.ReadString(lineSep)
		if err == io.EOF {
			break
		} else if err != nil {
			panic(err)
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

		found := false
		for _, v := range domainList {
			if v == domain {
				found = true
			}
		}

		if !found {
			domainList = append(domainList, domain)
			comment := fmt.Sprintf("# line: %d %s", n, line)
			serverConfig := fmt.Sprintf("server=/%s/%s#%d", domain, *dnsHost, *dnsPort)
			ipsetConfig := ""
			if *withIpset {
				ipsetConfig = fmt.Sprintf("ipset=/%s/%s\n", domain, *ipsetName)
			}
			writer.WriteString(fmt.Sprintf("%s\n%s\n%s\n", comment, serverConfig, ipsetConfig))
		}
	}

	writer.WriteString(fmt.Sprintf("# finish domain count: %d", len(domainList)))
	lg.Println("parse file finish, count:", len(domainList), ", output file: ", *output)
	return
}
