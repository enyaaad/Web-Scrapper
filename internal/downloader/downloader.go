package downloader

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"Web-Scrapper/internal/parser"
	"Web-Scrapper/internal/utils"
)

func DownloadIndex(doc *goquery.Document) error {
	fileName := utils.IndexFileName

	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("ошибка при создании index.html: %w", err)
	}
	defer file.Close()

	err = utils.RenderHTML(doc.Selection, file)
	if err != nil {
		return fmt.Errorf("ошибка при сохранении index.html: %w", err)
	}
	
	fmt.Println("index сохранен:", fileName)
	return nil
}

func DownloadImages(doc *goquery.Document, websiteURL string) error {
	if err := utils.EnsureDir("images"); err != nil {
		return err
	}

	var imageURLs []string

	doc.Find("img").Each(func(index int, s *goquery.Selection) {
		href, exists := s.Attr("src")

		tempHref, is := s.Attr("data-src")
		if is {
			href = tempHref
			s.RemoveAttr("data-src")
		}

		if !exists {
			return
		}

		normalizedURL, err := utils.NormalizeImageURL(href, websiteURL)
		if err != nil {
			fmt.Printf("ошибка нормализации URL изображения: %v\n", err)
			return
		}

		fileName := "images/" + path.Base(normalizedURL)

		if strings.Contains(normalizedURL, "base64") {
			return
		}

		imageURLs = append(imageURLs, normalizedURL)
		s.SetAttr("src", fileName)
	})

	runtime.GOMAXPROCS(4)

	var wg sync.WaitGroup
	for _, imageURL := range imageURLs {
		wg.Add(1)
		imgURLTemp := imageURL
		go func(url string) {
			defer wg.Done()
			downloadImage(url, "images/"+path.Base(url))
		}(imageURL)
	}
	wg.Wait()

	htmlString, err := doc.Html()
	if err != nil {
		return fmt.Errorf("ошибка при генерации html: %w", err)
	}
	return utils.SaveHTMLtoFile(htmlString)
}

func downloadImage(imageURL, fileName string) {
	resp, err := http.Get(imageURL)
	if err != nil {
		fmt.Printf("ошибка при загрузке фотографии: %v, %s\n", err, imageURL)
		return
	}
	defer resp.Body.Close()

	file, err := os.Create(fileName)
	if err != nil {
		fmt.Printf("ошибка при создании фотографии: %v\n", err)
		return
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		fmt.Printf("ошибка при сохранении фотографии: %v\n", err)
		return
	}

	fmt.Println("фотография сохранена:", fileName)
}

func DownloadStylesNFonts(doc *goquery.Document, websiteURL string) error {
	if err := utils.EnsureDir("styles"); err != nil {
		return err
	}

	doc.Find("link[rel='stylesheet']").Each(func(index int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			return
		}

		normalizedURL, err := utils.NormalizeStyleURL(href, websiteURL)
		if err != nil {
			fmt.Printf("ошибка нормализации URL стиля: %v\n", err)
			return
		}

		fmt.Println(normalizedURL)
		resp, err := http.Get(normalizedURL)
		if err != nil {
			fmt.Printf("ошибка при загрузке стиля: %v, %s\n", err, normalizedURL)
			return
		}
		defer resp.Body.Close()

		fileName := path.Base(normalizedURL)
		if len(fileName) < 3 || fileName[len(fileName)-3:] != "css" {
			fileName = fmt.Sprintf("styles/font%d.css", index)
		} else {
			fileName = "styles/" + fileName
		}

		file, err := os.Create(fileName)
		if err != nil {
			fmt.Printf("ошибка при создании файла: %v\n", err)
			return
		}

		_, err = io.Copy(file, resp.Body)
		if err != nil {
			log.Printf("ошибка при копировании файла: %v\n", err)
		}

		fmt.Println("стиль сохранен:", fileName)
		
		file.Close()
		
		DownloadFonts(fileName, websiteURL)
		DownloadImagesFromStyles(fileName, websiteURL)
		s.SetAttr("href", fileName)
	})

	htmlString, err := doc.Html()
	if err != nil {
		return fmt.Errorf("ошибка при генерации html: %w", err)
	}
	return utils.SaveHTMLtoFile(htmlString)
}

func DownloadFonts(fromFileName, websiteURL string) {
	if err := utils.EnsureDir("fonts"); err != nil {
		fmt.Printf("ошибка создания папки fonts: %v\n", err)
		return
	}

	file, err := os.Open(fromFileName)
	if err != nil {
		fmt.Printf("ошибка открытия файла: %v\n", err)
		return
	}
	defer file.Close()

	fontURLs, err := parser.FindFontURLs(file)
	if err != nil {
		fmt.Printf("ошибка поиска URL шрифтов: %v\n", err)
		return
	}

	var wg sync.WaitGroup
	for _, fontURL := range fontURLs {
		wg.Add(1)
		fontURLTemp := fontURL
		go func(url string) {
			defer wg.Done()
			downloadFont(fontURLTemp, websiteURL)
		}(fontURL)
	}
	wg.Wait()
	fmt.Println("загрузка шрифтов завершена")
}

func downloadFont(fontURL, websiteURL string) {
		normalizedURL, err := utils.NormalizeFontURL(fontURL, websiteURL)
	if err != nil {
		fmt.Printf("ошибка нормализации URL шрифта: %v\n", err)
		return
	}

	resp, err := http.Get(normalizedURL)
	if err != nil {
		fmt.Printf("ошибка при загрузке шрифта: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fileName := "fonts/" + path.Base(normalizedURL)
	file, err := os.Create(fileName)
	if err != nil {
		fmt.Printf("ошибка при создании файла: %v\n", err)
		return
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		fmt.Printf("ошибка при сохранении шрифта: %v\n", err)
		return
	}
	fmt.Println("шрифт сохранен:", fileName)
}

func DownloadImagesFromStyles(fromFileName, websiteURL string) {
	file, err := os.Open(fromFileName)
	if err != nil {
		fmt.Printf("ошибка открытия файла: %v\n", err)
		return
	}
	defer file.Close()

	imageURLs, err := parser.FindImageURLs(file)
	if err != nil {
		fmt.Printf("ошибка поиска URL изображений: %v\n", err)
		return
	}

	var wg sync.WaitGroup
	for _, imageURL := range imageURLs {
		wg.Add(1)
		imgURLTemp := imageURL
		go func(url string) {
			defer wg.Done()
			downloadStyleImage(imgURLTemp, websiteURL)
		}(imageURL)
	}
	wg.Wait()
	fmt.Println("загрузка картинки из стиля завершена")
}

func downloadStyleImage(imageURL, websiteURL string) {
		normalizedURL, err := utils.NormalizeStyleImageURL(imageURL, websiteURL)
		if err != nil {
			fmt.Printf("ошибка нормализации URL изображения: %v\n", err)
			return
		}

	resp, err := http.Get(normalizedURL)
	if err != nil {
		fmt.Printf("ошибка при загрузке картинки из стилей: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fileName := "images/" + path.Base(normalizedURL)
	file, err := os.Create(fileName)
	if err != nil {
		fmt.Printf("ошибка при создании файла: %v\n", err)
		return
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		fmt.Printf("ошибка при сохранении картинки из стиля: %v\n", err)
		return
	}
	fmt.Println("картинка сохранена:", fileName)
}

func DownloadScripts(doc *goquery.Document, websiteURL string) error {
	if err := utils.EnsureDir("script"); err != nil {
		return err
	}

	doc.Find("script").Each(func(index int, s *goquery.Selection) {
		href, exists := s.Attr("src")
		if !exists {
			return
		}

		normalizedURL, err := utils.NormalizeScriptURL(href, websiteURL)
		if err != nil {
			fmt.Printf("ошибка нормализации URL скрипта: %v\n", err)
			return
		}

		resp, err := http.Get(normalizedURL)
		if err != nil {
			fmt.Printf("ошибка при загрузке скриптов: %v, %s\n", err, normalizedURL)
			return
		}
		defer resp.Body.Close()

		fileName := "script/" + path.Base(normalizedURL)
		if idx := strings.IndexByte(fileName, '?'); idx >= 0 {
			fileName = fileName[:idx]
		}

		file, err := os.Create(fileName)
		if err != nil {
			fmt.Printf("ошибка при создании скрипта: %v\n", err)
			return
		}
		defer file.Close()

		_, err = io.Copy(file, resp.Body)
		if err != nil {
			fmt.Printf("ошибка при сохранении скрипта: %v\n", err)
			return
		}
		fmt.Println("скрипт сохранен:", fileName)
		s.SetAttr("src", fileName)
	})

	htmlString, err := doc.Html()
	if err != nil {
		return fmt.Errorf("ошибка при генерации html: %w", err)
	}
	return utils.SaveHTMLtoFile(htmlString)
}
