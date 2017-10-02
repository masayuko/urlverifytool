package cmd

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func init() {
	RootCmd.AddCommand(mergeCmd)
	mergeCmd.Flags().StringVarP(&optTargetFormat, "targetformat", "t", "csv", "Target format. default \"csv\" [csv, json]")
	mergeCmd.Flags().BoolVarP(&optNoHeader, "noheader", "z", false, "If true then it will not treat first line as header")
	mergeCmd.Flags().StringVarP(&optUrlColumnName, "columnname", "n", "URL", "Specify column name")
	mergeCmd.Flags().Int64VarP(&optUrlColumnIndex, "columnindex", "c", -1, "Specify column index")
	mergeCmd.Flags().BoolVarP(&optStdout, "stdout", "v", false, "Output to STDOUT")
	mergeCmd.Flags().BoolVarP(&optStrict, "strict", "s", false, "don't use LazyQuotes mode")
	mergeCmd.Flags().StringVarP(&optComma, "delimiter", "d", "", "string of delimiter")
	mergeCmd.Flags().StringVarP(&optDecoder, "decoder", "i", "utf-8", "decorder for reader. default is utf-8 [utf-8, shift_jis, cp932]")
}

var mergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "merge result.json",
	Long:  `merge result.json`,
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		merge(args)
	},
}

type siteCheckerResult struct {
	ID          string      `json:"id"`
	URL         string      `json:"url"`
	Status      json.Number `json:"status"`
	Title       string      `json:"title"`
	ResponseURL string      `json:"response_url"`
	FilePath    string      `json:"filepath"`
	FileName    string      `json:"filename"`
}

func merge(args []string) {
	_, inputfile := filepath.Split(args[0])
	ext := filepath.Ext(inputfile)
	var basename string
	if len(ext) == 0 {
		basename = inputfile
	} else if ext == "." {
		basename = inputfile[:len(inputfile)-1]
	} else {
		re := regexp.MustCompile(ext + "$")
		basename = re.ReplaceAllString(inputfile, "")
	}

	if len(basename) == 0 {
		er(fmt.Errorf("invalid file name : %s", args[0]))
	}

	fp, err := os.Open(args[0])
	if err != nil {
		er(err)
	}
	defer fp.Close()

	var ioReader io.Reader
	lDecoder := strings.ToLower(optDecoder)
	if lDecoder == "utf-8" || lDecoder == "utf8" {
		ioReader = fp
	} else if lDecoder == "shift_jis" ||
		lDecoder == "sjis" ||
		lDecoder == "cp932" {
		ioReader = transform.NewReader(fp, japanese.ShiftJIS.NewDecoder())
	} else if lDecoder == "eucjp" {
		ioReader = transform.NewReader(fp, japanese.EUCJP.NewDecoder())
	} else if lDecoder == "iso2022-jp" ||
		lDecoder == "iso2022jp" {
		ioReader = transform.NewReader(fp, japanese.ISO2022JP.NewDecoder())
	} else {
		er(fmt.Errorf("invalid decodrder : %s", optDecoder))
	}

	if len(optComma) == 0 {
		optComma = ","
	}

	reader := csv.NewReader(ioReader)
	reader.Comma = []rune(optComma[0:1])[0]
	reader.LazyQuotes = !optStrict

	resultMap := readSiteCheckerResult(args[1])

	switch optTargetFormat {
	case "csv":
		mergeToCSV(reader, resultMap, basename)
	default:
		er(fmt.Errorf("invalid target format: %s", optTargetFormat))
	}
}

func mergeToCSV(reader *csv.Reader, resultMap map[string]siteCheckerResult, basename string) {
	osFile, err := os.Create(basename + "_merged.csv")
	if err != nil {
		er(err)
	}
	defer osFile.Close()
	tWriter := transform.NewWriter(osFile, japanese.ShiftJIS.NewEncoder())
	csvWriter := csv.NewWriter(tWriter)
	csvWriter.UseCRLF = true

	not200File, err := os.Create(basename + "_not200.txt")
	if err != nil {
		er(err)
	}
	defer not200File.Close()
	bnot200Writer := bufio.NewWriter(not200File)

	eFile, err := os.Create(basename + "_error.txt")
	if err != nil {
		er(err)
	}
	defer eFile.Close()
	beWriter := bufio.NewWriter(eFile)

	firstline, err := reader.Read()
	if err == io.EOF {
		return
	}
	if err != nil {
		er(err)
	}

	replr := strings.NewReplacer("\\", "/")

	if optNoHeader {
		if optUrlColumnIndex == -1 {
			re := regexp.MustCompile("^http")
			for i, value := range firstline {
				if len(re.FindString(value)) > 0 {
					optUrlColumnIndex = int64(i)
					break
				}
			}
		}
		if optUrlColumnIndex == -1 {
			er(fmt.Errorf("cannot realize URL field"))
		}
		result, ok := resultMap[firstline[optUrlColumnIndex]]
		if ok {
			status := result.Status.String()
			switch status {
			case "ERROR":
				_, err = beWriter.WriteString(result.URL)
				if err != nil {
					er(err)
				}
				_, err = beWriter.WriteString("\n")
				if err != nil {
					er(err)
				}
				firstline = append(firstline, status)
			case "200":
				firstline = append(firstline, "file://"+replr.Replace(result.URL))
			default:
				_, err = bnot200Writer.WriteString(result.URL)
				if err != nil {
					er(err)
				}
				_, err = bnot200Writer.WriteString("\n")
				if err != nil {
					er(err)
				}
				firstline = append(firstline, status)
			}
			err := csvWriter.Write(firstline)
			if err != nil {
				er(err)
			}
		}
	} else {
		if optUrlColumnIndex == -1 {
			for i, value := range firstline {
				if value == optUrlColumnName {
					optUrlColumnIndex = int64(i)
				}
			}
			if optUrlColumnIndex == -1 {
				er(fmt.Errorf("column name not found : %s", optUrlColumnName))
			}
		}
		firstline = append(firstline, "image")
		err := csvWriter.Write(firstline)
		if err != nil {
			er(err)
		}
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			er(err)
		}

		result, ok := resultMap[record[optUrlColumnIndex]]
		if ok {
			status := result.Status.String()
			switch status {
			case "ERROR":
				_, err = beWriter.WriteString(result.URL)
				if err != nil {
					er(err)
				}
				_, err = beWriter.WriteString("\n")
				if err != nil {
					er(err)
				}
				record = append(record, status)
			case "200":
				record = append(record, "file://"+replr.Replace(result.URL))
			default:
				_, err = bnot200Writer.WriteString(result.URL)
				if err != nil {
					er(err)
				}
				_, err = bnot200Writer.WriteString("\n")
				if err != nil {
					er(err)
				}
				record = append(record, status)
			}
			err := csvWriter.Write(record)
			if err != nil {
				er(err)
			}
		}
	}

	csvWriter.Flush()
	bnot200Writer.Flush()
	beWriter.Flush()
}

func readSiteCheckerResult(path string) map[string]siteCheckerResult {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		er(err)
	}

	var results []siteCheckerResult
	err = json.Unmarshal(bytes, &results)
	if err != nil {
		er(err)
	}

	resultMap := make(map[string]siteCheckerResult)
	for _, result := range results {
		resultMap[result.URL] = result
	}

	return resultMap
}

/*
func (s *statusCode) UnmarshalJSON(b []byte) error {
	var n json.Number
	err := json.Unmarshal(b, &n)
	if err != nil {
		return err
	}
	i, err := n.Int64()
	if err != nil {
		return err
	}
	*s = statusCode(i)
	return nil
}
*/
