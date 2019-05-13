package structs

import (
	"os"
	"strings"

	xlsx "github.com/360EntSecGroup-Skylar/excelize"
	"github.com/extrame/xls"

	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

const (
	GoogleSheet = "google_sheet"
	CSV         = "csv"
	XLSX        = "xlsx"
	XLS         = "xls"
)

type Reader struct {
	Rows   [][]string
	Type   string
	Path   string
	Header map[string]int
}

func (reader *Reader) SetType(ext string) {
	reader.Type = strings.ToLower(ext)
}

func (reader *Reader) GetData() error {
	var err error
	switch reader.Type {
	case CSV:
		err = reader.CSVReader()
	case XLS:
		err = reader.XLSReader()
	case XLSX:
		err = reader.XLSXReader()
	case GoogleSheet:
		err = reader.SheetReader()
	}

	if err == nil && len(reader.Rows) == 0 {
		err = errors.New("File have no contain data")
	}
	return err
}

func (reader *Reader) GetHeaderFileImport() {
	reader.Header = map[string]int{}
	for index, header := range reader.Rows[0] {
		header = strings.TrimSpace(header)
		if header != "" {
			reader.Header[strings.TrimSpace(header)] = index
		}
	}
}

// For read CSV file
func (reader *Reader) CSVReader() error {

	file, err := os.Open(reader.Path)
	if err != nil {
		return err
	}
	defer file.Close()

	// Read File into a Variable
	lines, err := csv.NewReader(file).ReadAll()

	if err != nil {
		return err
	}

	reader.Rows = lines
	return nil
}

func (reader *Reader) XLSReader() error {
	xlFile, err := xls.Open(reader.Path, "utf-8")
	if err == nil {
		reader.Rows = xlFile.ReadAllCells(10000)
		return nil
	} else {
		return err
	}
}

func (reader *Reader) XLSXReader() error {
	sheet := ""
	xlsx, err := xlsx.OpenFile(reader.Path)
	if err != nil {
		return err
	}

	index := xlsx.GetActiveSheetIndex()
	sheet = xlsx.GetSheetName(index)

	reader.Rows = xlsx.GetRows(sheet)
	return nil
}

func getSpreadsheetId(url string) string {

	re := regexp.MustCompile(`\/spreadsheets\/d\/([a-zA-Z0-9-_]+)`)
	result := re.FindString(url)
	piece := strings.Split(result, "/")
	return piece[len(piece)-1]
}

func getSheetId(url string) int64 {
	result := ""
	re := regexp.MustCompile(`[#&]gid=([0-9]+)`)
	result = re.FindString(url)
	piece := strings.Split(result, "=")
	result = piece[len(piece)-1]
	if result == "" {
		return 0
	}
	if i, err := strconv.Atoi(result); err == nil {
		return int64(i)
	}
	return 0

}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "../apps/golang/read_files/token.json"
	// tokFile := "../../golang/read_files/token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		panic(err.Error())
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		panic(err.Error())
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		panic(err.Error())
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func getSheetTitle(sheets []*sheets.Sheet, sheetid int64) string {
	for _, sheet := range sheets {
		if sheet.Properties.SheetId == sheetid {
			return sheet.Properties.Title
		}
	}
	return ""

}

func (reader *Reader) SheetReader() error {
	credentialFile := "../apps/golang/read_files/structs/credentials.json"
	// credentialFile := "../../golang/read_files/structs/credentials.json"
	b, err := ioutil.ReadFile(credentialFile)
	if err != nil {
		return err
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets.readonly")
	if err != nil {
		return err
	}
	client := getClient(config)

	srv, err := sheets.New(client)
	if err != nil {
		return err
	}

	spreadsheetId := getSpreadsheetId(reader.Path)
	sheetId := getSheetId(reader.Path)

	sheets_property, nn := srv.Spreadsheets.Get(spreadsheetId).Do()
	if nn != nil {
		return err
	}

	sheet_title := getSheetTitle(sheets_property.Sheets, sheetId)

	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, sheet_title).Do()
	if err != nil {
		return err
	}

	if len(resp.Values) > 0 {
		for _, row := range resp.Values {
			s := make([]string, len(row))
			for i, v := range row {
				s[i] = fmt.Sprint(v)
			}
			reader.Rows = append(reader.Rows, s)
		}
	}
	return nil
}
