package store

import (
	"sync"

	"game-backend/models"
)

type MockStore struct {
	mu         sync.Mutex
	items      []models.Item
	user       models.User
	ownedItems map[string][]models.Item
}

func NewMockStore() *MockStore {
	return &MockStore{
		items: []models.Item{
			{
				ID:          "skin_red",
				Name:        "Red Skin",
				Description: "A bold red color skin for your character.",
				PriceCents:  199,
			},
			{
				ID:          "skin_blue",
				Name:        "Blue Skin",
				Description: "A cool blue color skin for your character.",
				PriceCents:  199,
			},
			{
				ID:          "skin_green",
				Name:        "Green Skin",
				Description: "A fresh green color skin for your character.",
				PriceCents:  199,
			},
			{
				ID:          "skin_purple",
				Name:        "Purple Skin",
				Description: "A vibrant purple color skin for your character.",
				PriceCents:  299,
			},
			{
				ID:          "skin_gold",
				Name:        "Gold Skin",
				Description: "A premium gold color skin for your character.",
				PriceCents:  499,
			},
		},
		user: models.User{
			ID:       "user_001",
			Username: "player_one",
			Crystals: 250,
		},
		ownedItems: map[string][]models.Item{
			"demo_player": {
				{
					ID:          "skin_blue",
					Name:        "Blue Skin",
					Description: "A cool blue color skin for your character.",
					PriceCents:  199,
				},
				{
					ID:          "skin_purple",
					Name:        "Purple Skin",
					Description: "A vibrant purple color skin for your character.",
					PriceCents:  299,
				},
			},
		},
	}
}

func (s *MockStore) Items() []models.Item {
	return s.items
}

func (s *MockStore) User() models.User {
	return s.user
}

func (s *MockStore) OwnedItems(playerID string) []models.Item {
	s.mu.Lock()
	defer s.mu.Unlock()

	items := s.ownedItems[playerID]
	result := make([]models.Item, len(items))
	copy(result, items)
	return result
}

func (s *MockStore) GrantItem(playerID string, item models.Item) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, ownedItem := range s.ownedItems[playerID] {
		if ownedItem.ID == item.ID {
			return
		}
	}

	s.ownedItems[playerID] = append(s.ownedItems[playerID], item)
}
