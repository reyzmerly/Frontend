package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/signintech/gopdf"
)

type pdfFontPair struct {
	regular string
	bold    string
}

func AddSystemPDFFonts(pdf *gopdf.GoPdf) error {
	fontPairs := systemPDFFontPairs()
	var attempted []string

	for _, pair := range fontPairs {
		attempted = append(attempted, pair.regular, pair.bold)
		if _, err := os.Stat(pair.regular); err != nil {
			continue
		}
		if _, err := os.Stat(pair.bold); err != nil {
			continue
		}
		if err := pdf.AddTTFFont("regular", pair.regular); err != nil {
			return fmt.Errorf("load regular PDF font %q: %w", pair.regular, err)
		}
		if err := pdf.AddTTFFont("bold", pair.bold); err != nil {
			return fmt.Errorf("load bold PDF font %q: %w", pair.bold, err)
		}
		return nil
	}

	return fmt.Errorf("no supported regular/bold TTF font pair found; checked: %s", strings.Join(attempted, ", "))
}

func systemPDFFontPairs() []pdfFontPair {
	switch runtime.GOOS {
	case "windows":
		windowsDir := os.Getenv("WINDIR")
		if windowsDir == "" {
			windowsDir = os.Getenv("SystemRoot")
		}
		if windowsDir != "" {
			fontsDir := filepath.Join(windowsDir, "Fonts")
			return []pdfFontPair{
				{filepath.Join(fontsDir, "arial.ttf"), filepath.Join(fontsDir, "arialbd.ttf")},
				{filepath.Join(fontsDir, "segoeui.ttf"), filepath.Join(fontsDir, "segoeuib.ttf")},
			}
		}
	case "darwin":
		return []pdfFontPair{
			{"/System/Library/Fonts/Supplemental/Arial.ttf", "/System/Library/Fonts/Supplemental/Arial Bold.ttf"},
			{"/Library/Fonts/Arial.ttf", "/Library/Fonts/Arial Bold.ttf"},
		}
	default:
		return []pdfFontPair{
			{"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf", "/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf"},
			{"/usr/share/fonts/truetype/liberation2/LiberationSans-Regular.ttf", "/usr/share/fonts/truetype/liberation2/LiberationSans-Bold.ttf"},
			{"/usr/share/fonts/truetype/noto/NotoSans-Regular.ttf", "/usr/share/fonts/truetype/noto/NotoSans-Bold.ttf"},
		}
	}

	return nil
}
