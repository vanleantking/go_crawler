package wordprocess

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
)

const (
	ScriptFunc    = `function\s*\w*\s*\(\w*\)\s*{.*?}(\s}\);\s*})?`
	DeclareVar    = `(var)?\s*([a-zA-Z_][a-zA-Z0-9_\.]*)\s*=(.*?;)?`
	Func          = `([a-zA-Z0-9_\.])*\s*\((.*?)\s*\)\s*;`
	Comment       = `(<!--.*?-->)|(<!--[\w\W\n\s]+?-->|(\/\*[\s\S]*?\*\/|([^:]|^)\/\/.*$))`
	GoogleTag     = `googletag\.cmd\.push`
	IfCondition   = `(if\()?document\.getElementById\((.*?)\)`
	StyleCSS      = `([\#\-\.@]?[a-zA-Z0-9])([\#\-\.@]?[a-zA-Z0-9]\s?)*{.*?}`
	MissSpaceWord = `(\p{L}+)([A-Z]\p{L}*)`
	VNUpperCase   = `[A-ZĐẠÀÁẢÃÂẤẦẨẪẬĂẮẰẲẴẶẸẺẼÉÈẾỀỂỄỆÊỈỊĨÍÌỌÓÒÕỎỐỒỔỖỘÔƠỚỜỞỠỢỤỦÚÚŨỨỪỬỮỰƯỲÝỴỶỸ]+`
	BiGramsFile   = "../wordprocess/bi-grams.txt"
	TriGramsFile  = "../wordprocess/tri-grams.txt"
	VNStopWords   = "../wordprocess/vnmstopwords.txt"
)

type WordProcess struct {
	Content             string
	Tokenizer           [][]string
	DocumentTermsMatrix map[string]float32
}

var StopWords = GenerateStopWordsList()
var BiGrams = GenerateNGramWords(BiGramsFile)
var TriGrams = GenerateNGramWords(TriGramsFile)

var SpecialToken = []string{"}", "{", "*", "&", "#", ">", "<", "]", "[", "\"", "-", "–", "/",
	",", "!", "?", "”", "“", "(", ")", "'", "`", "_", "+", ":", "=", "‘", "’", "if", "else", "^",
	"@", "″", "", "%", ".", "|", "function", ";", ".", "↑", "$", "´", "m²", "…"}

func (wp *WordProcess) RmScriptFuncDeclare() {
	re := regexp.MustCompile(ScriptFunc)
	wp.Content = re.ReplaceAllLiteralString(wp.Content, "")
}

func (wp *WordProcess) RmScriptDeclare() {
	re := regexp.MustCompile(DeclareVar)
	wp.Content = re.ReplaceAllLiteralString(wp.Content, "")
}

func (wp *WordProcess) RmScriptFunc() {
	re := regexp.MustCompile(Func)
	wp.Content = re.ReplaceAllLiteralString(wp.Content, "")
}

func (wp *WordProcess) RmComment() {
	re := regexp.MustCompile(Comment)
	wp.Content = re.ReplaceAllLiteralString(wp.Content, "")
}

func (wp *WordProcess) RmGoogleTag() {
	re := regexp.MustCompile(GoogleTag)
	wp.Content = re.ReplaceAllLiteralString(wp.Content, "")
}

func (wp *WordProcess) RmStyleCss() {
	re := regexp.MustCompile(StyleCSS)
	wp.Content = re.ReplaceAllLiteralString(wp.Content, "")
}

func (wp *WordProcess) RmIfCondition() {
	re := regexp.MustCompile(IfCondition)
	wp.Content = re.ReplaceAllLiteralString(wp.Content, "")
}

func (wp *WordProcess) RmNonCharater() {
	re := regexp.MustCompile(`\d+[h|km|m]?\d*`)
	wp.Content = re.ReplaceAllLiteralString(wp.Content, "")
}

// will implement later for regex
func (wp *WordProcess) AddSpaceWord() {
	re := regexp.MustCompile(MissSpaceWord)
	rs := re.FindAllStringSubmatch(wp.Content, -1)
	for _, words := range rs {
		// check string is all vietnamese upper case
		isUpper := regexp.MustCompile(VNUpperCase)
		if strings.TrimSpace(isUpper.FindString(words[0])) != strings.TrimSpace(words[0]) {
			wp.Content = strings.Replace(wp.Content, words[0], words[1]+" "+words[2], -1)
		}

	}

}

func rmSpecialToken(content string) string {
	for _, token := range SpecialToken {
		content = strings.Replace(content, token, " ", -1)
	}

	// re1 := regexp.MustCompile(`\.+`)
	// content = re1.ReplaceAllLiteralString(content, " ")
	re2 := regexp.MustCompile(`\s+`)
	content = re2.ReplaceAllLiteralString(content, " ")
	return strings.TrimSpace(content)
}

func (wp *WordProcess) CleanData() {
	wp.RmScriptFuncDeclare()
	wp.RmStyleCss()
	wp.RmScriptDeclare()
	wp.RmScriptFunc()
	wp.RmComment()
	wp.RmGoogleTag()
	wp.RmIfCondition()
	wp.RmNonCharater()
	wp.Content = strings.ToLower(rmSpecialToken(wp.Content))
}

// Tokenizer by dot "."
func (wp *WordProcess) Tokenizers() {
	tokenizer := [][]string{}
	if wp.Content != "" {
		string_arrs := strings.Split(wp.Content, ".")
		for _, line := range string_arrs {
			line = rmSpecialToken(line)
			pieces := strings.Split(line, " ")
			tokenizer = append(tokenizer, pieces)
		}
	}
	// word by word
	wp.Tokenizer = tokenizer
}

// Tokenizer by space
func (wp *WordProcess) TokenizersbySpace() {

	tokenizer := map[string]float32{}
	if wp.Content != "" {
		// split word by space (tab, newline, space...)
		words := regexp.MustCompile(`\s+`).Split(wp.Content, -1)
		w_length := len(words)
		index := 0
		for index < w_length {
			cur_word := strings.TrimSpace(words[index])

			// if cur_word != "" {
			if index+3 < w_length {
				next_word := words[index+1]
				next_next_word := words[index+2]
				bi_grams := strings.TrimSpace(strings.ToLower(cur_word + " " + next_word))
				tri_grams := strings.TrimSpace(strings.ToLower(bi_grams + " " + next_next_word))

				if TriGrams[tri_grams] == true {
					if _, ok := tokenizer[tri_grams]; ok {
						tokenizer[tri_grams] = tokenizer[tri_grams] + 1
					} else {
						tokenizer[tri_grams] = 1
					}
					index += 3
				} else {
					if BiGrams[bi_grams] == true {
						if _, ok := tokenizer[bi_grams]; ok {
							tokenizer[bi_grams] = tokenizer[bi_grams] + 1
						} else {
							tokenizer[bi_grams] = 1
						}
						index += 2
					} else {
						if _, ok := tokenizer[cur_word]; ok {
							tokenizer[cur_word] = tokenizer[cur_word] + 1
						} else {
							tokenizer[cur_word] = 1
						}
						index += 1
					}
				}
			} else {
				if index+2 < w_length {
					next_word := words[index+1]
					bi_grams := strings.ToLower(cur_word + " " + next_word)
					if BiGrams[bi_grams] == true {
						if _, ok := tokenizer[bi_grams]; ok {
							tokenizer[bi_grams] = tokenizer[bi_grams] + 1
						} else {
							tokenizer[bi_grams] = 1
						}
						index += 2
					} else {
						if _, ok := tokenizer[cur_word]; ok {
							tokenizer[cur_word] = tokenizer[cur_word] + 1
						} else {
							tokenizer[cur_word] = 1
						}
						index += 1
					}
				} else {
					if index+1 < w_length {
						if _, ok := tokenizer[cur_word]; ok {
							tokenizer[cur_word] = tokenizer[cur_word] + 1
						} else {
							tokenizer[cur_word] = 1
						}
						index += 1
					} else {
						if _, ok := tokenizer[cur_word]; ok {
							tokenizer[cur_word] = tokenizer[cur_word] + 1
						} else {
							tokenizer[cur_word] = 1
						}
						break
					}
				}
			}
			// }
		}
	}
	// word by word
	wp.DocumentTermsMatrix = tokenizer
}

func (wp *WordProcess) RemoveStopwords() {
	for key, _ := range wp.DocumentTermsMatrix {
		if StopWords[key] {
			delete(wp.DocumentTermsMatrix, key)
		}
	}
}

// Sort word pairs by frequencies
func SortByFrequencies(wordFrequencies map[string]float32) PairList {
	pl := make(PairList, len(wordFrequencies))
	i := 0
	for k, v := range wordFrequencies {
		pl[i] = Pair{k, v}
		i++
	}
	sort.Sort(sort.Reverse(pl))
	return pl
}

func SortByKeys(wordFrequencies map[string]float32) PairList {
	result := make(PairList, len(wordFrequencies))
	mk := make([]string, len(wordFrequencies))
	i := 0
	for k, _ := range wordFrequencies {
		mk[i] = k
		i++
	}
	sort.Strings(mk)
	j := 0
	for _, key := range mk {
		result[j] = Pair{key, wordFrequencies[key]}
		j++
	}
	return result
}

func GenerateStopWordsList() map[string]bool {
	var stopwords = map[string]bool{}
	file, err := os.Open(VNStopWords)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		stopwords[strings.TrimSpace(scanner.Text())] = true
	}
	return stopwords
}

func GenerateNGramWords(path string) map[string]bool {
	var biGrams = map[string]bool{}
	// Open our jsonFile
	jsonFile, err := os.Open(path)
	// if we os.Open returns an error then handle it
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var result []string
	_ = json.Unmarshal([]byte(byteValue), &result)
	for _, word := range result {
		biGrams[word] = true
	}
	return biGrams
}

// zip two list as zip python function
func zip(a1, a2 []string) []string {
	r := make([]string, 2*len(a1))
	for i, e := range a1 {
		r[i*2] = e
		r[i*2+1] = a2[i]
	}
	return r
}

func AddSpaceWord(words string) string {
	re := regexp.MustCompile(MissSpaceWord)
	rs := re.FindAllStringSubmatch(words, -1)
	for _, pieces := range rs {
		// check string is all vietnamese upper case
		isUpper := regexp.MustCompile(VNUpperCase)
		if strings.TrimSpace(isUpper.FindString(pieces[0])) != strings.TrimSpace(pieces[0]) {
			words = strings.Replace(words, pieces[0], pieces[1]+" "+pieces[2], -1)
		}
	}
	return words

}
