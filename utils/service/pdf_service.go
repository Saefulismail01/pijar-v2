package service

import (
	"strconv"
	"strings"
	"time"

	"pijar/model"

	"github.com/jung-kurt/gofpdf"
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
	// Inisialisasi PDF dengan dokumen A4
	pdf := gofpdf.New("P", "mm", "A4", "")
	
	// Atur margins
	pdf.SetMargins(15, 15, 15)
	pdf.SetAutoPageBreak(true, 15)
	
	// Tambahkan halaman pertama
	pdf.AddPage()
	
	// Definisi warna
	primaryColor := []int{41, 128, 185}    // Biru
	secondaryColor := []int{236, 240, 241} // Abu-abu muda
	accentColor := []int{52, 152, 219}     // Biru muda
	
	// ----- Header Dokumen -----
	// Judul utama
	pdf.SetFont("Arial", "B", 22)
	pdf.SetTextColor(primaryColor[0], primaryColor[1], primaryColor[2])
	pdf.Cell(0, 10, "Jurnal Pribadi")
	pdf.Ln(15)
	
	// Tanggal cetak
	pdf.SetFont("Arial", "I", 10)
	pdf.SetTextColor(100, 100, 100)
	pdf.Cell(0, 6, "Dicetak pada: "+time.Now().Format("02 January 2006 15:04:05"))
	pdf.Ln(15)
	
	// ----- Garis pemisah -----
	pdf.SetDrawColor(accentColor[0], accentColor[1], accentColor[2])
	pdf.SetLineWidth(0.5)
	pdf.Line(15, pdf.GetY(), 195, pdf.GetY())
	pdf.Ln(10)
	
	// Untuk setiap jurnal dalam koleksi
	for i, journal := range journals {
		// Cek apakah perlu halaman baru
		if i > 0 && pdf.GetY() > 250 {
			pdf.AddPage()
		}
		
		// ----- Header Jurnal -----
		// Judul jurnal dengan latar belakang
		pdf.SetFillColor(primaryColor[0], primaryColor[1], primaryColor[2])
		pdf.SetTextColor(255, 255, 255)
		pdf.SetFont("Arial", "B", 14)
		
		// Judul dengan background
		pdf.RoundedRect(15, pdf.GetY(), 180, 12, 3, "1234", "F")
		pdf.CellFormat(180, 12, "  "+journal.Title, "", 1, "L", false, 0, "")
		pdf.Ln(5)
		
		// Tanggal dibuat dan diperbarui
		pdf.SetFont("Arial", "", 9)
		pdf.SetTextColor(100, 100, 100)
		pdf.CellFormat(90, 5, "Dibuat: "+journal.CreatedAt.Format("02 Jan 2006 15:04"), "", 0, "L", false, 0, "")
		
		// Tampilkan updated_at jika berbeda dengan created_at
		if !journal.UpdatedAt.IsZero() && journal.UpdatedAt != journal.CreatedAt {
			pdf.CellFormat(90, 5, "Diperbarui: "+journal.UpdatedAt.Format("02 Jan 2006 15:04"), "", 0, "R", false, 0, "")
		}
		pdf.Ln(10)
		
		// ----- Isi Jurnal -----
		// Heading isi
		pdf.SetFont("Arial", "B", 12)
		pdf.SetTextColor(primaryColor[0], primaryColor[1], primaryColor[2])
		pdf.Cell(0, 6, "Isi Jurnal:")
		pdf.Ln(8)
		
		// Isi jurnal dengan formatting yang lebih baik
		pdf.SetFont("Arial", "", 11)
		pdf.SetTextColor(50, 50, 50)
		
		// Format teks isi agar lebih rapi
		pdf.SetX(15)
		textWidth := 180
		
		// Bersihkan whitespace berlebih dan format paragraf
		content := strings.TrimSpace(journal.Content)
		
		// Ganti multiple newlines dengan penanda paragraf
		content = strings.ReplaceAll(content, "\r\n", "\n")
		content = strings.ReplaceAll(content, "\n\n\n", "\n\n")
		
		// Pisahkan teks menjadi paragraf berdasarkan baris kosong
		paragraphs := strings.Split(content, "\n\n")
		for i, paragraph := range paragraphs {
			// Bersihkan whitespace di awal dan akhir paragraf
			paragraph = strings.TrimSpace(paragraph)
			
			// Ganti baris baru tunggal dengan spasi untuk menyambungkan kalimat dalam paragraf
			paragraph = strings.ReplaceAll(paragraph, "\n", " ")
			
			// Tambahkan indentasi paragraf dengan tab
			if len(paragraph) > 0 {
				// Cetak paragraf dengan rata kanan-kiri (justified)
				pdf.MultiCell(float64(textWidth), 5, paragraph, "", "J", false)
				
				// Tambahkan jarak antar paragraf kecuali untuk paragraf terakhir
				if i < len(paragraphs)-1 {
					pdf.Ln(3)
				}
			}
		}
		pdf.Ln(10)
		
		// ----- Perasaan -----
		// Box perasaan dengan latar belakang
		pdf.SetFillColor(secondaryColor[0], secondaryColor[1], secondaryColor[2])
		pdf.SetDrawColor(accentColor[0], accentColor[1], accentColor[2])
		pdf.SetLineWidth(0.2)
		pdf.RoundedRect(15, pdf.GetY(), 180, 10, 2, "1234", "FD")
		
		// Label perasaan
		pdf.SetFont("Arial", "B", 11)
		pdf.SetTextColor(primaryColor[0], primaryColor[1], primaryColor[2])
		pdf.SetX(20)
		pdf.CellFormat(25, 10, "Perasaan:", "", 0, "L", false, 0, "")
		
		// Isi perasaan
		pdf.SetFont("Arial", "", 11)
		pdf.SetTextColor(50, 50, 50)
		pdf.CellFormat(150, 10, journal.Feeling, "", 1, "L", false, 0, "")
		
		// Jarak antar entri jurnal
		pdf.Ln(15)
		
		// Garis pembatas antar jurnal (kecuali entri terakhir)
		if i < len(journals)-1 {
			pdf.SetDrawColor(200, 200, 200)
			pdf.SetLineWidth(0.2)
			pdf.Line(25, pdf.GetY()-5, 185, pdf.GetY()-5)
			pdf.SetDrawColor(accentColor[0], accentColor[1], accentColor[2])
			pdf.Ln(10)
		}
	}
	
	// Footer
	pdf.SetY(-15)
	pdf.SetFont("Arial", "I", 8)
	pdf.SetTextColor(100, 100, 100)
	pdf.CellFormat(0, 10, "Halaman "+strconv.Itoa(pdf.PageNo())+"/{nb}", "", 0, "C", false, 0, "")
	pdf.AliasNbPages("")
	
	return pdf, nil
}