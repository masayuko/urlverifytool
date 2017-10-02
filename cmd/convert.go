package cmd

import (
	"encoding/csv"
	"encoding/json"
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
	RootCmd.AddCommand(convertCmd)
	convertCmd.Flags().StringVarP(&optTargetFormat, "targetformat", "t", "json", "Target format. default \"json\" [json, xlsx]")
	convertCmd.Flags().BoolVarP(&optNoHeader, "noheader", "z", false, "If true then it will not treat first line as header")
	convertCmd.Flags().BoolVarP(&optStdout, "stdout", "v", false, "Output to STDOUT")
	convertCmd.Flags().BoolVarP(&optStrict, "strict", "s", false, "don't use LazyQuotes mode")
	convertCmd.Flags().StringVarP(&optComma, "delimiter", "d", "", "string of delimiter")
	convertCmd.Flags().StringVarP(&optDecoder, "decoder", "i", "utf-8", "decorder for reader. default is utf-8 [utf-8, shift_jis, cp932]")
}

var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "convert from CSV",
	Long:  `convert from CSV`,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		convert(args)
	},
}

func convert(args []string) {
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

	if optTargetFormat == "json" {
		toJSON(reader, basename+".json")
	}
}

func toJSON(reader *csv.Reader, jsonPath string) {

	firstline, err := reader.Read()
	if err == io.EOF {
		return
	}
	if err != nil {
		er(err)
	}

	records := make([]map[string]string, 0)
	if optNoHeader {
		dummykeys := make([]string, len(firstline), len(firstline))
		for i, _ := range dummykeys {
			dummykeys[i] = fmt.Sprintf("field%d", i)
		}
		vmap := make(map[string]string)
		for i, key := range dummykeys {
			vmap[key] = firstline[i]
		}
		records = append(records, vmap)
		firstline = dummykeys
	} else {
		for i, value := range firstline {
			if len(value) == 0 {
				firstline[i] = fmt.Sprintf("field%d", i)
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
		vmap := make(map[string]string)
		for i, key := range firstline {
			vmap[key] = record[i]
		}
		records = append(records, vmap)
	}

	var encoder *json.Encoder
	if optStdout {
		encoder = json.NewEncoder(os.Stdout)
	} else {
		osWriter, err := os.Create(jsonPath)
		if err != nil {
			er(err)
		}
		defer osWriter.Close()
		encoder = json.NewEncoder(osWriter)
	}

	encoder.Encode(records)
}
