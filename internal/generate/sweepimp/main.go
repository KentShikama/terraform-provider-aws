//go:build generate
// +build generate

package main

import (
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-provider-aws/internal/generate/common"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	filename      = `../../sweep/sweep_test.go`
	namesDataFile = "../../../names/names_data.csv"
)

type ServiceDatum struct {
	ProviderPackage string
}

type TemplateData struct {
	Services []ServiceDatum
}

func main() {
	g := common.NewGenerator()

	g.Infof("Generating %s", strings.TrimPrefix(filename, "../../"))

	data, err := common.ReadAllCSVData(namesDataFile)

	if err != nil {
		g.Fatalf("error reading %s: %s", namesDataFile, err.Error())
	}

	td := TemplateData{}

	for i, l := range data {
		if i < 1 { // no header
			continue
		}

		if l[names.ColExclude] != "" {
			continue
		}

		if l[names.ColProviderPackageActual] == "" && l[names.ColProviderPackageCorrect] == "" {
			continue
		}

		p := l[names.ColProviderPackageCorrect]

		if l[names.ColProviderPackageActual] != "" {
			p = l[names.ColProviderPackageActual]
		}

		if _, err := os.Stat(fmt.Sprintf("../../service/%s", p)); err != nil || errors.Is(err, fs.ErrNotExist) {
			continue
		}

		if _, err := os.Stat(fmt.Sprintf("../../service/%s/sweep.go", p)); err != nil || errors.Is(err, fs.ErrNotExist) {
			continue
		}

		s := ServiceDatum{
			ProviderPackage: p,
		}

		td.Services = append(td.Services, s)
	}

	sort.SliceStable(td.Services, func(i, j int) bool {
		return td.Services[i].ProviderPackage < td.Services[j].ProviderPackage
	})

	if err := g.ApplyAndWriteTemplateGoFormat(filename, "sweepimport", tmpl, td); err != nil {
		g.Fatalf("error: %s", err.Error())
	}
}

//go:embed file.tmpl
var tmpl string
