package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	"io"
	"log"
	"net/http"
	url "net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	const webpageUrl = "https://quick-fit-tr.online-shop2023.com/?lttracking=6afa6af8d59715e0b35f83abbefc74a0&ltpostclick=1691291976&source=leadtrade&ltsource=46622&lthash=ElJXN&landing=0&st=0KLQuNC30LXRgNC90YvQtSDRgdC10YLQuA%3D%3D&offer_id=8212&s1=&s2=&s3=&s4="

	resp, err := http.Get(webpageUrl)

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

	log.Println(doc)

	downloadIndex(doc)
	downloadImages(doc, webpageUrl)
	downloadStylesNFonts(doc, webpageUrl)
	downloadScripts(doc, webpageUrl)
	refactorStyles()

}

func getDomain(websiteUrl string) (string, error) {
	parsedURL, err := url.Parse(websiteUrl)
	if err != nil {
		fmt.Println("ошибка при парсе URL: ", err)
		return "ошибка при парсе URL: ", err
	}
	return parsedURL.Hostname(), nil
}

func saveHTMLtoFile(htmlString string) error {
	fileName := "index.html"

	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	_, err = file.WriteString(htmlString)
	if err != nil {
		return err
	}

	return nil
}

func downloadIndex(doc *goquery.Document) {
	fileName := "index.html"

	file, err := os.Create(fileName)
	if err != nil {
		fmt.Println("ошибка при создании index.html: ", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	err = renderHTML(doc.Selection, file)

	if err != nil {
		fmt.Println("ошибка при сохранении index.html: ", err)
		return
	}
	fmt.Println("index сохранен:", fileName)
}

func renderHTML(n *goquery.Selection, w io.Writer) error {
	for _, node := range n.Nodes {
		err := html.Render(w, node)
		if err != nil {
			return err
		}
	}
	return nil
}

func downloadStylesNFonts(doc *goquery.Document, websiteUrl string) {
	err := os.Mkdir("styles", os.ModePerm)
	if err != nil && !os.IsExist(err) {
		fmt.Println("ошибка при создании папки: ", err)
	}
	doc.Find("link[rel='stylesheet']").Each(func(index int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists {
			if strings.HasPrefix(href, "//") {
				href = "https:" + href
				log.Println("успешно " + href)
			} else if strings.HasPrefix(href, "/") {
				domain, _ := getDomain(websiteUrl)
				href = "https://" + domain + href
				log.Println("успешно " + href)
			} else if strings.Contains(href[0:5], "css/") || strings.Contains(href[0:8], "styles/") || strings.Contains(href[0:7], "style/") {
				domain, _ := getDomain(websiteUrl)
				href = "https://" + domain + "/" + href
			}
		} else {
			return
		}

		resp, err := http.Get(href)
		if err != nil {
			fmt.Println("ошибка при загрузке стиля: ", err, href)
			return
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {

			}
		}(resp.Body)

		fileName := path.Base(href)

		if fileName[len(fileName)-3:] != "css" {
			fileName = "styles/font" + fmt.Sprintf("%d", index) + ".css"
		} else {
			fileName = "styles/" + fileName
		}

		file, err := os.Create(fileName)
		if err != nil {
			fmt.Println("ошибка при создании файла: ", err)
		}

		defer func(file *os.File) {
			err := file.Close()
			if err != nil {

			}
		}(file)

		_, err = io.Copy(file, resp.Body)
		if err != nil {
			fmt.Println("ошибка при сохранении стиля: ", err)
			return
		}
		fmt.Println("стиль сохранен:", fileName)
		downloadFonts(fileName, websiteUrl)
		downloadImagesFromStyles(fileName, websiteUrl)
		s.SetAttr("href", fileName)

		err = file.Close()
		if err != nil {
			return
		}
	})
	htmlString, err := doc.Html()
	if err != nil {
		fmt.Println("ошибка при генерации html: ", err)
		return
	}
	err = saveHTMLtoFile(htmlString)

}

func downloadFonts(fromFileName string, websiteUrl string) {
	err := os.Mkdir("fonts", os.ModePerm)

	filefrom, err := os.Open(fromFileName)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(filefrom)
	all := io.ReadCloser(filefrom)
	if err != nil {
		return
	}

	fontURLs := findFontUrls(all)

	for _, fontURL := range fontURLs {

		if strings.HasPrefix(fontURL, "//") {
			fontURL = "https:" + fontURL
			fmt.Println(fontURL)
		} else if strings.HasPrefix(fontURL, "/") {
			fontURL = websiteUrl + fontURL
			fmt.Println(fontURL)
		} else if strings.HasPrefix(fontURL, "../") || strings.HasPrefix(fontURL, "./") {
			domain, _ := getDomain(websiteUrl)
			fontURL = "https://" + domain + strings.Replace(fontURL, ".", "", 2)
			fmt.Println(fontURL)
		}

		resp, err := http.Get(fontURL)
		if err != nil {
			fmt.Println("ошибка при загрузке шрифта: ", err)
			return
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				panic("panic")
			}
		}(resp.Body)

		fileName := "fonts/" + path.Base(fontURL)

		file, err := os.Create(fileName)
		if err != nil {
			fmt.Println("ошибка при создании файла: ", err)
		}

		defer func(file *os.File) {
			err := file.Close()
			if err != nil {

			}
		}(file)

		_, err = io.Copy(file, resp.Body)
		if err != nil {
			fmt.Println("ошибка при сохранении шрифта: ", err)
			return
		}
		fmt.Println("шрифт сохранен: ", fileName)

		err = file.Close()
		if err != nil {
			return
		}
	}
	fmt.Println("загрузка шрифтов завершена")
}

func findFontUrls(def io.ReadCloser) []string {
	// Паттерн регулярного выражения для поиска ссылок на шрифты
	regexPattern := `url\s*\(\s*['"]?([^'"\)]+\.(ttf|woff2))['"]?\s*\)`

	// Компилируем регулярное выражение
	regex, err := regexp.Compile(regexPattern)
	if err != nil {
		fmt.Println("Ошибка при компиляции регулярного выражения:", err)
		return nil
	}
	data, err := io.ReadAll(def)
	if err != nil {
		log.Fatal(err)
	}

	// Находим все совпадения
	matches := regex.FindAllStringSubmatch(string(data), -1)

	// Извлекаем URL-адреса шрифтов из совпадений
	fontURLs := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			fontURLs = append(fontURLs, match[1])
		}
	}

	return fontURLs
}

func findImageUrls(def io.ReadCloser) []string {
	// Паттерн регулярного выражения для поиска ссылок на картинки
	regexPattern := `url\s*\(\s*['"]?([^'"\)]+\.(svg|png|jpg|jpeg))['"]?\s*\)`

	// Компилируем регулярное выражение
	regex, err := regexp.Compile(regexPattern)
	if err != nil {
		fmt.Println("Ошибка при компиляции регулярного выражения:", err)
		return nil
	}
	data, err := io.ReadAll(def)
	if err != nil {
		log.Fatal(err)
	}

	// Находим все совпадения
	matches := regex.FindAllStringSubmatch(string(data), -1)

	// Извлекаем URL-адреса шрифтов из совпадений
	imageUrls := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			imageUrls = append(imageUrls, match[1])
		}
	}

	return imageUrls
}

func downloadScripts(doc *goquery.Document, websiteUrl string) {
	err := os.Mkdir("script", os.ModePerm)
	if err != nil && !os.IsExist(err) {
		fmt.Println("ошибка при создании папки: ", err)
	}
	doc.Find("script").Each(func(index int, s *goquery.Selection) {
		href, exists := s.Attr("src")
		if exists {
			if strings.HasPrefix(href, "//") {
				href = "https:" + href
			} else if strings.HasPrefix(href, "/") {
				domain, _ := getDomain(websiteUrl)
				href = "https://" + domain + href
			} else if strings.Contains(href[0:4], "js/") || strings.Contains(href[0:8], "script/") || strings.Contains(href[0:9], "scripts/") {
				domain, _ := getDomain(websiteUrl)
				href = "https://" + domain + "/" + href
			}
		} else {
			return
		}

		resp, err := http.Get(href)
		if err != nil {
			fmt.Println("ошибка при загрузке скриптов: ", err, href)
			return
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {

			}
		}(resp.Body)

		fileName := "script/" + path.Base(href)

		file, err := os.Create(fileName)
		if err != nil {
			fmt.Println("ошибка при создании скрипта: ", err)
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {

			}
		}(file)
		_, err = io.Copy(file, resp.Body)
		if err != nil {
			fmt.Println("ошибка при сохранении скрипта: ", err)
			return
		}
		fmt.Println("скрипт сохранен:", fileName)
		s.SetAttr("src", fileName)
	})

	htmlString, err := doc.Html()
	if err != nil {
		fmt.Println("ошибка при генерации html: ", err)
		return
	}
	err = saveHTMLtoFile(htmlString)
}

func downloadImages(doc *goquery.Document, websiteUrl string) {
	err := os.Mkdir("images", os.ModePerm)
	if err != nil && !os.IsExist(err) {
		fmt.Println("ошибка при создании папки: ", err)
	}
	doc.Find("img").Each(func(index int, s *goquery.Selection) {
		href, exists := s.Attr("src")
		if exists {
			if strings.HasPrefix(href, "//") {
				href = "https:" + href
			} else if strings.HasPrefix(href, "/") {
				domain, _ := getDomain(websiteUrl)
				href = "https://" + domain + href

			} else if strings.HasPrefix(href, "../") || strings.HasPrefix(href, "./") {
				domain, _ := getDomain(websiteUrl)
				href = "https://" + domain + "/img/" + path.Base(href)

			} else if strings.Contains(href[0:4], "img/") || strings.Contains(href[0:8], "images/") || strings.Contains(href[0:7], "image/") {
				domain, _ := getDomain(websiteUrl)
				href = "https://" + domain + "/" + href

			}
		} else {
			return
		}

		resp, err := http.Get(href)
		if err != nil {
			fmt.Println("ошибка при загрузке фотографии: ", err, href)
			return
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {

			}
		}(resp.Body)

		fileName := "images/" + path.Base(href)

		file, err := os.Create(fileName)
		if err != nil {
			fmt.Println("ошибка при создании фотографии: ", err)
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {

			}
		}(file)

		_, err = io.Copy(file, resp.Body)
		if err != nil {
			fmt.Println("ошибка при сохранении фотографии: ", err)
			return
		}

		fmt.Println("фотография сохранена:", fileName)

		s.SetAttr("src", fileName)
	})

	htmlString, err := doc.Html()
	if err != nil {
		fmt.Println("ошибка при генерации html: ", err)
		return
	}
	err = saveHTMLtoFile(htmlString)
}

func downloadImagesFromStyles(fromFileName string, websiteUrl string) {
	filefrom, err := os.Open(fromFileName)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(filefrom)

	all := io.ReadCloser(filefrom)
	if err != nil {
		return
	}

	imageUrls := findImageUrls(all)

	for _, imageUrl := range imageUrls {

		if strings.HasPrefix(imageUrl, "//") {
			imageUrl = "https:" + imageUrl
		} else if strings.HasPrefix(imageUrl, "/") {
			imageUrl = websiteUrl + imageUrl
		} else if strings.HasPrefix(imageUrl, "../") || strings.HasPrefix(imageUrl, "./") {
			domain, _ := getDomain(websiteUrl)
			imageUrl = "https://" + domain + strings.Replace(imageUrl, ".", "", 2)
		}

		fmt.Println(imageUrl)
		resp, err := http.Get(imageUrl)
		if err != nil {
			fmt.Println("ошибка при загрузке картинки из стилей: ", err)
			return
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				panic("panic")
			}
		}(resp.Body)

		fileName := "images/" + path.Base(imageUrl)

		file, err := os.Create(fileName)
		if err != nil {
			fmt.Println("491 ошибка при создании файла: ", err)
		}

		defer func(file *os.File) {
			err := file.Close()
			if err != nil {

			}
		}(file)

		_, err = io.Copy(file, resp.Body)
		if err != nil {
			fmt.Println("ошибка при сохранении картинки из стиля: ", err)
			return
		}
		fmt.Println("картинка сохранен: ", fileName)
	}
	fmt.Println("загрузка картинки из стиля завершена")
}

func refactorStyles() {
	regex := regexp.MustCompile(`url\s*\(\s*['"]?([^'"\)]+\.(svg|png|jpg|jpeg|ttf|woff2))['"]?\s*\)`)
	styleFolder := "styles"

	err := filepath.Walk(styleFolder, func(pathToFile string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("ошибка доступа к файлу %v", err)
			return nil
		}

		if !info.IsDir() {
			file, err := os.Open(pathToFile)
			if err != nil {
				log.Printf("ошибка открытия файла: %v\n", err)
				return nil
			}
			defer func(file *os.File) {
				err := file.Close()
				if err != nil {
					return
				}
			}(file)

			input, err := io.ReadAll(file)
			if err != nil {
				log.Fatalln(err)
			}
			lines := strings.Split(string(input), "\n")
			for _, line := range lines {
				if regex.MatchString(line) {
					urls := regex.FindAllStringSubmatch(line, -1)
					urlsArray := make([]string, 0, len(urls))
					for _, match := range urls {
						if len(match) > 1 {
							urlsArray = append(urlsArray, match[1])
						}
					}
					newUrl, _ := url.Parse(urlsArray[0])
					if strings.Contains(line, "background") || strings.Contains(line, "background-image") {
						newStr := regex.ReplaceAllString(line, "url(../images/"+path.Base(newUrl.Path)+")")
						line = newStr
					} else if strings.Contains(line, "src") {
						newStr := regex.ReplaceAllString(line, "url(../fonts/"+path.Base(newUrl.Path)+")")
						line = newStr
					}
					line = line
				}

			}
			output := strings.Join(lines, "\n")
			err = os.WriteFile("myfile.css", []byte(output), 0644)
			if err != nil {
				log.Fatalln(err)
			}

			fmt.Println("файл закончен")
		}
		return nil
	})

	if err != nil {
		log.Fatalf("ошибка при обходе папки: %v", err)
	}

	fmt.Println("обход завершен")
}
