package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
)

const (
	TorProxyAddr = "socks5://127.0.0.1:9150"
	InputFile    = "urls.yaml"
	OutputDir    = "screenshots"
)

func main() {
	fmt.Println("--- TOR Scraper Başladı ---")

	os.MkdirAll(OutputDir, 0755)

	urls, err := readTargets(InputFile)
	if err != nil {
		log.Fatal(err)
	}

	for _, url := range urls {
		fmt.Println("[*] Screenshot alınıyor:", url)

		err := captureScreenshot(url)
		if err != nil {
			fmt.Println("[ERR]", url, "->", err)
			continue
		}

		fmt.Println("[OK]", url)
	}
}


func captureScreenshot(target string) error {
	if !strings.HasPrefix(target, "http") {
		target = "http://" + target
	}

	opts := append(
		chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("proxy-server", TorProxyAddr),
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	var buf []byte

	err := chromedp.Run(ctx,
		chromedp.Navigate(target),
		chromedp.Sleep(10*time.Second), 
		chromedp.FullScreenshot(&buf, 90),
	)
	if err != nil {
		return err
	}

	fileName := OutputDir + "/" + sanitize(target) + ".png"
	return os.WriteFile(fileName, buf, 0644)
}

// -------------------------------------------------

func readTargets(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var urls []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		urls = append(urls, strings.TrimPrefix(line, "- "))
	}
	return urls, sc.Err()
}

// -------------------------------------------------

func sanitize(u string) string {
	r := strings.NewReplacer(
		"http://", "",
		"https://", "",
		"/", "_",
		".", "_",
		":", "",
	)
	return r.Replace(u)
}
