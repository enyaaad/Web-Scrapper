package parser

import (
	"io"
	"log"
	"regexp"
)

func FindFontURLs(reader io.ReadCloser) ([]string, error) {
	regexPattern := `url\s*\(\s*['"]?([^'"\)]+\.(ttf|woff2))['"]?\s*\)`
	
	regex, err := regexp.Compile(regexPattern)
	if err != nil {
		return nil, err
	}
	
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	matches := regex.FindAllStringSubmatch(string(data), -1)

	fontURLs := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			fontURLs = append(fontURLs, match[1])
		}
	}

	return fontURLs, nil
}

func FindImageURLs(reader io.ReadCloser) ([]string, error) {
	regexPattern := `url\s*\(\s*['"]?([^'"\)]+\.(svg|png|jpg|jpeg))['"]?\s*\)`
	
	regex, err := regexp.Compile(regexPattern)
	if err != nil {
		return nil, err
	}
	
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	matches := regex.FindAllStringSubmatch(string(data), -1)

	imageURLs := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			imageURLs = append(imageURLs, match[1])
		}
	}

	return imageURLs, nil
}
