package bot

import (
	"context"
	"log"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"pikobot/internal/database"
	"pikobot/internal/mis"
)

type Bot struct {
	API           *tgbotapi.BotAPI
	UserRepo      *database.UserRepository
	TxRepo        *database.TransactionRepository
	BannerURL     string
	MISClient     *mis.Client
	userStates    map[int64]string
	stateMutex    sync.RWMutex
	logbookDrafts map[int64]*mis.LogbookEntry
	draftMutex    sync.RWMutex
}

func New(token string, userRepo *database.UserRepository, txRepo *database.TransactionRepository, bannerURL string) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	return &Bot{
		API:           api,
		UserRepo:      userRepo,
		TxRepo:        txRepo,
		BannerURL:     bannerURL,
		userStates:    make(map[int64]string),
		logbookDrafts: make(map[int64]*mis.LogbookEntry),
	}, nil
}

func (b *Bot) SetMISClient(client *mis.Client) {
	b.MISClient = client
}

func (b *Bot) SetLogbookDraft(userID int64, draft *mis.LogbookEntry) {
	b.draftMutex.Lock()
	defer b.draftMutex.Unlock()
	b.logbookDrafts[userID] = draft
}

func (b *Bot) GetLogbookDraft(userID int64) *mis.LogbookEntry {
	b.draftMutex.RLock()
	defer b.draftMutex.RUnlock()
	return b.logbookDrafts[userID]
}

func (b *Bot) ClearLogbookDraft(userID int64) {
	b.draftMutex.Lock()
	defer b.draftMutex.Unlock()
	delete(b.logbookDrafts, userID)
}

func (b *Bot) SetUserState(userID int64, state string) {
	b.stateMutex.Lock()
	defer b.stateMutex.Unlock()
	b.userStates[userID] = state
}

func (b *Bot) GetUserState(userID int64) string {
	b.stateMutex.RLock()
	defer b.stateMutex.RUnlock()
	return b.userStates[userID]
}

func (b *Bot) ClearUserState(userID int64) {
	b.stateMutex.Lock()
	defer b.stateMutex.Unlock()
	delete(b.userStates, userID)
}

func (b *Bot) ProcessUpdate(update tgbotapi.Update) {

	if update.CallbackQuery != nil {
		if b.UserRepo != nil && update.CallbackQuery.From != nil {
			from := update.CallbackQuery.From
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()

				u := &database.User{
					TelegramID: from.ID,
					Username:   from.UserName,
					FirstName:  from.FirstName,
					LastName:   from.LastName,
				}
				if err := b.UserRepo.UpsertUser(ctx, u); err != nil {
					log.Printf("[WARN] Failed to upsert user on callback: %v", err)
				}
			}()
		}

		if err := b.HandleCallbackQuery(update.CallbackQuery); err != nil {
			log.Printf("[ERROR] Failed to handle callback query: %v", err)
		}
		return
	}

	if update.Message == nil {
		return
	}

	if b.UserRepo != nil && update.Message.From != nil {
		from := update.Message.From
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			u := &database.User{
				TelegramID: from.ID,
				Username:   from.UserName,
				FirstName:  from.FirstName,
				LastName:   from.LastName,
			}
			if err := b.UserRepo.UpsertUser(ctx, u); err != nil {
				log.Printf("[WARN] Failed to upsert user in MongoDB: %v", err)
			}
		}()
	}

	var err error
	if update.Message.IsCommand() {
		err = b.HandleCommand(update.Message)
	} else if update.Message.Document != nil {
		err = b.HandleDocument(update.Message)
	} else if len(update.Message.Photo) > 0 {
		err = b.HandlePhoto(update.Message)
	} else if update.Message.Text != "" {
		err = b.HandleMessage(update.Message)
	}

	if err != nil {
		log.Printf("[ERROR] Failed to handle update %d: %v", update.UpdateID, err)
	}
}
