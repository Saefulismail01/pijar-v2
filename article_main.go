package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"

	"pijar/repository"
	"pijar/utils"
)

func main() {
	// Sementara hardcode user ID
	userID := 1

	// Cek apakah API key tersedia
	// Load file .env
	if err := godotenv.Load(); err != nil {
		log.Fatal("Gagal memuat file .env:", err)
	}

	// Cek apakah key tersedia
	if os.Getenv("DEEPSEEK_API_KEY") == "" {
		log.Fatal("DEEPSEEK_API_KEY belum diatur di environment")
	}

	// Ambil topik-topik mock untuk user ini
	topics := repository.GetMockTopicsByUserID(userID)

	for _, topic := range topics {
		fmt.Println("🔍 Memproses topik:", topic.Preference)

		article, err := utils.GenerateArticleFromDeepseek(topic.Preference, topic.Preference, topic.ID)
		if err != nil {
			log.Printf("❌ Gagal generate artikel untuk topik '%s': %v\n", topic.Preference, err)
			continue
		}

		fmt.Println("\n✅ Artikel berhasil dibuat:")
		fmt.Println("Judul  :", article.Title)
		fmt.Println("Sumber :", article.Source)
		fmt.Println("Isi    :\n" + article.Content)
		fmt.Println("===")
	}
}
