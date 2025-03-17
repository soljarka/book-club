package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/kelseyhightower/envconfig"
	"github.com/soljarka/book-club/config"
	"github.com/soljarka/book-club/hosts"
)

var bookClub *hosts.Bookclub

func main() {
	var c config.Config
	err := envconfig.Process("bookclub", &c)
	if err != nil {
		log.Fatal(err.Error())
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	opts := []bot.Option{}

	b, err := bot.New(c.BotToken, opts...)
	if err != nil {
		panic(err)
	}

	b.RegisterHandlerMatchFunc(startHandlerMatchFunc, startHandler)
	b.RegisterHandlerMatchFunc(helpHandlerMatchFunc, helpHandler)
	b.RegisterHandlerMatchFunc(nextSessionMatchFunc, nextSessionHandler)
	b.RegisterHandlerMatchFunc(nextNthSessionMatchFunc, nextNthSessionHandler)
	b.RegisterHandlerMatchFunc(listHostsHandlerMatchFunc, listHostsHandler)
	b.RegisterHandlerMatchFunc(listBooksHandlerMatchFunc, listBooksHandler)
	b.RegisterHandlerMatchFunc(registerHostMatchFunc, registerHostHandler)
	b.RegisterHandlerMatchFunc(deregisterHostMatchFunc, deregisterHostHandler)
	b.RegisterHandlerMatchFunc(registerBookHandlerMatchFunc, registerBookHandler)
	b.RegisterHandlerMatchFunc(deleteBookHandlerMatchFunc, deleteBookHandler)
	b.RegisterHandlerMatchFunc(setNextBookHandlerMatchFunc, setNextBookHandler)
	b.RegisterHandlerMatchFunc(setQueueHandlerMatchFunc, setQueueHandler)
	b.RegisterHandlerMatchFunc(getQueueHandlerMatchFunc, getQueueHandler)

	b.Start(ctx)
}

func helpHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text: `Доступные команды:
/help - Показать список доступных команд.
/start - Начать работу с ботом.
/next - Показать информацию о следующем собрании клуба.
/next <n> - Показать информацию о n-ом собрании клуба.
/register - Зарегистрировать себя как участника.
/deregister - Удалить себя из списка участников.
/book <author/title> - Добавить книгу в список.
/list_hosts - Показать пронумерованный список всех участников.
/list_books - Показать пронумерованный список всех книг.
/delete_book <номер книги> - Удалить книгу из списка.
/my_next_book <номер книги> - Установить следующую книгу для чтения.
/set_queue <номера участников через запятую> - Установить очередь ведущих.
/get_queue - Показать очередь ведущих.`,
	})
}

func startHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	club, err := hosts.LoadBookclub(update.Message.Chat.ID)
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Не удалось загрузить данные.",
		})
		return
	}

	bookClub = club

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Привет! Я бот книжного клуба.",
	})
}

func registerHostHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	bookClub.AddHost(update.Message.From.ID, update.Message.From.Username)

	err := bookClub.Save()
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("He удалось сохранить."),
		})
		return
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Готово!"),
	})
}

func deregisterHostHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	bookClub.RemoveHost(update.Message.From.ID)

	err := bookClub.Save()
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("He удалось сохранить."),
		})
		return
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Готово!"),
	})
}

func registerBookHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	bookDescription := strings.TrimPrefix(update.Message.Text, "/book ")
	bookDetails := strings.SplitN(bookDescription, "/", 2)
	if len(bookDetails) != 2 {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Неверный формат. Пожалуйста, используйте формат 'author/title'.",
		})
		return
	}

	author := strings.TrimSpace(bookDetails[0])
	title := strings.TrimSpace(bookDetails[1])

	bookClub.AddBook(author, title)

	err := bookClub.Save()
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("He удалось сохранить."),
		})
		return
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   fmt.Sprintf("Готово!"),
	})
}

func listHostsHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	message := "Участники:"

	counter := 1
	for _, host := range bookClub.GetHosts() {
		message += fmt.Sprintf("\n%d. %s", counter, host.Username)
		counter++
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   message,
	})
}

func listBooksHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	message := "Книги:"

	counter := 1
	for _, book := range bookClub.GetBooks() {
		message += fmt.Sprintf("\n%d. %s - %s", counter, book.Author, book.Title)
		counter++
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   message,
	})
}

func deleteBookHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	bookIndex, err := strconv.Atoi(strings.TrimPrefix(update.Message.Text, "/delete_book "))
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Неверный формат. Пожалуйста, используйте формат '/delete_book <номер книги>'.",
		})
		return
	}

	bookId, err := bookIndexToId(bookIndex)
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Книга не найдена.",
		})
		return
	}

	err = bookClub.DeleteBook(bookId)

	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Книга не найдена.",
		})
		return
	}

	err = bookClub.Save()
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("He удалось сохранить."),
		})
		return
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Готово!",
	})
}

func setNextBookHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	bookIndex, err := strconv.Atoi(strings.TrimPrefix(update.Message.Text, "/my_next_book "))
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Неверный формат. Пожалуйста, используйте формат '/my_next_book <номер книги>'.",
		})
		return
	}

	bookId, err := bookIndexToId(bookIndex)
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Книга не найдена.",
		})
		return
	}

	bookClub.SetNextBook(update.Message.From.ID, bookId)

	err = bookClub.Save()
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("He удалось сохранить."),
		})
		return
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Готово!",
	})
}

func bookIndexToId(bookIndex int) (string, error) {
	count := 0
	for _, book := range bookClub.GetBooks() {
		if count == bookIndex-1 {
			return book.Id, nil
		}
		count++
	}
	return "", errors.New("Книга не найдена.")
}

func hostIndexToId(hostIndex int) (int64, error) {
	count := 0
	for _, host := range bookClub.GetHosts() {
		if count == hostIndex-1 {
			return host.TelegramId, nil
		}
		count++
	}
	return 0, errors.New("Участник не найден.")
}

func nextNthSessionHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	n, err := strconv.Atoi(strings.TrimPrefix(update.Message.Text, "/next "))

	if err != nil {
		return
	}

	session, err := bookClub.GetNthSession(n)
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "В очередь не добавлен ни один участник.",
		})
		return
	}

	nextBookId := session.Host.NextBookId
	book := bookClub.GetBook(nextBookId)
	message := fmt.Sprintf("Ближайшее %d-е собрание клуба: %s, ведущий: %s", n, session.Date.Format("01/02/2006"), session.Host.Username)
	if nextBookId != "" {
		message += fmt.Sprintf(", книга: %s - %s", book.Title, book.Author)
	}
	message += "."

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   message,
	})
}

func setQueueHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	hostIndexes := strings.Split(strings.TrimPrefix(update.Message.Text, "/set_queue "), ",")
	telegramIds := []int64{}

	for _, hostIndex := range hostIndexes {
		index, err := strconv.Atoi(strings.TrimSpace(hostIndex))
		if err != nil {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "Неверный формат. Пожалуйста, используйте формат '/set_queue <номера участников через запятую>'.",
			})
			return
		}

		telegramId, err := hostIndexToId(index)
		if err != nil {
			b.SendMessage(ctx, &bot.SendMessageParams{
				ChatID: update.Message.Chat.ID,
				Text:   "Участник не найден.",
			})
			return
		}

		telegramIds = append(telegramIds, telegramId)
	}

	err := bookClub.SetQueue(telegramIds)

	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Один или несколько участников не найдены.",
		})
		return
	}

	err = bookClub.Save()
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   fmt.Sprintf("He удалось сохранить."),
		})
		return
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   "Готово!",
	})
}

func nextSessionHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	session, err := bookClub.GetNthSession(1)
	if err != nil {
		b.SendMessage(ctx, &bot.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "В очередь не добавлен ни один участник.",
		})
		return
	}

	nextBookId := session.Host.NextBookId
	message := fmt.Sprintf("Cледующее собрание клуба: %s, ведущий: %s", session.Date.Format("01/02/2006"), session.Host.Username)
	if nextBookId != "" {
		message += fmt.Sprintf(", книга: %s", bookClub.GetBook(nextBookId).Title)
	}
	message += "."

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   message,
	})
}

func getQueueHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	message := "Очередь ведущих:"

	hosts := bookClub.GetHosts()

	for _, telegramId := range bookClub.GetQueue() {
		message += fmt.Sprintf("\n%s", hosts[telegramId].Username)
	}

	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   message,
	})
}

func startHandlerMatchFunc(update *models.Update) bool {
	if update.Message == nil {
		return false
	}
	return update.Message.Text == "/start"
}

func helpHandlerMatchFunc(update *models.Update) bool {
	if update.Message == nil {
		return false
	}
	return update.Message.Text == "/help"
}

func nextSessionMatchFunc(update *models.Update) bool {
	if update.Message == nil {
		return false
	}
	return update.Message.Text == "/next"
}

func nextNthSessionMatchFunc(update *models.Update) bool {
	if update.Message == nil {
		return false
	}
	return strings.HasPrefix(update.Message.Text, "/next ")
}

func registerHostMatchFunc(update *models.Update) bool {
	if update.Message == nil {
		return false
	}
	return strings.HasPrefix(update.Message.Text, "/register")
}

func deregisterHostMatchFunc(update *models.Update) bool {
	if update.Message == nil {
		return false
	}
	return strings.HasPrefix(update.Message.Text, "/deregister")
}

func registerBookHandlerMatchFunc(update *models.Update) bool {
	if update.Message == nil {
		return false
	}
	return strings.HasPrefix(update.Message.Text, "/book")
}

func deleteBookHandlerMatchFunc(update *models.Update) bool {
	if update.Message == nil {
		return false
	}
	return strings.HasPrefix(update.Message.Text, "/delete_book")
}

func setNextBookHandlerMatchFunc(update *models.Update) bool {
	if update.Message == nil {
		return false
	}
	return strings.HasPrefix(update.Message.Text, "/my_next_book")
}

func listHostsHandlerMatchFunc(update *models.Update) bool {
	if update.Message == nil {
		return false
	}
	return update.Message.Text == "/list_hosts"
}

func listBooksHandlerMatchFunc(update *models.Update) bool {
	if update.Message == nil {
		return false
	}
	return update.Message.Text == "/list_books"
}

func setQueueHandlerMatchFunc(update *models.Update) bool {
	if update.Message == nil {
		return false
	}
	return strings.HasPrefix(update.Message.Text, "/set_queue")
}

func getQueueHandlerMatchFunc(update *models.Update) bool {
	if update.Message == nil {
		return false
	}
	return update.Message.Text == "/get_queue"
}
