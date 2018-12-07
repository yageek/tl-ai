package main

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"flag"
	"html/template"
	"os"

	"github.com/yageek/tl-ai/dataprovider"
)

var (
	output      string
	packageName string
	tmpl        *template.Template
)

func init() {
	flag.StringVar(&output, "output", "", "output path")
	flag.StringVar(&packageName, "package", "", "package")

	tm := `package {{.PackageName}}

	var embeddedGOB string = "{{.Data}}"
	`
	tmpl = template.Must(template.New("tmpl").Parse(tm))
}

func main() {

	flag.Parse()

	if output == "" || packageName == "" {
		flag.Usage()
		return
	}

	data, err := dataprovider.GetAPIData()
	if err != nil {
		panic(err)
	}

	file, err := os.Create(output)
	if err != nil {
		panic(err)
	}

	buff := new(bytes.Buffer)
	if err := gob.NewEncoder(buff).Encode(&data); err != nil {
		panic(err)
	}

	val := struct {
		PackageName string
		Data        string
	}{
		PackageName: packageName,
		Data:        hex.EncodeToString(buff.Bytes()),
	}

	if err := tmpl.Execute(file, val); err != nil {
		panic(err)
	}

}
