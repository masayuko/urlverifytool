package cmd

import (
	"bufio"
	"encoding/csv"
	"fmt"
	//"github.com/loadoff/excl"
	"github.com/spf13/cobra"
	//"github.com/tealeg/xlsx"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

//	"golang.org/x/text/unicode/norm"

func init() {
	RootCmd.AddCommand(urllistCmd)
	urllistCmd.Flags().BoolVarP(&optNoHeader, "noheader", "z", false, "If true then it will not treat first line as header")
	urllistCmd.Flags().StringVarP(&optUrlColumnName, "columnname", "n", "URL", "Specify column name")
	urllistCmd.Flags().Int64VarP(&optUrlColumnIndex, "columnindex", "c", -1, "Specify column index")
	urllistCmd.Flags().BoolVarP(&optStdout, "stdout", "v", false, "Output to STDOUT")
	urllistCmd.Flags().BoolVarP(&optStrict, "strict", "s", false, "don't use LazyQuotes mode")
	urllistCmd.Flags().StringVarP(&optComma, "delimiter", "d", "", "string of delimiter")
	urllistCmd.Flags().StringVarP(&optDecoder, "decoder", "i", "utf-8", "decorder for reader. default is utf-8 [utf-8, shift_jis, cp932]")
	urllistCmd.Flags().Int64VarP(&optSplitCount, "splitcount", "j", -1, "Specify count of lines")
}

var urllistCmd = &cobra.Command{
	Use:   "urllist",
	Short: "make URL list from CSV",
	Long:  `make URL list from CSV`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		urllist(args)
	},
}

func urllist(args []string) {
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

	if optSplitCount == 0 {
		optSplitCount = -1
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

	toUrlList(reader, basename)
}

func toUrlList(reader *csv.Reader, basename string) {

	firstline, err := reader.Read()
	if err == io.EOF {
		return
	}
	if err != nil {
		er(err)
	}

	var outputName string
	var osFile *os.File
	var bWriter *bufio.Writer
	if optSplitCount < 0 {
		outputName = basename + "_url.txt"
	} else {
		outputName = fmt.Sprintf("%s_url%d.txt", basename, 0)
	}
	osFile, err = os.Create(outputName)
	if err != nil {
		er(err)
	}
	bWriter = bufio.NewWriter(osFile)

	restcnt := optSplitCount
	fcnt := 0

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
		_, err = bWriter.WriteString(firstline[optUrlColumnIndex])
		if err != nil {
			er(err)
		}
		_, err = bWriter.WriteString("\n")
		if err != nil {
			er(err)
		}
		if restcnt > 0 {
			restcnt--
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
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			er(err)
		}

		if restcnt == 0 {
			fcnt++
			bWriter.Flush()
			osFile.Close()
			outputName = fmt.Sprintf("%s_url%d.txt", basename, fcnt)
			osFile, err = os.Create(outputName)
			if err != nil {
				er(err)
			}
			bWriter = bufio.NewWriter(osFile)
			restcnt = optSplitCount
		}

		_, err = bWriter.WriteString(record[optUrlColumnIndex])
		//_, err = bWriter.WriteString(strings.Join(record, ","))
		if err != nil {
			er(err)
		}
		_, err = bWriter.WriteString("\n")
		if err != nil {
			er(err)
		}
		if restcnt > 0 {
			restcnt--
		}
	}

	bWriter.Flush()
	osFile.Close()
}
