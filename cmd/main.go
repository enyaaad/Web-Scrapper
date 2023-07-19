package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
)

type PurifyResponse struct {
	Purified struct {
		Content string `json:"content"`
		Length  int    `json:"length"`
	} `json:"purified"`
}

func main() {
	const webpageUrl = "https://www.tvpoolonline.com/content/1856770"
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

	downloadIndex(doc)
	downloadImages(doc, webpageUrl)
	downloadStylesNFonts(doc, webpageUrl)
	downloadScripts(doc, webpageUrl)
	refactorStyles()
	refactorIndexImages(doc)
	refactorIndexAutoDomain(doc)
	refactorIndexTrashScriptCleaning(doc)
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
				pathto, _ := url.Parse(websiteUrl)
				href = "https://" + domain + pathto.Path + href
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
	var wg sync.WaitGroup

	for _, fontURL := range fontURLs {
		wg.Add(1)
		fontURLTemp := fontURL
		go func(url string) {

			defer wg.Done()

			uploadFontUrl(fontURLTemp, websiteUrl)
		}(fontURL)
	}
	wg.Wait()
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
func uploadFontUrl(fontURL string, websiteUrl string) {
	if strings.HasPrefix(fontURL, "//") {
		fontURL = "https:" + fontURL
	} else if strings.HasPrefix(fontURL, "/") {
		fontURL = websiteUrl + fontURL
	} else if strings.HasPrefix(fontURL, "../") || strings.HasPrefix(fontURL, "./") {
		domain, _ := getDomain(websiteUrl)
		fontURL = "https://" + domain + strings.Replace(fontURL, ".", "", 2)
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
				pathto, _ := url.Parse(websiteUrl)
				href = "https://" + domain + pathto.Path + href
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

		if idx := strings.IndexByte(fileName, '?'); idx >= 0 {
			fileName = fileName[:idx]
		} else {
			fileName = "script/" + path.Base(href)
		}

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
func downloadImages(doc *goquery.Document, websiteUrl string) {
	err := os.Mkdir("images", os.ModePerm)
	if err != nil && !os.IsExist(err) {
		fmt.Println("ошибка при создании папки: ", err)
	}
	var imageUrls []string

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
				pathto, _ := url.Parse(websiteUrl)
				href = "https://" + domain + pathto.Path + href
			}
		} else {
			return
		}
		fileName := "images/" + path.Base(href)

		if strings.Contains(href, "data:mage/svg+xml;base64") {
			return
		} else {
			imageUrls = append(imageUrls, href)
		}
		s.SetAttr("src", fileName)
	})

	runtime.GOMAXPROCS(4)

	var wg sync.WaitGroup

	for _, imageUrl := range imageUrls {
		wg.Add(1)
		imgURLTemp := imageUrl
		go func(url string) {

			defer wg.Done()
			uploadIndexImages(imgURLTemp, "images/"+path.Base(imgURLTemp))

		}(imageUrl)
	}
	wg.Wait()

	htmlString, err := doc.Html()
	if err != nil {
		fmt.Println("ошибка при генерации html: ", err)
		return
	}
	err = saveHTMLtoFile(htmlString)
}
func uploadIndexImages(imageUrl string, fileName string) {

	resp, err := http.Get(imageUrl)
	if err != nil {
		fmt.Println("ошибка при загрузке фотографии: ", err, imageUrl)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

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
	var wg sync.WaitGroup

	for _, imageUrl := range imageUrls {
		wg.Add(1)
		imgURLTemp := imageUrl
		go func(url string) {

			defer wg.Done()

			uploadStyleImage(imgURLTemp, websiteUrl)
		}(imageUrl)
	}
	wg.Wait()
	fmt.Println("загрузка картинки из стиля завершена")
}
func uploadStyleImage(imageUrl string, websiteUrl string) {

	if strings.HasPrefix(imageUrl, "//") {
		imageUrl = "https:" + imageUrl
	} else if strings.HasPrefix(imageUrl, "/") {
		imageUrl = websiteUrl + imageUrl
	} else if strings.HasPrefix(imageUrl, "../") || strings.HasPrefix(imageUrl, "./") {
		domain, _ := getDomain(websiteUrl)
		pathto, _ := url.Parse(websiteUrl)
		imageUrl = "https://" + domain + pathto.Path + strings.Replace(imageUrl, ".", "", 2)[1:]
	}

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
			for i, line := range lines {
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
						line = regex.ReplaceAllString(line, "url(../images/"+path.Base(newUrl.Path)+")")

					} else if strings.Contains(line, "src") {
						line = regex.ReplaceAllString(line, "url(../fonts/"+path.Base(newUrl.Path)+")")
					}
				}
				lines[i] = line
			}
			output := strings.Join(lines, "\n")
			err = os.WriteFile(pathToFile, []byte(output), 0644)
			if err != nil {
				log.Fatalln(err)
			}
			purifyStyle(pathToFile)
			//to uncomment in future
			fmt.Println("файл закончен")
		}
		return nil
	})

	if err != nil {
		log.Fatalf("ошибка при обходе папки: %v", err)
	}

	fmt.Println("обход завершен")
}
func purifyStyle(pathToCss string) {
	const purifyUrl = "https://purifycss.online"
	resource := "/api/purify/"

	indexFile, err := os.Open("index.html")
	if err != nil {
		fmt.Println("ошибка при открытии index.html: ", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(indexFile)

	cssFile, err := os.Open(pathToCss)
	if err != nil {
		log.Printf("ошибка открытия файла: %v\n", err)
		return
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			return
		}
	}(cssFile)

	bytedCSS, err := io.ReadAll(cssFile)
	if err != nil {
		log.Fatalln(err)
	}

	bytedIndex, err := io.ReadAll(indexFile)
	if err != nil {
		log.Fatalln(err)
	}
	data := url.Values{}

	data.Set("cssCode", string(bytedCSS))
	data.Set("htmlCode", string(bytedIndex))
	u, _ := url.ParseRequestURI(purifyUrl)
	u.Path = resource
	urlStr := u.String()

	client := &http.Client{}
	r, _ := http.NewRequest(http.MethodPost, urlStr, strings.NewReader(data.Encode()))
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, _ := client.Do(r)
	if err != nil {
		panic(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	body, _ := io.ReadAll(resp.Body)

	cssContent := PurifyResponse{}

	err = json.Unmarshal(body, &cssContent)

	if err != nil {
		log.Printf("Reading body failed: %s", err)
		return
	}
	err = os.WriteFile(pathToCss, []byte(cssContent.Purified.Content), 0644)
	if err != nil {
		log.Fatalln(err)
	}

	return
}

func refactorIndexImages(doc *goquery.Document) {
	doc.Find("img").Each(func(index int, s *goquery.Selection) {
		s.SetAttr("loading", "lazy")
		s.SetAttr("alt", "")
		s.RemoveAttr("srcset")

	})
	doc.Find("script").Each(func(index int, s *goquery.Selection) {
		s.RemoveAttr("integrity")
		s.RemoveAttr("crossorigin")
	})
	htmlString, err := doc.Html()
	if err != nil {
		fmt.Println("ошибка при генерации html: ", err)
		return
	}
	err = saveHTMLtoFile(htmlString)
}
func refactorIndexAutoDomain(doc *goquery.Document) {
	const autoDomainUrl = "https://<?= $domain ?>/click.php?lp=1&place=comebacker<?= $uclick ?>"
	var autoDomain string
	fmt.Println("домен для автодомена/ничего не вводите")
	_, err := fmt.Scanf("%s\n", &autoDomain)
	if err != nil {
		return
	}
	switch autoDomain {
	default:
		doc.Find("a").Each(func(index int, s *goquery.Selection) {
			s.SetAttr("href", autoDomainUrl)
		})
		if err := utilsInsertStringToFile("index.html", "<?php\n$domain = (isset($_GET['domain'])) ? $_GET['domain'] : '"+autoDomain+"';\n$uclick = (isset($_GET['uclick'])) ? '&uclick=' . $_GET['uclick'] : '';\n?>", 0); err != nil {
			log.Fatal(err)
		}
	case "":
		return
	}
}
func refactorIndexTrashScriptCleaning(doc *goquery.Document) {
	doc.Find("script").Each(func(index int, s *goquery.Selection) {
		if strings.Contains(s.Text(), "https://mc.yandex.ru/metrika/tag.js") {
			s.Remove()
		} else if strings.Contains(s.Text(), "$jsonData = {") {
			s.Remove()
		} else if strings.Contains(s.Text(), "adpushup") {
			s.Remove()
		} else if strings.Contains(s.Text(), "googletag") || strings.Contains(s.Text(), "gtm") || strings.Contains(s.Text(), "gtag") {
			s.Remove()
		} else if strings.Contains(s.Text(), "mgid") {
			s.Remove()
		} else if strings.Contains(s.Text(), "acrum") {
			s.Remove()
		} else {
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
func refactorIndexDates(doc *goquery.Document) {
	//TODO algh search dates with regexp(xz) and
}
func refactorIndexGtmAcrum(doc *goquery.Document) {
	//TODO gtmAcrum
}
func refactorIndexComebacker(doc *goquery.Document) {
	//TODO comebacker
}

func utilsFileToLines(filePath string) ([]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			return
		}
	}(f)
	return utilsLinesFromReader(f)
}
func utilsLinesFromReader(r io.Reader) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}
func utilsInsertStringToFile(path, str string, index int) error {
	lines, err := utilsFileToLines(path)
	if err != nil {
		return err
	}

	fileContent := ""
	for i, line := range lines {
		if i == index {
			fileContent += str
		}
		fileContent += line
		fileContent += "\n"
	}

	return os.WriteFile(path, []byte(fileContent), 0644)
}
