package utils

import (
	"fmt"
	"net/url"
	"path"
	"strings"
)

func GetDomain(websiteURL string) (string, error) {
	parsedURL, err := url.Parse(websiteURL)
	if err != nil {
		return "", fmt.Errorf("ошибка при парсе URL: %w", err)
	}
	return parsedURL.Hostname(), nil
}

func NormalizeStyleURL(href, websiteURL string) (string, error) {
	if strings.HasPrefix(href, "//") {
		return "https:" + href, nil
	}
	if strings.HasPrefix(href, "/") {
		domain, err := GetDomain(websiteURL)
		if err != nil {
			return "", err
		}
		return "https://" + domain + href, nil
	}
	if strings.HasPrefix(href, "./") {
		domain, err := GetDomain(websiteURL)
		if err != nil {
			return "", err
		}
		parsedURL, err := url.Parse(websiteURL)
		if err != nil {
			return "", err
		}
		return "https://" + domain + parsedURL.Path + strings.Replace(href, ".", "", 1)[1:], nil
	}
	if strings.Contains(href, "css/") || strings.Contains(href, "styles/") || strings.Contains(href, "style/") {
		domain, err := GetDomain(websiteURL)
		if err != nil {
			return "", err
		}
		parsedURL, err := url.Parse(websiteURL)
		if err != nil {
			return "", err
		}
		return "https://" + domain + parsedURL.Path + href, nil
	}
	if strings.Contains(href, "fonts.googleapis.com") {
		return "https://fonts.googleapis.com/" + path.Base(href), nil
	}
	
	domain, err := GetDomain(websiteURL)
	if err != nil {
		return "", err
	}
	parsedURL, err := url.Parse(websiteURL)
	if err != nil {
		return "", err
	}
	return "https://" + domain + parsedURL.Path + href, nil
}

func NormalizeScriptURL(href, websiteURL string) (string, error) {
	if strings.HasPrefix(href, "//") {
		return "https:" + href, nil
	}
	if strings.HasPrefix(href, "/") {
		domain, err := GetDomain(websiteURL)
		if err != nil {
			return "", err
		}
		return "https://" + domain + href, nil
	}
	if strings.Contains(href, "js/") || strings.Contains(href, "script/") || strings.Contains(href, "scripts/") {
		domain, err := GetDomain(websiteURL)
		if err != nil {
			return "", err
		}
		parsedURL, err := url.Parse(websiteURL)
		if err != nil {
			return "", err
		}
		return "https://" + domain + parsedURL.Path + href, nil
	}
	
	domain, err := GetDomain(websiteURL)
	if err != nil {
		return "", err
	}
	parsedURL, err := url.Parse(websiteURL)
	if err != nil {
		return "", err
	}
	return "https://" + domain + parsedURL.Path + href, nil
}

func NormalizeImageURL(href, websiteURL string) (string, error) {
	if strings.HasPrefix(href, "//") {
		return "https:" + href, nil
	}
	if strings.HasPrefix(href, "/") {
		domain, err := GetDomain(websiteURL)
		if err != nil {
			return "", err
		}
		return "https://" + domain + href, nil
	}
	if strings.HasPrefix(href, "../") || strings.HasPrefix(href, "./") {
		domain, err := GetDomain(websiteURL)
		if err != nil {
			return "", err
		}
		return "https://" + domain + "/img/" + path.Base(href), nil
	}
	if strings.Contains(href, "img/") || strings.Contains(href, "images/") || strings.Contains(href, "image/") {
		domain, err := GetDomain(websiteURL)
		if err != nil {
			return "", err
		}
		parsedURL, err := url.Parse(websiteURL)
		if err != nil {
			return "", err
		}
		return "https://" + domain + parsedURL.Path + href, nil
	}
	
	domain, err := GetDomain(websiteURL)
	if err != nil {
		return "", err
	}
	parsedURL, err := url.Parse(websiteURL)
	if err != nil {
		return "", err
	}
	return "https://" + domain + parsedURL.Path + href, nil
}

func NormalizeFontURL(fontURL, websiteURL string) (string, error) {
	if strings.HasPrefix(fontURL, "//") {
		return "https:" + fontURL, nil
	}
	if strings.HasPrefix(fontURL, "/") {
		return websiteURL + fontURL, nil
	}
	if strings.HasPrefix(fontURL, "../") || strings.HasPrefix(fontURL, "./") {
		domain, err := GetDomain(websiteURL)
		if err != nil {
			return "", err
		}
		return "https://" + domain + strings.Replace(fontURL, ".", "", 2), nil
	}
	
	domain, err := GetDomain(websiteURL)
	if err != nil {
		return "", err
	}
	parsedURL, err := url.Parse(websiteURL)
	if err != nil {
		return "", err
	}
	return "https://" + domain + parsedURL.Path + fontURL, nil
}

func NormalizeStyleImageURL(imageURL, websiteURL string) (string, error) {
	if strings.HasPrefix(imageURL, "//") {
		return "https:" + imageURL, nil
	}
	if strings.HasPrefix(imageURL, "/") {
		return websiteURL + imageURL, nil
	}
	if strings.HasPrefix(imageURL, "../") || strings.HasPrefix(imageURL, "./") {
		domain, err := GetDomain(websiteURL)
		if err != nil {
			return "", err
		}
		parsedURL, err := url.Parse(websiteURL)
		if err != nil {
			return "", err
		}
		return "https://" + domain + parsedURL.Path + strings.Replace(imageURL, ".", "", 2)[1:], nil
	}
	
	domain, err := GetDomain(websiteURL)
	if err != nil {
		return "", err
	}
	parsedURL, err := url.Parse(websiteURL)
	if err != nil {
		return "", err
	}
	return "https://" + domain + parsedURL.Path + imageURL, nil
}
