package model_util

type GeneratedArticle struct {
	Title   string // Judul artikel
	Content string // Isi lengkap, bisa multiline
	Source  string // Referensi atau sumber artikel
	TopicID int    // Relasi ke topik yang menghasilkan artikel
}
