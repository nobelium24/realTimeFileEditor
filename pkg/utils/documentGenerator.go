package utils

import (
	"bytes"
	"encoding/json"
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

type ContentNode struct {
	Type    string        `json:"type"`
	Text    string        `json:"text,omitempty"`
	Content []ContentNode `json:"content,omitempty"`
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

	pdf.SetMargins(marginLeft, marginTop, marginRight)
	pdf.SetAutoPageBreak(true, marginBottom)

	pdf.AddUTF8Font(fontName, "", fontFile)
	pdf.SetFont(fontName, "", fontSize)

	pdf.SetFont(fontName, "B", fontSize+2)
	pdf.CellFormat(0, 10, document.Title, "", 1, "C", false, 0, "")
	pdf.Ln(5)
	pdf.SetFont(fontName, "", fontSize)

	if document.Content != nil {
		var contentNodes []ContentNode
		if err := json.Unmarshal(*document.Content, &contentNodes); err != nil {
			return nil, fmt.Errorf("failed to parse document content: %v", err)
		}
		p.renderContent(pdf, contentNodes, fontSize, lineSpacing)
	}

	pdf.SetFont(fontName, "I", 8)
	pdf.SetXY(0, 280)
	pdf.CellFormat(0, 5, fmt.Sprintf("Generated on %s", time.Now().UTC().Format("January 2, 2006")), "", 1, "C", false, 0, "")

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	return buf.Bytes(), err
}

func (p *PDFService) renderContent(pdf *gofpdf.Fpdf, nodes []ContentNode, fontSize, lineSpacing float64) {
	for _, node := range nodes {
		switch node.Type {
		case "text":
			pdf.MultiCell(0, fontSize*lineSpacing, node.Text, "", "L", false)
			pdf.Ln(-1)
		case "paragraph":
			p.renderContent(pdf, node.Content, fontSize, lineSpacing)
			pdf.Ln(fontSize / 2)
		case "heading":
			level := 1
			if strings.HasPrefix(node.Type, "heading") {
				level = int(node.Type[len(node.Type)-1] - '0')
			}
			headingSize := fontSize + float64(4-level)
			pdf.SetFont("", "B", headingSize)
			p.renderContent(pdf, node.Content, headingSize, lineSpacing)
			pdf.SetFont("", "", fontSize)
			pdf.Ln(fontSize / 2)
		case "bulletList", "orderedList":
			p.renderList(pdf, node.Content, fontSize, lineSpacing, node.Type == "orderedList")
		default:
			p.renderContent(pdf, node.Content, fontSize, lineSpacing)
		}
	}
}

func (p *PDFService) renderList(pdf *gofpdf.Fpdf, items []ContentNode, fontSize, lineSpacing float64, ordered bool) {
	for i, item := range items {
		if len(item.Content) > 0 {
			marker := "• "
			if ordered {
				marker = fmt.Sprintf("%d. ", i+1)
			}
			pdf.CellFormat(10, fontSize*lineSpacing, marker, "", 0, "", false, 0, "")

			x, y := pdf.GetXY()
			pdf.SetXY(x+10, y)
			p.renderContent(pdf, item.Content, fontSize, lineSpacing)
			pdf.SetXY(x, y+fontSize*lineSpacing)
		}
	}
}
