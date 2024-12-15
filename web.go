package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func htmlCek(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("HTTP isteği hazırlanamadı: %v", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	client := &http.Client{}
	yanit, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP isteği başarısız: %v", err)
	}
	return yanit, nil
}

func veriCek(url string, baslikSecici func(*goquery.Document) string, ozelCek func(*goquery.Document) (string, []string)) (string, string, []string, error) {
	yanit, err := htmlCek(url)
	if err != nil {
		return "", "", nil, err
	}
	defer yanit.Body.Close()

	govde, err := io.ReadAll(yanit.Body)
	if err != nil {
		return "", "", nil, fmt.Errorf("Yanıt gövdesi okunamadı: %v", err)
	}
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(govde)))
	if err != nil {
		return "", "", nil, fmt.Errorf("HTML parse hatası: %v", err)
	}

	baslik := baslikSecici(doc)
	aciklama, tarihler := ozelCek(doc)

	return baslik, aciklama, tarihler, nil
}

func dosyayaKaydet(dosyaAdi, baslik, aciklama string, tarihler []string) error {
	dosya, err := os.Create(dosyaAdi)
	if err != nil {
		return fmt.Errorf("Dosya oluşturulamadı: %v", err)
	}
	defer dosya.Close()

	tarihIcerik := strings.Join(tarihler, "\n")
	icerik := fmt.Sprintf("Site Başlığı: %s\nAçıklama: %s\nTarih:\n%s\n", baslik, aciklama, tarihIcerik)
	_, err = dosya.WriteString(icerik)
	if err != nil {
		return fmt.Errorf("Dosyaya yazılamadı: %v", err)
	}

	fmt.Printf("%s dosyasına başarıyla kaydedildi.\n", dosyaAdi)
	return nil
}

func menuGoster() {
	fmt.Println("\n--- Web Scraper Menü ---")
	fmt.Println("1. The Hacker News'ten veri çek")
	fmt.Println("2. VABS Haber Blog'dan veri çek")
	fmt.Println("3. Kayawraps Blog'dan veri çek")
	fmt.Println("4. Çıkış")
	fmt.Print("Seçiminizi yapınız: ")
}

func main() {
	for {
		menuGoster()

		var secim int
		fmt.Scanln(&secim)

		switch secim {
		case 1:
			fmt.Println("The Hacker News'ten veri çekiliyor...")
			baslik, aciklama, tarihler, err := veriCek(
				"https://thehackernews.com/",
				func(doc *goquery.Document) string {
					return strings.TrimSpace(doc.Find("h2.home-title").Text())
				},
				func(doc *goquery.Document) (string, []string) {
					aciklama := doc.Find("meta[name='twitter:title']").AttrOr("content", "Açıklama bulunamadı")
					var tarihler []string
					doc.Find("span.h-datetime").Each(func(i int, s *goquery.Selection) {
						tarih := strings.TrimSpace(s.Text())
						if tarih != "" {
							tarihler = append(tarihler, tarih)
						}
					})
					return aciklama, tarihler
				},
			)
			if err != nil {
				log.Printf("Hata (The Hacker News): %v\n", err)
				continue
			}
			err = dosyayaKaydet("thehackernews.txt", baslik, aciklama, tarihler)
			if err != nil {
				log.Printf("Dosya yazma hatası (The Hacker News): %v\n", err)
			}

		case 2:
			fmt.Println("VABS Haber Blog'dan veri çekiliyor...")
			baslik, aciklama, tarihler, err := veriCek(
				"https://www.vabs.com/haber-blog/",
				func(doc *goquery.Document) string {
					return strings.TrimSpace(doc.Find("h3").Text())
				},
				func(doc *goquery.Document) (string, []string) {
					aciklama := doc.Find("meta[property='og:url']").AttrOr("content", "Açıklama bulunamadı")
					var tarihler []string
					doc.Find("span").Each(func(i int, s *goquery.Selection) {
						tarih := strings.TrimSpace(s.Text())
						if regexp.MustCompile(`\d{2}\.\d{2}\.\d{4}`).MatchString(tarih) {
							tarihler = append(tarihler, tarih)
						}
					})
					return aciklama, tarihler
				},
			)
			if err != nil {
				log.Printf("Hata (VABS Haber Blog): %v\n", err)
				continue
			}
			err = dosyayaKaydet("vabs.txt", baslik, aciklama, tarihler)
			if err != nil {
				log.Printf("Dosya yazma hatası (VABS Haber Blog): %v\n", err)
			}

		case 3:
			fmt.Println("Kayawraps Blog'dan veri çekiliyor...")
			baslik, aciklama, tarihler, err := veriCek(
				"https://kayawraps.com/blog",
				func(doc *goquery.Document) string {
					return strings.TrimSpace(doc.Find("a[rel='bookmark']").Text())
				},
				func(doc *goquery.Document) (string, []string) {
					aciklama := doc.Find("meta[name='description']").AttrOr("content", "Açıklama bulunamadı")
					var tarihler []string
					doc.Find("span.date_in_number").Each(func(i int, s *goquery.Selection) {
						tarih := strings.TrimSpace(s.Text())
						if tarih != "" {
							tarihler = append(tarihler, tarih)
						}
					})
					return aciklama, tarihler
				},
			)
			if err != nil {
				log.Printf("Hata (Kayawraps Blog): %v\n", err)
				continue
			}
			err = dosyayaKaydet("kayawraps.txt", baslik, aciklama, tarihler)
			if err != nil {
				log.Printf("Dosya yazma hatası (Kayawraps Blog): %v\n", err)
			}

		case 4:
			fmt.Println("Programdan çıkılıyor...")
			return
		default:
			fmt.Println("Geçersiz seçim, lütfen tekrar deneyin.")
		}
	}
}
