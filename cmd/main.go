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
	//const webpageUrl = "https://www.tvpoolonline.com/content/1856770"
	//const webpageUrl = "https://quick-fit-tr.online-shop2023.com/?lttracking=6afa6af8d59715e0b35f83abbefc74a0&ltpostclick=1691291976&source=leadtrade&ltsource=46622&lthash=ElJXN&landing=0&st=0KLQuNC30LXRgNC90YvQtSDRgdC10YLQuA%3D%3D&offer_id=8212&s1=&s2=&s3=&s4="
	const webpageUrl = "https://doctorseu.com/ro2/pages/landing/diet/LP12/keto_probiotic_gt_tl/?clickid=c519escb71za8c43&uclick=scb71za8&uclickhash=scb71za8-scb71za8-gxvr-0-9lus-q5qdfe-q5qd0-72e1db&t_id=2&domain=doctorseu.com&bf_lander=to_showcase-30-to_offer&bf_offer=to_showcase-30-to_offer2&manager_id=37&campaign_id=113&lander_id=71&offer_id=1800&click_from=page&traffic_source=MGID&traffic_source_id=Bucuresti&exit=1$"

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

	refactorIndexTrashScriptCleaning(doc)
	downloadScripts(doc, webpageUrl)

	refactorStyles()
	refactorIndexImages(doc)
	refactorIndexGtmAcrum(doc)
	refactorIndexAutoDomain(doc)

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
			} else if strings.HasPrefix(href, "/") {
				domain, _ := getDomain(websiteUrl)
				href = "https://" + domain + href
			} else if strings.HasPrefix(href, "./") {
				domain, _ := getDomain(websiteUrl)
				pathto, _ := url.Parse(websiteUrl)
				href = "https://" + domain + pathto.Path + strings.Replace(href, ".", "", 1)[1:]
			} else if strings.Contains(href[0:], "css/") || strings.Contains(href[0:], "styles/") || strings.Contains(href[0:], "style/") {
				domain, _ := getDomain(websiteUrl)
				pathto, _ := url.Parse(websiteUrl)
				href = "https://" + domain + pathto.Path + href
			} else if strings.Contains(href, "fonts.googleapis.com") {
				href = "https://fonts.googleapis.com/" + path.Base(href)
			} else {
				domain, _ := getDomain(websiteUrl)
				pathto, _ := url.Parse(websiteUrl)
				href = "https://" + domain + pathto.Path + href
			}
		} else {
			return
		}

		fmt.Println(href)
		resp, httpGetErr := http.Get(href)
		if httpGetErr != nil {
			fmt.Println("ошибка при загрузке стиля: ", httpGetErr, href)
			return
		}

		defer func(Body io.ReadCloser) {
			bodyCloseErr := Body.Close()
			if bodyCloseErr != nil {

			}
		}(resp.Body)

		fileName := path.Base(href)

		if fileName[len(fileName)-3:] != "css" {
			fileName = "styles/font" + fmt.Sprintf("%d", index) + ".css"
		} else {
			fileName = "styles/" + fileName
		}

		file, createError := os.Create(fileName)
		if createError != nil {
			fmt.Println("ошибка при создании файла: ", createError)
		}

		_, copyError := io.Copy(file, resp.Body)
		if copyError != nil {
			log.Println("ошибка при открытии файла: ", copyError)
		}

		fmt.Println("стиль сохранен:", fileName)
		downloadFonts(fileName, websiteUrl)
		downloadImagesFromStyles(fileName, websiteUrl)
		s.SetAttr("href", fileName)

		copyError = file.Close()
		if copyError != nil {
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
	} else {
		domain, _ := getDomain(websiteUrl)
		pathto, _ := url.Parse(websiteUrl)
		fontURL = "https://" + domain + pathto.Path + fontURL
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
			} else if strings.Contains(href[0:], "js/") || strings.Contains(href[0:], "script/") || strings.Contains(href[0:], "scripts/") {
				domain, _ := getDomain(websiteUrl)
				pathto, _ := url.Parse(websiteUrl)
				href = "https://" + domain + pathto.Path + href
			} else {
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

		tempHref, is := s.Attr("data-src")
		if is {
			href = tempHref
			s.RemoveAttr("data-src")
		}

		if exists {
			if strings.HasPrefix(href, "//") {
				href = "https:" + href
			} else if strings.HasPrefix(href, "/") {
				domain, _ := getDomain(websiteUrl)
				href = "https://" + domain + href

			} else if strings.HasPrefix(href, "../") || strings.HasPrefix(href, "./") {
				domain, _ := getDomain(websiteUrl)
				href = "https://" + domain + "/img/" + path.Base(href)

			} else if strings.Contains(href[0:], "img/") || strings.Contains(href[0:], "images/") || strings.Contains(href[0:], "image/") {
				domain, _ := getDomain(websiteUrl)
				pathto, _ := url.Parse(websiteUrl)
				href = "https://" + domain + pathto.Path + href
			} else {
				domain, _ := getDomain(websiteUrl)
				pathto, _ := url.Parse(websiteUrl)
				href = "https://" + domain + pathto.Path + href
			}
		} else {
			return
		}

		fileName := "images/" + path.Base(href)

		if strings.Contains(href, "base64") {
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
	} else {
		domain, _ := getDomain(websiteUrl)
		pathto, _ := url.Parse(websiteUrl)
		imageUrl = "https://" + domain + pathto.Path + imageUrl
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
	doc.Find("source").Each(func(i int, selection *goquery.Selection) {
		selection.Remove()
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
	errorToPaste := utilsInsertStringToFile("index.html", "<?php\n$domain = (isset($_GET['domain'])) ? $_GET['domain'] : '"+autoDomain+"';\n$uclick = (isset($_GET['uclick'])) ? '&uclick=' . $_GET['uclick'] : '';\n?>", 0)
	if errorToPaste != nil {
		return
	}
	doc.Find("a").Each(func(index int, s *goquery.Selection) {
		s.SetAttr("href", autoDomainUrl)
	})
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
	var gtmCode, gtmName string

	fmt.Println("gtm crypto/adult/nutra или оставьте пустым")
	_, err := fmt.Scanf("%s\n", &gtmName)
	if err != nil {
		return
	}
	switch gtmName {
	default:
		gtmCode = ""
		return
	case "crypto":
		gtmCode = "CRYPTOCODE"
	case "nutra":
		gtmCode = "NUTRACODE"
	case "adult":
		gtmCode = "ADULTCODE"
	}
	if gtmCode != "" {
		var gtmHead string = "<script>var acrum_extra = {land: \"\", lang: \"\", funnel: \"lp\", offer: \"\", project: \"" + gtmName + "\", comebacker: true}\n</script>\n <script>\n    (function (w, d, s, l, i) {\n        w[l] = w[l] || [];\n        w[l].push({\n            'gtm.start':\n                new Date().getTime(), event: 'gtm.js'\n        });\n        var f = d.getElementsByTagName(s)[0],\n            j = d.createElement(s), dl = l != 'dataLayer' ? '&l=' + l : '';\n        j.async = true;\n        j.src =\n            'https://www.googletagmanager.com/gtm.js?id=' + i + dl;\n        f.parentNode.insertBefore(j, f);\n    })(window, document, 'script', 'dataLayer', '" + gtmCode + "');\n</script>"
		var gtmBody string = "<noscript>\n    <iframe src=\"https://www.googletagmanager.com/ns.html?id=" + gtmCode + "\"\n            height=\"0\" width=\"0\" style=\"display:none;visibility:hidden\"></iframe>\n</noscript>"
		doc.Find("head").AppendHtml(gtmHead)
		doc.Find("body").PrependHtml(gtmBody)

		htmlString, errror := doc.Html()
		if errror != nil {
			fmt.Println("ошибка при генерации html: ", err)
			return
		}
		errror = saveHTMLtoFile(htmlString)
	}
}
func refactorIndexComebacker(doc *goquery.Document) {
	const drTime = "months_localized = {\n    'ru': ['января', 'февраля', 'марта', 'апреля', 'мая', 'июня', 'июля', 'августа', 'сентября', 'октября', 'ноября', 'декабря'],\n    'fr': ['janvier', 'février', 'mars', 'avril', 'mai', 'juin', 'juillet', 'août', 'septembre', 'octobre', 'novembre', 'décembre'],\n    'bg': ['Януари', 'Февруари', 'Март', 'Април', 'Май', 'Юни', 'Юли', 'Август', 'Септември', 'Октомври', 'Ноември', 'Декември'],\n    'nl': ['januari', 'februari', 'maart', 'april', 'mei', 'juni', 'juli', 'augustus', 'september', 'oktober', 'november', 'december'],\n    'pt': ['Janeiro', 'Fevereiro', 'Março', 'Abril', 'Maio', 'Junho', 'Julho', 'Agosto', 'Setembro', 'Outubro', 'Novembro', 'Dezembro'],\n    'de': ['Januar', 'Februar', 'März', 'April', 'Mai', 'Juni', 'Juli', 'August', 'September', 'Oktober', 'November', 'Dezember'],\n    'tr': ['Ocak', 'Şubat', 'Mart', 'Nisan', 'Mayıs', 'Haziran', 'Temmuz', 'Ağustos', 'Eylül', 'Ekim', 'Kasım', 'Aralık'],\n    'it': ['Gennaio', 'Febbraio', 'Marzo', 'Aprile', 'Maggio', 'Giugno', 'Luglio', 'Agosto', 'Settembre', 'Ottobre', 'Novembre', 'Dicembre'],\n    'hu': ['Január', 'Február', 'Március', 'Április', 'Május', 'Június', 'Július', 'Augusztus', 'Szeptember', 'Október', 'November', 'December'],\n    'en': ['January', 'February', 'March', 'April', 'May', 'June', 'July', 'August', 'September', 'October', 'November', 'December'],\n    'id': ['Januari', 'Februari', 'Maret', 'April', 'Mei', 'Juni', 'Juli', 'Agustus', 'September', 'Oktober', 'November', 'Desember'],\n    'ms': ['Januari', 'Februari', 'Mac', 'April', 'Mei', 'Jun', 'Julai', 'Ogos', 'September', 'Oktober', 'November', 'Disember'],\n    'hi': ['जनवर', 'फरबर', 'मार्च', 'अप्रैल', 'मई', 'जून', 'जुलाई', 'अगस्त', 'सितम्बर', 'अक्टूबर', 'नवंबर', 'दिसंबर'],\n    'es': ['Enero', 'Febrero', 'Marzo', 'Abril', 'Mayo', 'Junio', 'Julio', 'Agosto', 'Septiembre', 'Octubre', 'Noviembre', 'Diciembre'],\n    'ro': ['Ianuarie', 'Februarie', 'Martie', 'Aprilie', 'Mai', 'Iunie', 'Iulie', 'August', 'Septembrie', 'Octombrie', 'Noiembrie', 'Decembrie'],\n    'pl': ['stycznia', 'lutego', 'marca', 'kwietnia', 'maja', 'czerwca', 'lipca', 'sierpnia', 'września', 'października', 'listopada', 'grudnia'],\n    'sr': ['Januar', 'Februar', 'Mart', 'April', 'Maj', 'Jun', 'Jul', 'Avgust', 'Septembar', 'Oktobar', 'Novembar', 'Decembar'],\n    'cs': ['ledna', 'února', 'března', 'dubna', 'května', 'června', 'července', 'srpna', 'září', 'října', 'listopadu', 'prosince'],\n    'sk': ['januára', 'februára', 'marca', 'apríla', 'mája', 'júna', 'júla', 'augusta', 'septembra', 'októbra', 'novembra', 'decembra'],\n    'el': ['Ιανουάριος', 'Φεβρουάριος', 'Μάρτιος', 'Απρίλιος', 'Μάιος', 'Ιούνιος', 'Ιούλιος', 'Αύγουστος', 'Σεπτέμβριος', 'Οκτώβριος', 'Νοέμβριος', 'Δεκέμβριος'],\n    'th': ['มกราคม', 'กุมภาพันธ์', 'มีนาคม', 'เมษายน', 'พฤษภาคม', 'มิถุนายน', 'กรกฎาคม', 'สิงหาคม', 'กันยายน', 'ตุลาคม', 'พฤศจิกายน', 'ธันวาคม'],\n    'vi': ['Tháng Một', 'Tháng Hai', 'Tháng Ba', 'Tháng Bốn', 'Tháng Năm', 'Tháng Sáu', 'Tháng Bảy', 'Tháng Tám'],\n    'fil': ['Enero', 'Pebrero', 'Marso', 'Abril', 'Mayo', 'Hunyo', 'Hulyo', 'Agosto', 'Setyembre', 'Oktubre', 'Nobyembre', 'Disyembre'],\n    'ar': ['يناير', 'فبراير', 'مارس', 'أبريل', 'مايو', 'يونيو', 'يوليو', 'أغسطس', 'سبتمبر', 'أكتوبر', 'نوفمبر', 'ديسمبر'],\n    'ur': ['جنوری', 'فروری', 'مارچ', 'اپریل', 'مئی', 'جون', 'جولائی', 'اگست', 'ستمبر', 'اکتوبر', 'نومبر', 'دسمبر'],\n    'nb': ['Januar', 'Februar', 'Mars ', 'April ', 'May ', 'Juni ', 'Juli ', 'August ', 'September ', 'Oktober ', 'November ', 'Desember '],\n    'nn': ['Januar', 'Februar', 'Mars ', 'April ', 'May ', 'Juni ', 'Juli ', 'August ', 'September ', 'Oktober ', 'November ', 'Desember '],\n    'no': ['Januar', 'Februar', 'Mars ', 'April ', 'May ', 'Juni ', 'Juli ', 'August ', 'September ', 'Oktober ', 'November ', 'Desember '],\n    'nb_NO': ['Januar', 'Februar', 'Mars ', 'April ', 'May ', 'Juni ', 'Juli ', 'August ', 'September ', 'Oktober ', 'November ', 'Desember '],\n    'km': ['មករា', 'កុម្ភៈ', 'មិនា', 'មេសា', 'ឧសភា', 'មិថុនា', 'កក្កដា', 'សីហា', 'កញ្ញា', 'តុលា', '‘វិច្ឆិកា', 'ធ្នូ'],\n    'zh': ['一月', '二月', '三月', '四月', '五月', '六月', '七月', '八月', '九月', '十月', '十一月', '十二月']\n};days_localized = {\n    'ru': ['воскресенье', 'понедельник', 'вторник', 'среда', 'четверг', 'пятница', 'суббота'],\n    'fr': ['Dimanche', 'Lundi', 'Mardi', 'Mercredi', 'Jeudi', 'Vendredi', 'Samedi'],\n    'bg': ['Неделя', 'Понеделник', 'Вторник', 'Сряда', 'Четвъртък', 'Петък', 'Събота'],\n    'nl': ['zondag', 'maandag', 'dinsdag', 'woensdag', 'donderdag', 'vrijdag', 'zaterdag'],\n    'pt': ['Domingo', 'Segunda Feira', 'Terça Feira', 'Quarta Feira', 'Quinta Feira', 'Sexta Feira', 'Sábado'],\n    'de': ['Sonntag', 'Montag', 'Dienstag', 'Mittwoch', 'Donnerstag', 'Freitag', 'Samstag'],\n    'tr': ['Pazar', 'Pazartesi', 'Salı', 'Çarşamba', 'Perşembe', 'Cuma', 'Cumartesi'],\n    'it': ['Domenica', 'Lunedì', 'Martedì', 'Mercoledì', 'Giovedì', 'Venerdì', 'Sabato'],\n    'hu': ['Vasárnap', 'Hétfő', 'Kedd', 'Szerda', 'Csütörtök', 'Péntek', 'Szombat'],\n    'en': ['Sunday', 'Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday'],\n    'hi': ['सोमवार', 'मंगलवार', 'बुधवार', 'गुरूवार', 'शुक्रवार', 'शनिवार', 'रविवार'],\n    'ms': ['Ahad', 'Isnin', 'Selasa', 'Rabu', 'Khamis', 'Jumaat', 'Sabtu'],\n    'id': ['Minggu', 'Senin', 'Selasa', 'Rabu', 'Kamis', 'Jumat', 'Sabtu'],\n    'es': ['Domingo', 'Lunes', 'Martes', 'Miércoles', 'Jueves', 'Viernes', 'Sábado'],\n    'ro': ['Duminică', 'Luni', 'Marţi', 'Miercuri', 'Joi', 'Vineri', 'Sâmbătă'],\n    'pl': ['niedziela', 'poniedziałek', 'wtorek', 'środa', 'czwartek', 'piątek', 'sobota'],\n    'sr': ['Nedelja', 'Ponedeljak', 'Utorak', 'Sreda', 'Četvrtak', 'Petak', 'Subota'],\n    'cs': ['neděle', 'pondělí', 'úterý', 'středa', 'čtvrtek', 'pátek', 'sobota'],\n    'sk': ['nedeľa', 'pondelok', 'utorok', 'streda', 'štvrtok', 'piatok', 'sobota'],\n    'el': ['Κυριακή', 'Δευτέρα', 'Τρίτη', 'Τετάρτη', 'Πέμπτη', 'Παρασκευή', 'Σάββατο'],\n    'th': ['วันอาทิตย์', 'วันจันทร์', 'วันอังคาร', 'วันพุธ', 'วันพฤหัสบดี', 'วันศุกร์', 'วันเสาร์'],\n    'vi': ['Chủ Nhật', 'Thứ Hai', 'Thứ Ba', 'Thứ Tư', 'Thứ Năm', 'Thứ Sáu', 'Thứ Bảy'],\n    'ar': ['الاحد', 'الاثنين', 'الثلاثاء', 'الاربعاء', 'الخميس', 'الجمعة', 'السبت'],\n    'fil': ['Linggo', 'Lunes', 'Martes', 'Miyerkoles', 'Huebes', 'Biyernes', 'Sabado'],\n    'ur': ['اتوار', 'پیر', 'منگل', 'بدھ', 'جمعرات', 'جمعہ', 'ہفتہ'],\n    'nb': ['Søndag', 'Mandag', 'Tirsdag', 'Onsdag', 'Torsdag', 'Friday', 'Lørdag'],\n    'nn': ['Søndag', 'Mandag', 'Tirsdag', 'Onsdag', 'Torsdag', 'Friday', 'Lørdag'],\n    'no': ['Søndag', 'Mandag', 'Tirsdag', 'Onsdag', 'Torsdag', 'Friday', 'Lørdag'],\n    'nb_NO': ['Søndag', 'Mandag', 'Tirsdag', 'Onsdag', 'Torsdag', 'Friday', 'Lørdag'],\n    'km': ['អាទិត្យ', 'ច័ន្ធ', 'អង្គារ៍', 'ពុធ', 'ព្រហស្បិ៍', 'សុក្រ', 'សៅរ៍'],\n    'zh': ['星期天', '星期一', '星期二', '星期三', '星期四', '星期五', '星期六']\n};function dtimes(d = 0) {\n    let lang_locale = 'ru';\n    let now = new Date();\n    now.setDate(now.getDate() + d);\n    document.write((now.getDate()) + \" \" + months_localized[lang_locale][now.getMonth()]);\n}function dtime_nums(d = 0) {\n    var now = new Date();\n    now.setDate(now.getDate() + d);\n    var dayNum = '';\n    if (now.getDate() < 10) dayNum = '0'\n    dayNum += now.getDate();\n    var monthNum = '';\n    if (now.getMonth() + 1 < 10) monthNum = '0';\n    monthNum += now.getMonth() + 1;\n    document.write(dayNum + \".\" + monthNum + \".\" + now.getFullYear());\n}"
	const comebackJs = "var comebacker=document.getElementById(\"comeback\");let comeback=()=>document.querySelector(\"#comeback\").style.display=\"block\";var stateObj={foo:\"bar\"},curURL=window.location.href,curTitle=document.title;history.pushState(stateObj,curTitle,curURL),window.onpopstate=function(e){history.pushState(stateObj,curTitle,curURL),comeback()},document.body.onmouseout=function(e){e.clientY<0&&comeback()},comebacker.onclick=function(e){\"comeback\"===e.target.id&&(document.querySelector(\"#comeback\").style.display=\"none\")};"
	const comebackCss = "#comeback,.footer_fixed{position:fixed;bottom:0}#comeback{display:none;top:0;right:0;left:0;z-index:1000;background:rgba(0,0,0,.75);overflow-y:scroll;-ms-overflow-style:none;overflow:-moz-scrollbars-none}#comeback::-webkit-scrollbar{width:0}#comeback .comeback_container{position:relative;background-color:#fff;padding:25px;margin-top:2%;text-align:center;border-radius:10px}#comeback .close{position:absolute;top:8px;right:8px;display:block;width:21px;height:21px;font-size:0;cursor:pointer}#comeback .close:after,#comeback .close:before{position:absolute;top:50%;left:50%;width:2px;height:25px;background-color:#bb1919;transform:rotate(45deg) translate(-50%,-50%);transform-origin:top left;content:''}#comeback .comeback_container .btn,#footer-href{position:relative;overflow:hidden;padding:10px;background-color:#d72222;border-radius:17px;color:#fff;text-decoration:none}#comeback .close:after{transform:rotate(-45deg) translate(-50%,-50%)}#comeback .close:hover{transform:scale(.9)}#comeback .comeback_info_text{margin-top:7px;font-size:12px}#comeback .mt-10{margin-top:10px}#comeback .curr_num{color:#bb1919}#comeback .comeback_wrapper{width:100%;max-width:650px;margin:0 auto}#comeback .comeback_container .btn{display:block;max-width:280px;box-shadow:0 3px 0 0 #910000;text-transform:uppercase;font-size:19px;font-weight:400;text-align:center;white-space:normal;vertical-align:middle;margin:0 auto;transition:.1s ease-in-out}#comeback .comeback_container .btn:hover{transform:translateY(3px);box-shadow:none}#comeback .comeback_container .btn:before,#footer-href::after{position:absolute;content:\"\";width:0;height:100%;top:0;right:0;z-index:-1;background-color:#c51919;border-radius:17px;transition:.3s}#comeback .comeback_container .btn:hover:before,#footer-href:hover::after{left:0;width:100%}#comeback p{font-size:22px;margin:0}@media (max-width:700px){#comeback .comeback_wrapper{width:100%;max-width:500px;margin:0 auto}#comeback .comeback_container{padding:10px}#comeback .comeback_box{margin-top:20px}}.footer_fixed{background:rgba(0,0,0,.68);display:flex;align-items:center;justify-content:center;flex-wrap:wrap;width:100%;padding:5px 0;text-align:center;z-index:999;font-size:22px}.footer_fixed .footer_fixed_text{color:#fff;margin:0 10px 0 0;font-family:'Roboto Condensed',sans-serif}.footer_fixed .other_text{color:#e1c231;padding:0 2px;font-weight:700}#footer-href{z-index:1;width:220px}@media (max-width:767px){.footer_fixed .footer_fixed_text{font-size:17px;margin:0 0 5px}#footer-href{padding:5px 10px}}"
	const countdownTimer = "if(document.getElementById(\"countdownTimer\")){let e=document.getElementById(\"countdownTimer\").getAttribute(\"data-minutes\"),t=document.getElementById(\"countdownTimer\").getAttribute(\"data-seconds\");function n(){-1==--t&&(t=59,e-=1),t<=9&&(t=\"0\"+t),time=(e<=9?\"0\"+e:e)+\":\"+t,document.getElementById(\"countdownTimer\").innerHTML=time,document.getElementById(\"countdownTimer2\").innerHTML=time,SD=window.setTimeout(\"countDown();\",1e3),\"00\"==e&&\"00\"==t&&(t=\"00\",window.clearTimeout(SD))}window.onload=n}"

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

	fmt.Println(path)
	return os.WriteFile(path, []byte(fileContent), 0644)
}
