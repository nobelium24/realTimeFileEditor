package utils

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"realTimeEditor/internal/model"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"
)

type PDFService struct {
	assetsPath string
}

func NewPDFService(assetsPath string) *PDFService {
	return &PDFService{
		assetsPath: assetsPath,
	}
}

func (p *PDFService) VerifyFonts() error {
	defaultFonts := []string{
		"TimesNewRoman-Regular.ttf",
	}

	optionalFonts := []string{
		"Roboto-Regular.ttf",
		"Arial.ttf",
		"Georgia.ttf",
	}

	allFonts := append(defaultFonts, optionalFonts...)

	for _, font := range allFonts {
		path := filepath.Join(p.assetsPath, "fonts", font)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if strings.Contains(font, "TimesNewRoman") {
				return fmt.Errorf("required default font missing: %s", font)
			}
		}
	}
	return nil
}

func (p *PDFService) GenerateDocumentPDF(document *model.Document, m *model.DocumentMetadata) ([]byte, error) {
	if err := p.VerifyFonts(); err != nil {
		return nil, err
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 15, 15)
	pdf.AddPage()

	fontPath := filepath.Join(p.assetsPath, "fonts")

	// Fallbacks
	fontName := "times"
	fontFile := filepath.Join(fontPath, "TimesNewRoman-Regular.ttf")
	fontSize := 12.0
	lineSpacing := 1.5
	marginTop := 15.0
	marginLeft := 15.0
	marginRight := 15.0
	marginBottom := 15.0

	// Override if metadata is present
	if m != nil {
		if m.Metadata.Font != "" {
			fontName = strings.ToLower(strings.ReplaceAll(m.Metadata.Font, " ", "")) // Normalize
			customFontFile := filepath.Join(fontPath, m.Metadata.Font+".ttf")
			if _, err := os.Stat(customFontFile); err == nil {
				pdf.AddUTF8Font(fontName, "", customFontFile)
				fontFile = customFontFile
			}
		}
		if m.Metadata.FontSize != 0 {
			fontSize = m.Metadata.FontSize
		}
		if m.Metadata.LineSpacing != 0 {
			lineSpacing = m.Metadata.LineSpacing
		}
		if m.Metadata.MarginTop != 0 {
			marginTop = m.Metadata.MarginTop
		}
		if m.Metadata.MarginLeft != 0 {
			marginLeft = m.Metadata.MarginLeft
		}
		if m.Metadata.MarginRight != 0 {
			marginRight = m.Metadata.MarginRight
		}
		if m.Metadata.MarginBottom != 0 {
			marginBottom = m.Metadata.MarginBottom
		}
	}

	// Apply margins
	pdf.SetMargins(marginLeft, marginTop, marginRight)
	pdf.SetAutoPageBreak(true, marginBottom)

	// Register fallback font if not overridden
	pdf.AddUTF8Font(fontName, "", fontFile)
	pdf.SetFont(fontName, "", fontSize)

	// Title
	pdf.SetFont(fontName, "B", fontSize+2)
	pdf.CellFormat(0, 10, document.Title, "", 1, "C", false, 0, "")
	pdf.Ln(5)
	pdf.SetFont(fontName, "", fontSize)

	// Content (very basic word wrap)
	if document.Content != nil {
		lines := strings.Split(*document.Content, "\n")
		for _, line := range lines {
			pdf.MultiCell(0, fontSize*lineSpacing, line, "", "L", false)
			pdf.Ln(-1)
		}
	}

	// Footer
	pdf.SetFont(fontName, "I", 8)
	pdf.SetXY(0, 280)
	pdf.CellFormat(0, 5, fmt.Sprintf("Generated on %s", time.Now().Format("January 2, 2006")), "", 1, "C", false, 0, "")

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	return buf.Bytes(), err
}
