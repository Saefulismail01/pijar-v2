package usecase

import (
	"fmt"
	"log"
	"pijar/repository"
	"pijar/utils/service"
	"time"
)

type NotificationUseCase struct {
	NotifRepo repository.NotifRepoInterface
	GoalRepo  repository.DailyGoalRepository
	FCMClient *service.FCMClient
}

func NewNotificationUseCase(
	notifRepo repository.NotifRepoInterface,
	goalRepo repository.DailyGoalRepository,
	fcmClient *service.FCMClient,
) *NotificationUseCase {
	return &NotificationUseCase{
		NotifRepo: notifRepo,
		GoalRepo:  goalRepo,
		FCMClient: fcmClient,
	}
}

func (uc *NotificationUseCase) SendScheduledReminders() {
	users, err := uc.GoalRepo.GetAllUsersWithPendingArticles()
	if err != nil {
		log.Printf("Error fetching users: %v", err)
		return
	}

	for _, user := range users {
		count, err := uc.GoalRepo.GetPendingArticlesCount(user.ID)
		if err != nil || count == 0 {
			continue
		}

		message := uc.generateMessage(user.Name, count, time.Now().Hour())
		tokens, _ := uc.NotifRepo.GetDeviceTokens(user.ID)

		if len(tokens) > 0 {
			if err := uc.FCMClient.SendNotification(tokens, "📚 Reading Reminder", message); err != nil {
				log.Printf("Failed to send notification to user %d: %v", user.ID, err)
			}
		}
	}
}

func (uc *NotificationUseCase) generateMessage(name string, count int, hour int) string {
	var template string
	switch {
	case hour >= 9 && hour < 13:
		template = "Good morning %s! 🌞 You have %d article%s to read today. Let's start learning!"
	case hour >= 13 && hour < 17:
		template = "Hi %s! ☀️ Midday reminder: %d article%s left. Keep up the good work!"
	case hour >= 17 && hour < 20:
		template = "Good evening %s! 🌇 %d article%s remaining. Finish strong!"
	default:
		template = "Hi %s! 🌙 Don't forget your %d article%s before bed"
	}

	plural := "s"
	if count == 1 {
		plural = ""
	}
	return fmt.Sprintf(template, name, count, plural)
}
