package model

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"code.sajari.com/docconv"
	"github.com/agnivade/levenshtein"
	"github.com/franzwilhelm/uio-exam-helper/db"
	log "github.com/sirupsen/logrus"
)

type Resource struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	URL       string `json:"URL" gorm:"unique_index"`
	SubjectID string `json:"subjectID"`
}

func (r *Resource) Folder() string {
	return fmt.Sprintf("resources/%s", r.SubjectID)
}

func (r *Resource) FilePath() string {
	return fmt.Sprintf("%s/%s", r.Folder(), r.Name)

}

func (r *Resource) UseURL(url string) {
	r.URL = url
	tokens := strings.Split(url, "/")
	name := tokens[len(tokens)-1]
	if filepath.Ext(name) == "" {
		name += ".pdf"
	}
	r.Name = name
}

func (r *Resource) IsDownloaded() bool {
	_, err := os.Stat(r.FilePath())
	return err == nil
}

func (r *Resource) Download() error {
	if err := os.MkdirAll(r.Folder(), 0777); err != nil && !os.IsExist(err) {
		log.WithError(err).Error("Error while creating folder " + r.Folder())
		return err
	}
	file, err := os.Create(r.FilePath())
	defer file.Close()
	if err != nil {
		log.WithError(err).Error("Error while creating " + r.Name)
		return err
	}

	log.Infof("Downloading %s to %s", r.URL, r.Name)
	log.Warn(r.URL)
	response, err := http.Get(r.URL)
	if response == nil {
		log.Error("go not response from", r.URL)
		return nil
	}
	defer response.Body.Close()
	if err != nil {
		log.WithError(err).Error("Error while downloading" + r.URL)
		return err
	}

	_, err = io.Copy(file, response.Body)
	if err != nil {
		log.WithError(err).Error("Error while copying to" + file.Name())
		return err
	}
	return nil
}

func (r *Resource) Delete() error {
	log.Warn("Deleting file: ", r.Name)
	return os.Remove(r.Name)
}

func (r *Resource) GetWords() ([]string, error) {
	res, err := docconv.ConvertPath(r.FilePath())
	if err != nil {
		log.WithError(err).Error("Could not covert pdf to text", r.FilePath())
		return []string{}, nil
	}
	lower := strings.ToLower(res.Body)
	lines := strings.Split(lower, "\n")
	words := words(lines)
	processed, err := deleteSpecials(words)
	return processed, err
}

func ResourceByURL(url string) (r Resource, err error) {
	return r, db.Default.Where("url = ?", url).First(&r).Error
}

func words(lines []string) (words []string) {
	for _, line := range lines {
		words = append(words, strings.Split(line, " ")...)
	}
	return
}

func deleteSpecials(words []string) (processed []string, err error) {
	for _, line := range words {
		reg, err := regexp.Compile("[^a-zA-ZæøåÆØÅ]+")
		if err != nil {
			return nil, err
		}
		replaced := reg.ReplaceAllString(line, "")
		if len(replaced) > 2 {
			processed = append(processed, replaced)
		}
	}
	return
}

type Group struct {
	Count int
	Words []string
}

func groupWords(a []string) []Group {
	countMap := make(map[string]int)
	for _, s := range a {
		countMap[s]++
	}

	for _, word := range frequentWords {
		delete(countMap, word)
	}

	for _, word := range examWords {
		delete(countMap, word)
	}

	grouped := make(map[string]bool)
	groups := make([]Group, 0)
	for w1, c1 := range countMap {
		if _, ok := grouped[w1]; ok {
			continue
		}
		g := Group{
			Count: c1,
			Words: []string{w1},
		}
		for w2, c2 := range countMap {
			if _, ok := grouped[w2]; ok {
				continue
			}
			distance := levenshtein.ComputeDistance(w1, w2)
			if distance == 1 || distance == 2 {
				grouped[w2] = true
				g.Count += c2
				g.Words = append(g.Words, w2)
			}
		}
		groups = append(groups, g)
	}

	return groups
}

var frequentWords = []string{
	"og", "i", "det", "på", "som", "er", "en", "til", "å", "han", "av", "for", "med", "at", "var", "de", "ikke", "den", "har", "jeg", "om", "et", "men", "så", "seg", "hun", "hadde", "fra", "vi", "du", "kan", "da", "ble", "ut", "skal", "vil", "ham", "etter", "over", "ved", "også", "bare", "eller", "sa", "nå", "dette", "noe", "være", "meg", "mot", "opp", "der", "når", "inn", "dem", "kunne", "andre", "blir", "alle", "noen", "sin", "ha", "år", "henne", "må", "selv", "sier", "få", "kom", "denne", "enn", "to", "hans", "bli", "ville", "før", "vært", "skulle", "går", "her", "slik", "gikk", "mer", "hva", "igjen", "fikk", "man", "alt", "mange", "dash", "ingen", "får", "oss", "hvor", "under", "siden", "hele", "dag", "gang", "sammen", "ned", "kommer", "sine", "deg", "se", "første", "godt", "mellom", "måtte", "gå", "helt", "litt", "nok", "store", "aldri", "ta", "sig", "uten", "ho", "kanskje", "blitt", "ser", "hvis", "bergen", "sitt", "jo", "vel", "si", "vet", "hennes", "min", "tre", "ja", "samme", "mye", "nye", "tok", "gjøre", "disse", "siste", "tid", "rundt", "tilbake", "mens", "satt", "flere", "folk", "fordi", "både", "la", "gjennom", "fått", "like", "nei", "annet", "komme", "kroner", "gjorde", "hvordan", "norge", "norske", "gjør", "oslo", "står", "stor", "gamle", "langt", "annen", "sett", "først", "mener", "hver", "barn", "rett", "ny", "tatt", "derfor", "fram", "hos", "heller", "lenge", "alltid", "tror", "nesten", "mann", "gi", "god", "lå", "blant", "norsk", "gjort", "visste", "bak", "tar", "liv", "mennesker", "frem", "bort", "ein", "verden", "deres", "ikkje", "tiden", "del", "vår", "mest", "eneste", "likevel", "hatt", "dei", "tidligere", "fire", "liten", "hvorfor", "tenkte", "hverandre", "holdt", "bedre", "meget", "ting", "lite", "stod", "ei", "hvert", "begynte", "gir", "ligger", "grunn", "dere", "livet", "a", "sagt", "land", "kommet", "e", "neste", "far", "efter", "egen", "side", "gått", "mor", "ute", "videre", "millioner", "prosent", "svarte", "sto", "begge", "allerede", "inne", "finne", "enda", "hjem", "foran", "måte", "mannen", "dagen", "hodet", "saken", "ganger", "kjente", "stort", "blev", "mindre", "landet", "byen", "plass", "kveld", "ord", "øynene", "fem", "større", "gode", "nu", "synes", "beste", "kvinner", "ett", "satte", "hvem", "all", "klart", "holde", "ofte", "stille", "spurte", "lenger", "sted", "dager", "mulig", "utenfor", "små", "frå", "nytt", "slike", "viser", "mig", "kjenner", "samtidig", "senere", "særlig", "våre", "akkurat", "menn", "hørte", "mdash", "arbeidet", "altså", "par", "din", "unge", "n", "borte", "plutselig", "fant", "fast", "kunde", "snart", "svært", "fall", "vei", "bergens", "dessuten", "forhold", "gjerne", "snakket", "foto", "snakke", "bør", "dersom", "imidlertid", "lett", "tenke", "gud", "tro", "jan", "gitt", "penger", "egentlig", "mitt", "ønsker", "ansiktet", "kl", "dermed", "slo", "politiet", "faren", "eit", "bra", "je", "sitter", "sikkert", "vite", "full", "lille", "glad", "fleste", "slutt", "ene", "mine", "gjelder", "lagt", "virkelig", "laget", "alene", "ennå", "lang", "ganske", "johan", "omkring", "hjemme", "vårt", "vanskelig", "arne", "gammel", "skulde", "tidende", "riktig", "huset", "følte", "møte", "lørdag", "klar", "m", "kort", "viktig", "ellers", "minst", "fortsatt", "op", "veien", "seier", "mål", "kjent", "slags", "frode", "stund", "arbeid", "finnes", "ingenting", "lange", "gangen", "stå", "lot", "rekke", "redd", "høre", "vilde", "ga", "ti", "forteller", "overfor", "stadig", "burde", "visst", "syntes", "fjor", "sette", "funnet", "hjelp", "største", "løpet", "meter", "norges", "hånden", "spørsmål", "s", "mente", "søndag", "f", "følge", "fremdeles", "imot", "hus", "kvinne", "ventet", "reiste", "hendene", "trodde", "usa", "legger", "viste", "regjeringen", "eg", "årene", "eksempel", "tenkt", "ole", "slikt", "erik", "moren", "holder", "seks", "tenker", "stedet", "tillegg", "helst", "bruke", "skolen", "kampen", "nettopp", "døren", "egne", "eget", "sterkt", "betyr", "vant", "enkelte", "nærmere", "hvad", "dårlig", "per", "trenger", "menneske", "måten", "vise", "oppe", "finner", "legge", "januar", "februar", "mars", "april", "mai", "juni", "juli", "august", "september", "oktober", "november", "desember", "pa", "a", "sa", "ogsa", "na", "nar", "ar", "ma", "fa", "gar", "far", "matte", "ga", "bade", "fatt", "star", "la", "var", "gatt", "mate", "sma", "fra", "vare", "altsa", "enna", "vart", "mal", "sta", "handen", "spørsmal", "arene", "darlig", "maten", "mandag", "tirsdag", "onsdag", "torsdag", "fredag", "lørdag", "søndag",
}

var examWords = []string{
	"eksamen", "oppgave", "poeng", "svar", "svarene", "svaret", "løsningen", "fortsettes", "feil", "riktig", "universitetet", "universitet", "finn", "løsning", "begrunn", "oppgavesettet", "dvs", "beregn", "bestem", "betrakt", "angi", "systemet", "deretter", "vis", "betegne", "minste", "definer", "tilhørende", "vedlegg", "begynner", "tillatte", "besvare", "sider", "eksamensdag", "hjelpemidler", "fakultet", "betrakter", "kontroller", "matematisknaturvitenskapelige", "hensyn", "lar", "variabel", "vanlige", "lik", "relativt", "består", "løsninger", "betegner", "alternativt", "definert", "merk", "spørsmålene", "oppgaven", "henvise", "hensiktsmessig", "notat", "utregning", "page", "løser", "setter", "lykke", "oppfyller", "examination", "sjekk", "henholdsvis", "vedlagte", "bestar", "løse", "forklar", "varierer", "kjernen", "spørsmalene", "generelle", "anta", "vedlagt", "bruk", "problem", "opplagt", "løsningene", "forrige", "feks", "vanlig", "bruker", "høyst", "mengden", "enhver", "best", "utskrift", "skriv", "løsningsforslag", "oppgitte", "oppgavene", "følgende", "deloppgaver", "definerer", "opptil", "tilsvarende", "ønsket",
}
