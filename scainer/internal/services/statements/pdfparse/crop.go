package pdfparse

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

func CropPages(srcPDF, dstPDF string, pr PageRange) error {
	if pr.StartPage < 1 || pr.EndPage < pr.StartPage {
		return fmt.Errorf("pdfparse: invalid page range %d-%d", pr.StartPage, pr.EndPage)
	}
	if err := os.MkdirAll(filepath.Dir(dstPDF), 0o755); err != nil {
		return err
	}
	selector := fmt.Sprintf("%d-%d", pr.StartPage, pr.EndPage)
	conf := model.NewDefaultConfiguration()
	conf.ValidationMode = model.ValidationRelaxed
	return api.TrimFile(srcPDF, dstPDF, []string{selector}, conf)
}
