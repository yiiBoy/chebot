package repositories

import (
	"github.com/andboson/chebot/models"
	"github.com/andrewstuart/goq"
	"log"
	"net/http"
	"regexp"
	"strings"
	"github.com/essentialkaos/translit"
)

const (
	LUBAVA_URL = "https://multiplex.ua/cinema/cherkasy/lyubava"
	PLAZA_URL  = "https://multiplex.ua/cinema/cherkasy/dniproplaza"
	URL_PREFIX = "https://multiplex.ua/"
)

var KinoUrls = map[string]string{
	"lyubava": LUBAVA_URL,
	"plaza":   PLAZA_URL,
}

var KinoNames = map[string]string{
	"lyubava": "Lyubava ",
	"plaza":   "Dniproplaza ",
}

type Films struct {
	FilmTds []Film `goquery:"div.film"`
}

type Film struct {
	Img       string `goquery:"div.poster,[style]"`
	Link      string `goquery:"a,[href]"`
	Title     string `goquery:".info a"`
	TimeBlock string `goquery:".info .sessions,text"`
}

func GetMovies(cinema string) []Film {
	url, ok := KinoUrls[cinema]
	if !ok {
		log.Fatal("Cinema not found!")
	}

	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()

	var ex Films

	err = goq.NewDecoder(res.Body).Decode(&ex)
	if err != nil {
		log.Fatal(err)
	}

	re := regexp.MustCompile(`\s+|\n+`)
	for idx, flm := range ex.FilmTds {
		img := strings.TrimRight(flm.Img, "')")
		img = strings.TrimLeft(img, "background-image: url('")
		times := re.ReplaceAllLiteralString(flm.TimeBlock, " ")
		ex.FilmTds[idx].Img = img
		ex.FilmTds[idx].TimeBlock = times
	}

	return ex.FilmTds
}

func GetMovieListResponse(films []Film, cinema string) models.Data {
	var empty string
	if len(films) == 0 {
		empty = " is empty"
	}

	name := KinoNames[cinema]
	data := models.Data{}
	data.Google.ExpectUserResponse = false
	data.Google.RichResponse = models.RichResponse{
		Items: []map[string]interface{}{
			{
				"simpleResponse": models.SimpleResponse{
					DisplayText:  name + "films list" + empty,
					TextToSpeech: name + "films list" + empty,
				},
			},
		},
		Suggestions: []models.Suggestion{
			{
				Title: "Lubava",
			},
			{
				Title: "Dniproplaza",
			},
		},
	}

	var items []models.CouruselItems
	var uniq = map[string]string{}
	var speechFilms = ""

	for _, film := range films {
		if _, ok := uniq[film.Title]; ok {
			continue
		} else {
			uniq[film.Title] = film.Title
		}

		speechFilms += "film " + translit.EncodeToISO9A(film.Title) + " \n  \n"

		var item = models.CouruselItems{
			Title:       film.Title,
			Description: film.TimeBlock,
			Image: models.Image{
				URL_PREFIX + film.Img,
				film.Title,
			},
			OptionInfo: models.OptionInfo{
				Key:      film.Title,
				Synonyms: []string{},
			},
		}
		items = append(items, item)
	}

	simpleTitle := map[string]interface{}{
		"simpleResponse": models.SimpleResponse{
			DisplayText:  "",
			TextToSpeech: speechFilms,
		},
	}
	data.Google.RichResponse.Items = append(data.Google.RichResponse.Items, simpleTitle)

	data.Google.SystemIntent = models.SystemIntent{
		Intent: "actions.intent.OPTION",
		Data: models.CouruselData{
			Type: "type.googleapis.com/google.actions.v2.OptionValueSpec",
			CarouselSelect: models.CarouselSelect{
				Items: items,
			},
		},
	}



	return data
}
