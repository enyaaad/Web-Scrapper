package refactor

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"Web-Scrapper/internal/utils"
)

func CleanTrashScripts(doc *goquery.Document) error {
	trashKeywords := []string{
		"https://mc.yandex.ru/metrika/tag.js",
		"$jsonData = {",
		"adpushup",
		"googletag",
		"gtm",
		"gtag",
		"mgid",
		"acrum",
	}

	doc.Find("script").Each(func(index int, s *goquery.Selection) {
		scriptText := s.Text()
		for _, keyword := range trashKeywords {
			if strings.Contains(scriptText, keyword) {
				s.Remove()
				return
			}
		}
	})

	htmlString, err := doc.Html()
	if err != nil {
		return fmt.Errorf("ошибка при генерации html: %w", err)
	}
	return utils.SaveHTMLtoFile(htmlString)
}

func RefactorImages(doc *goquery.Document) error {
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
		return fmt.Errorf("ошибка при генерации html: %w", err)
	}
	return utils.SaveHTMLtoFile(htmlString)
}

func RefactorAutoDomain(doc *goquery.Document) error {
	const autoDomainURL = "https://<?= $domain ?>/click.php?lp=1&place=comebacker<?= $uclick ?>"
	var autoDomain string
	
	fmt.Println("домен для автодомена/ничего не вводите")
	_, err := fmt.Scanf("%s\n", &autoDomain)
	if err != nil {
		return nil
	}

	phpCode := fmt.Sprintf("<?php\n$domain = (isset($_GET['domain'])) ? $_GET['domain'] : '%s';\n$uclick = (isset($_GET['uclick'])) ? '&uclick=' . $_GET['uclick'] : '';\n?>", autoDomain)
	
	err = utils.InsertStringToFile(utils.IndexFileName, phpCode, 0)
	if err != nil {
		return err
	}

	doc.Find("a").Each(func(index int, s *goquery.Selection) {
		s.SetAttr("href", autoDomainURL)
	})

	return nil
}

func RefactorGTMAcrum(doc *goquery.Document) error {
	var gtmCode, gtmName string

	fmt.Println("gtm crypto/adult/nutra или оставьте пустым")
	_, err := fmt.Scanf("%s\n", &gtmName)
	if err != nil {
		return nil
	}

	switch gtmName {
	case "crypto":
		gtmCode = "CRYPTOCODE"
	case "nutra":
		gtmCode = "NUTRACODE"
	case "adult":
		gtmCode = "ADULTCODE"
	default:
		return nil
	}

	if gtmCode == "" {
		return nil
	}

	gtmHead := fmt.Sprintf(`<script>var acrum_extra = {land: "", lang: "", funnel: "lp", offer: "", project: "%s", comebacker: true}
</script>
 <script>
    (function (w, d, s, l, i) {
        w[l] = w[l] || [];
        w[l].push({
            'gtm.start':
                new Date().getTime(), event: 'gtm.js'
        });
        var f = d.getElementsByTagName(s)[0],
            j = d.createElement(s), dl = l != 'dataLayer' ? '&l=' + l : '';
        j.async = true;
        j.src =
            'https://www.googletagmanager.com/gtm.js?id=' + i + dl;
        f.parentNode.insertBefore(j, f);
    })(window, document, 'script', 'dataLayer', '%s');
</script>`, gtmName, gtmCode)

	gtmBody := fmt.Sprintf(`<noscript>
    <iframe src="https://www.googletagmanager.com/ns.html?id=%s"
            height="0" width="0" style="display:none;visibility:hidden"></iframe>
</noscript>`, gtmCode)

	doc.Find("head").AppendHtml(gtmHead)
	doc.Find("body").PrependHtml(gtmBody)

	htmlString, err := doc.Html()
	if err != nil {
		return fmt.Errorf("ошибка при генерации html: %w", err)
	}
	return utils.SaveHTMLtoFile(htmlString)
}
