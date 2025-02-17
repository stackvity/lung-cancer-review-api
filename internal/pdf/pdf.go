package pdf

import "context"

// PDFGenerator defines the interface for PDF generation services.
type PDFGenerator interface {
	GeneratePDF(ctx context.Context, data interface{}) (string, error) // Placeholder method
}

// MockPDFGenerator is a mock implementation of the PDFGenerator interface for testing.
type MockPDFGenerator struct{}

func NewMockPDFGenerator() *MockPDFGenerator { // Added for creating mock pdf generator.
	return &MockPDFGenerator{}
}

func (mkb *MockPDFGenerator) GeneratePDF(ctx context.Context, data interface{}) (string, error) {
	// TODO: Implement mock logic for testing.
	// For now, return a placeholder value and no error.
	return "/tmp/report_mock.pdf", nil // Placeholder return
}
