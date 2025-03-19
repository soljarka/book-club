package main

import (
	"context"
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

func loadBookclub(ctx context.Context, b *bot.Bot, update *models.Update) *hosts.Bookclub {
	bookClub, err := hosts.LoadBookclub(update.Message.Chat.ID)
	if err != nil {
		handleError(ctx, b, update, "He удалось загрузить клуб.", err)
		return nil
	}
	return bookClub
}

func helpHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	sendMessage(ctx, b, update, `Доступные команды:
/help - Показать список доступных команд.
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
/get_queue - Показать очередь ведущих.`)
}

func registerHostHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	bookClub := loadBookclub(ctx, b, update)
	if bookClub == nil {
		return
	}

	bookClub.AddHost(update.Message.From.ID, update.Message.From.Username)

	err := bookClub.Save()
	if err != nil {
		handleError(ctx, b, update, "He удалось сохранить.", err)
		return
	}

	sendMessage(ctx, b, update, "Готово!")
}

func deregisterHostHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	bookClub := loadBookclub(ctx, b, update)
	if bookClub == nil {
		return
	}

	bookClub.RemoveHost(update.Message.From.ID)

	err := bookClub.Save()
	if err != nil {
		handleError(ctx, b, update, "He удалось сохранить.", err)
		return
	}

	sendMessage(ctx, b, update, "Готово!")
}

func registerBookHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	bookClub := loadBookclub(ctx, b, update)
	if bookClub == nil {
		return
	}

	bookDescription := strings.TrimPrefix(update.Message.Text, "/book ")
	bookDetails := strings.SplitN(bookDescription, "/", 2)
	if len(bookDetails) != 2 {
		sendMessage(ctx, b, update, "Неверный формат. Пожалуйста, используйте формат '/book <автор>/<название книги>'.")
		return
	}

	author := strings.TrimSpace(bookDetails[0])
	title := strings.TrimSpace(bookDetails[1])

	bookClub.AddBook(author, title)

	err := bookClub.Save()
	if err != nil {
		handleError(ctx, b, update, "He удалось сохранить.", err)
		return
	}

	sendMessage(ctx, b, update, "Готово!")
	listBooksHandler(ctx, b, update)
}

func listHostsHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	bookClub := loadBookclub(ctx, b, update)
	if bookClub == nil {
		return
	}

	message := "Участники:"

	for i, host := range bookClub.GetHosts() {
		message += fmt.Sprintf("\n%d. %s", i+1, host.Username)
	}

	sendMessage(ctx, b, update, message)
}

func listBooksHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	bookClub := loadBookclub(ctx, b, update)
	if bookClub == nil {
		return
	}

	message := "Книги:"

	for i, book := range bookClub.GetBooks() {
		message += fmt.Sprintf("\n%d. %s - «%s»", i+1, book.Author, book.Title)
	}

	sendMessage(ctx, b, update, message)
}

func deleteBookHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	bookClub := loadBookclub(ctx, b, update)
	if bookClub == nil {
		return
	}

	bookIndex, err := strconv.Atoi(strings.TrimPrefix(update.Message.Text, "/delete_book "))
	if err != nil {
		sendMessage(ctx, b, update, "Неверный формат. Пожалуйста, используйте формат '/delete_book <номер книги>'.")
		return
	}

	err = bookClub.DeleteBook(bookIndex - 1)

	if err != nil {
		sendMessage(ctx, b, update, "Книга не найдена.")
		return
	}

	err = bookClub.Save()
	if err != nil {
		handleError(ctx, b, update, "He удалось сохранить.", err)
		return
	}

	sendMessage(ctx, b, update, "Готово!")
}

func setNextBookHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	bookClub := loadBookclub(ctx, b, update)
	if bookClub == nil {
		return
	}

	bookIndex, err := strconv.Atoi(strings.TrimPrefix(update.Message.Text, "/my_next_book "))
	if err != nil {
		sendMessage(ctx, b, update, "Неверный формат. Пожалуйста, используйте формат '/my_next_book <номер книги>'.")
		return
	}

	err = bookClub.SetNextBook(update.Message.From.ID, bookIndex-1)
	if err != nil {
		sendMessage(ctx, b, update, "Книга не найдена.")
		return
	}

	err = bookClub.Save()
	if err != nil {
		handleError(ctx, b, update, "He удалось сохранить.", err)
		return
	}

	sendMessage(ctx, b, update, "Готово!")
}

func nextNthSessionHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	bookClub := loadBookclub(ctx, b, update)
	if bookClub == nil {
		return
	}

	n, err := strconv.Atoi(strings.TrimPrefix(update.Message.Text, "/next "))

	if err != nil {
		return
	}

	session, err := bookClub.GetNthSession(n)
	if err != nil {
		sendMessage(ctx, b, update, "В очередь не добавлен ни один участник.")
		return
	}

	nextBookId := session.Host.NextBookId
	book := bookClub.GetBookById(nextBookId)
	message := fmt.Sprintf("Ближайшее %d-е собрание клуба: %s, ведущий: %s", n, session.Date.Format("01/02/2006"), session.Host.Username)
	if nextBookId != "" {
		message += fmt.Sprintf(", книга: %s - «%s»", book.Title, book.Author)
	}
	message += "."

	sendMessage(ctx, b, update, message)
}

func setQueueHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	bookClub := loadBookclub(ctx, b, update)
	if bookClub == nil {
		return
	}

	hostIndexes := strings.Split(strings.TrimPrefix(update.Message.Text, "/set_queue "), ",")
	telegramIds := []int64{}

	for _, hostIndex := range hostIndexes {
		index, err := strconv.Atoi(strings.TrimSpace(hostIndex))
		if err != nil {
			sendMessage(ctx, b, update, "Неверный формат. Пожалуйста, используйте формат '/set_queue <номера участников через запятую>'.")
			return
		}

		host := bookClub.GetHosts()[index-1]

		if host == nil {
			sendMessage(ctx, b, update, "Участник не найден.")
			return
		}

		telegramIds = append(telegramIds, host.TelegramId)
	}

	err := bookClub.SetQueue(telegramIds)

	if err != nil {
		sendMessage(ctx, b, update, "Один или несколько участников не найдены.")
		return
	}

	err = bookClub.Save()
	if err != nil {
		handleError(ctx, b, update, "He удалось сохранить.", err)
		return
	}

	sendMessage(ctx, b, update, "Готово!")
}

func nextSessionHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	bookClub := loadBookclub(ctx, b, update)
	if bookClub == nil {
		return
	}

	session, err := bookClub.GetNthSession(1)
	if err != nil {
		sendMessage(ctx, b, update, "В очередь не добавлен ни один участник.")
		return
	}

	nextBookId := session.Host.NextBookId
	message := fmt.Sprintf("Cледующее собрание клуба: %s, ведущий: %s", session.Date.Format("01/02/2006"), session.Host.Username)
	if nextBookId != "" {
		book := bookClub.GetBookById(nextBookId)
		message += fmt.Sprintf(", книга: %s - «%s»", book.Author, book.Title)
	}
	message += "."

	sendMessage(ctx, b, update, message)
}

func getQueueHandler(ctx context.Context, b *bot.Bot, update *models.Update) {
	bookClub := loadBookclub(ctx, b, update)
	if bookClub == nil {
		return
	}

	message := "Очередь ведущих:"

	hosts := bookClub.GetHosts()

	for _, telegramId := range bookClub.GetQueue() {
		message += fmt.Sprintf("\n%s", hosts[telegramId].Username)
	}

	sendMessage(ctx, b, update, message)
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

func handleError(ctx context.Context, b *bot.Bot, update *models.Update, text string, err error) {
	fmt.Println(err)
	sendMessage(ctx, b, update, text)
}

func sendMessage(ctx context.Context, b *bot.Bot, update *models.Update, text string) {
	b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: update.Message.Chat.ID,
		Text:   text,
	})
}
