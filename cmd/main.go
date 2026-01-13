package main

import (
	"io"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"Web-Scrapper/internal/downloader"
	"Web-Scrapper/internal/refactor"
)

func main() {
	const webpageURL = "webpageURL"

	resp, err := http.Get(webpageURL)
	if err != nil {
		log.Fatal(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println(err)
		}
	}(resp.Body)

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	if err := downloader.DownloadIndex(doc); err != nil {
		log.Printf("ошибка загрузки индекса: %v\n", err)
	}

	if err := downloader.DownloadImages(doc, webpageURL); err != nil {
		log.Printf("ошибка загрузки изображений: %v\n", err)
	}

	if err := downloader.DownloadStylesNFonts(doc, webpageURL); err != nil {
		log.Printf("ошибка загрузки стилей и шрифтов: %v\n", err)
	}

	if err := refactor.CleanTrashScripts(doc); err != nil {
		log.Printf("ошибка очистки скриптов: %v\n", err)
	}

	if err := downloader.DownloadScripts(doc, webpageURL); err != nil {
		log.Printf("ошибка загрузки скриптов: %v\n", err)
	}

	if err := refactor.RefactorStyles(); err != nil {
		log.Printf("ошибка рефакторинга стилей: %v\n", err)
	}

	if err := refactor.RefactorImages(doc); err != nil {
		log.Printf("ошибка рефакторинга изображений: %v\n", err)
	}

	if err := refactor.RefactorGTMAcrum(doc); err != nil {
		log.Printf("ошибка добавления GTM: %v\n", err)
	}

	if err := refactor.RefactorAutoDomain(doc); err != nil {
		log.Printf("ошибка добавления автодомена: %v\n", err)
	}
}
