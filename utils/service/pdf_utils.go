package service

import (
	"time"

	"github.com/jung-kurt/gofpdf"
	"pijar/model"
)

// JournalPDF represents the structure for PDF generation
type JournalPDF struct {
	Title     string
	Content   string
	Feeling   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ConvertToPDFFormat converts Journal models to PDF format
func ConvertToPDFFormat(journals []model.Journal) []JournalPDF {
	pdfJournals := make([]JournalPDF, len(journals))
	for i, journal := range journals {
		pdfJournals[i] = JournalPDF{
			Title:     journal.Judul,
			Content:   journal.Isi,
			Feeling:   journal.Perasaan,
			CreatedAt: journal.CreatedAt,
			UpdatedAt: journal.UpdatedAt,
		}
	}
	return pdfJournals
}

// GenerateJournalsPDF generates a PDF from journal entries
func GenerateJournalsPDF(journals []JournalPDF) (*gofpdf.Fpdf, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetFont("Arial", "", 12)
	pdf.AddPage()

	// Add title
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 6, "Jurnal Pribadi")
	pdf.Ln(8)
	
	// Add print date
	pdf.SetFont("Arial", "", 9)
	pdf.Cell(0, 6, "Dicetak pada: "+time.Now().Format("02 January 2006 15:04:05"))
	pdf.Ln(10)

	for _, journal := range journals {
		// Journal dates
		pdf.SetFont("Arial", "", 9)
		pdf.Cell(0, 5, "Dibuat pada: "+journal.CreatedAt.Format("2006-01-02 15:04:05"))
		pdf.Ln(5)
		
		// Tampilkan updated_at jika berbeda dengan created_at
		if !journal.UpdatedAt.IsZero() && journal.UpdatedAt != journal.CreatedAt {
			pdf.Cell(0, 5, "Diperbarui: "+journal.UpdatedAt.Format("2006-01-02 15:04:05"))
			pdf.Ln(5)
		}

		// Journal title
		pdf.SetFont("Arial", "B", 14)
		pdf.SetFillColor(230, 230, 250) // Light purple background
		pdf.CellFormat(190, 8, "[PENTING] "+journal.Title, "", 1, "L", true, 0, "")
		pdf.Ln(3)

		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(10, 6, "Isi Jurnal:")
		pdf.Ln(4)
		pdf.SetFont("Arial", "", 11)
		pdf.MultiCell(190, 5, "• "+journal.Content, "", "L", false)
		pdf.Ln(3)

		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(20, 5, "Perasaan:")
		pdf.SetFont("Arial", "", 11)
		pdf.Cell(0, 5, journal.Feeling)
		pdf.Ln(8)
	}

	return pdf, nil
}
