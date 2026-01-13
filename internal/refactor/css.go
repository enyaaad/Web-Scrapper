package refactor

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"Web-Scrapper/internal/utils"
)

type PurifyResponse struct {
	Purified struct {
		Content string `json:"content"`
		Length  int    `json:"length"`
	} `json:"purified"`
}

func RefactorStyles() error {
	regex := regexp.MustCompile(`url\s*\(\s*['"]?([^'"\)]+\.(svg|png|jpg|jpeg|ttf|woff2))['"]?\s*\)`)
	styleFolder := "styles"

	err := filepath.Walk(styleFolder, func(pathToFile string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("ошибка доступа к файлу %v", err)
			return nil
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(pathToFile)
		if err != nil {
			log.Printf("ошибка открытия файла: %v\n", err)
			return nil
		}
		defer file.Close()

		input, err := io.ReadAll(file)
		if err != nil {
			return err
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
				if len(urlsArray) == 0 {
					continue
				}

				newURL, _ := url.Parse(urlsArray[0])
				if strings.Contains(line, "background") || strings.Contains(line, "background-image") {
					line = regex.ReplaceAllString(line, "url(../images/"+path.Base(newURL.Path)+")")
				} else if strings.Contains(line, "src") {
					line = regex.ReplaceAllString(line, "url(../fonts/"+path.Base(newURL.Path)+")")
				}
			}
			lines[i] = line
		}

		output := strings.Join(lines, "\n")
		err = os.WriteFile(pathToFile, []byte(output), 0644)
		if err != nil {
			return err
		}

		if err := PurifyStyle(pathToFile); err != nil {
			log.Printf("ошибка при очистке стиля: %v\n", err)
		}

		fmt.Println("файл закончен")
		return nil
	})

	if err != nil {
		return fmt.Errorf("ошибка при обходе папки: %w", err)
	}

	fmt.Println("обход завершен")
	return nil
}

func PurifyStyle(pathToCSS string) error {
	const purifyURL = "https://purifycss.online"
	resource := "/api/purify/"

	indexFile, err := os.Open(utils.IndexFileName)
	if err != nil {
		return fmt.Errorf("ошибка при открытии index.html: %w", err)
	}
	defer indexFile.Close()

	cssFile, err := os.Open(pathToCSS)
	if err != nil {
		return fmt.Errorf("ошибка открытия файла: %w", err)
	}
	defer cssFile.Close()

	bytedCSS, err := io.ReadAll(cssFile)
	if err != nil {
		return err
	}

	bytedIndex, err := io.ReadAll(indexFile)
	if err != nil {
		return err
	}

	data := url.Values{}
	data.Set("cssCode", string(bytedCSS))
	data.Set("htmlCode", string(bytedIndex))
	
	u, err := url.ParseRequestURI(purifyURL)
	if err != nil {
		return err
	}
	u.Path = resource
	urlStr := u.String()

	client := &http.Client{}
	r, err := http.NewRequest(http.MethodPost, urlStr, strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	cssContent := PurifyResponse{}
	err = json.Unmarshal(body, &cssContent)
	if err != nil {
		return fmt.Errorf("ошибка парсинга ответа: %w", err)
	}

	err = os.WriteFile(pathToCSS, []byte(cssContent.Purified.Content), 0644)
	if err != nil {
		return err
	}

	return nil
}
