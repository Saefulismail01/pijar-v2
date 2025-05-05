package repository

type Topic struct {
	ID         int
	UserID     int
	Preference string // ini satu-satunya field yang akan dijadikan topik
}

// Mock topik sesuai user_id
func GetMockTopicsByUserID(userID int) []Topic {
	return []Topic{
		{ID: 1, UserID: userID, Preference: "AI untuk Startup"},
		{ID: 2, UserID: userID, Preference: "Strategi Pemasaran Digital"},
	}
}
