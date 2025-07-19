package utils

import (
	"bytes"
	"fmt"
	"realTimeEditor/internal/model"
	"realTimeEditor/internal/repositories"
)

// DocumentHandler generates a PDF from a Document (with optional formatting metadata) and uploads it to Cloudinary.
func DocumentHandler(
	document *model.Document,
	metadata *model.DocumentMetadata,
) (*repositories.UploadedMedia, error) {
	pdfService := NewPDFService("assets/")
	byteSlice, err := pdfService.GenerateDocumentPDF(document, metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	reader := bytes.NewReader(byteSlice)
	fileName := fmt.Sprintf("document_%s.pdf", document.ID.String())

	result, err := repositories.CloudinaryUploaderStream(reader, fileName, repositories.RawResource)
	if err != nil {
		return nil, fmt.Errorf("cloudinary upload failed: %w", err)
	}

	return &result, nil
}
