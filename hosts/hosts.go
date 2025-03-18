package hosts

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Host struct {
	TelegramId int64  `bson:"telegramId"`
	Username   string `bson:"username"`
	NextBookId string `bson:"nextBookId"`
}

type Book struct {
	Id     string `bson:"id"`
	Author string `bson:"author"`
	Title  string `bson:"title"`
}

type Session struct {
	Host *Host
	Date time.Time
}

type Bookclub struct {
	ChatId              int64   `bson:"_id,omitempty"`
	Hosts               []*Host `bson:"hosts"`
	hostMap             map[int64]*Host
	Queue               []int64 `bson:"queue"`
	Books               []*Book `bson:"books"`
	bookMap             map[string]*Book
	StartHostQueueIndex int       `bson:"startHostQueueIndex"`
	StartTime           time.Time `bson:"startTime"`
}

func LoadBookclub(chatId int64) (*Bookclub, error) {
	fmt.Println("Loading bookclub.")
	bookClub, err := loadBookclub(chatId)
	fmt.Println("Bookclub loaded: ", bookClub)
	if err != nil {
		fmt.Println("Bookclub not found. Creating new bookclub.")
		bookClub = newBookclub(chatId)
		err = insertBookclub(bookClub)
		if err != nil {
			return nil, err
		}
	}

	if bookClub.hostMap == nil {
		bookClub.hostMap = make(map[int64]*Host)
	}

	if bookClub.bookMap == nil {
		bookClub.bookMap = make(map[string]*Book)
	}

	for _, host := range bookClub.Hosts {
		bookClub.hostMap[host.TelegramId] = host
	}

	for _, book := range bookClub.Books {
		bookClub.bookMap[book.Id] = book
	}

	return bookClub, nil
}

func (bookclub *Bookclub) Save() error {
	return updateBookclub(bookclub)
}

func newBookclub(chatId int64) *Bookclub {
	return &Bookclub{
		ChatId:              chatId,
		Hosts:               make([]*Host, 0),
		hostMap:             make(map[int64]*Host),
		Queue:               make([]int64, 0),
		Books:               make([]*Book, 0),
		bookMap:             make(map[string]*Book),
		StartHostQueueIndex: 0,
		StartTime:           time.Now(),
	}
}

func (b *Bookclub) AddHost(telegramId int64, username string) error {
	if b.hostMap[telegramId] != nil {
		return errors.New("Host already exists.")
	}

	host := &Host{TelegramId: telegramId, Username: username}

	b.Hosts = append(b.Hosts, host)
	b.hostMap[telegramId] = host

	return nil
}

func (b *Bookclub) RemoveHost(telegramId int64) error {
	toDelete := b.hostMap[telegramId]

	if toDelete == nil {
		return errors.New("Host does not exist.")
	}

	delete(b.hostMap, toDelete.TelegramId)

	for i, host := range b.Hosts {
		if host == toDelete {
			b.Hosts = remove(b.Hosts, i)
			break
		}
	}

	return nil
}

func (b *Bookclub) SetNextBook(telegramId int64, bookIndex int) error {
	book := b.Books[bookIndex]
	if book == nil {
		return errors.New("Book does not exist.")
	}
	b.hostMap[telegramId].NextBookId = book.Id
	return nil
}

func (b *Bookclub) SetQueue(telegramIds []int64) error {
	for _, id := range telegramIds {
		if b.hostMap[id] == nil {
			return errors.New("One or more hosts do not exist.")
		}
	}

	b.Queue = telegramIds
	return nil
}

func (b *Bookclub) GetQueue() []int64 {
	return b.Queue
}

func (b *Bookclub) SetStartHost(telegramId int64) error {
	for i, id := range b.Queue {
		if id == telegramId {
			b.StartHostQueueIndex = i
			return nil
		}
	}

	return errors.New("Host not found in queue.")
}

func (b *Bookclub) AddBook(author string, title string) {
	uuid := uuid.New().String()
	book := &Book{Id: uuid, Author: author, Title: title}
	b.Books = append(b.Books, book)
	b.bookMap[uuid] = book
}

func (b *Bookclub) GetBookByIndex(bookIndex int) *Book {
	return b.Books[bookIndex]
}

func (b *Bookclub) GetBookById(bookId string) *Book {
	return b.bookMap[bookId]
}

func (b *Bookclub) GetBooks() []*Book {
	return b.Books
}

func (b *Bookclub) GetHosts() []*Host {
	return b.Hosts
}

func (b *Bookclub) DeleteBook(bookIndex int) error {
	book := b.Books[bookIndex]
	if book == nil {
		return errors.New("Book does not exist.")
	}
	b.Books = remove(b.Books, bookIndex)
	delete(b.bookMap, book.Id)
	return nil
}

func (b *Bookclub) GetNthSession(n int) (*Session, error) {
	if len(b.Queue) == 0 {
		return nil, errors.New("No hosts in queue.")
	}

	nextDate := time.Now()
	for i := 0; i < n; i++ {
		nextDate = nextSecondTuesday(nextDate)
		nextDate = nextDate.AddDate(0, 0, 1)
	}

	nextDate = nextDate.AddDate(0, 0, -1)
	queueIndex := (monthDiff(b.StartTime, nextDate) + b.StartHostQueueIndex) % len(b.Queue)
	host := b.hostMap[b.Queue[queueIndex]]

	return &Session{
		Host: host,
		Date: nextDate,
	}, nil
}

func monthDiff(t1, t2 time.Time) int {
	year1, month1, _ := t1.Date()
	year2, month2, _ := t2.Date()
	return (year2-year1)*12 + int(month2-month1)
}

func nextSecondTuesday(date time.Time) time.Time {
	for !isSecondTuesday(date) {
		date = date.AddDate(0, 0, 1)
	}
	return date
}

func isSecondTuesday(t time.Time) bool {
	if t.Weekday() == time.Tuesday && t.Day() > 7 && t.Day() < 15 {
		return true
	}
	return false
}

func remove[T any](slice []T, s int) []T {
	return append(slice[:s], slice[s+1:]...)
}
