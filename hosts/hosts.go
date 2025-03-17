package hosts

import (
	"errors"
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
	ChatId              int64            `bson:"_id,omitempty"`
	Hosts               map[int64]*Host  `bson:"hosts"`
	Queue               []int64          `bson:"queue"`
	Books               map[string]*Book `bson:"books"`
	StartHostQueueIndex int              `bson:"startHostQueueIndex"`
	StartTime           time.Time        `bson:"startTime"`
}

func LoadBookclub(chatId int64) (*Bookclub, error) {
	bookClub, err := loadBookclub(chatId)
	if err != nil {
		bookClub = newBookclub(chatId)
		err = insertBookclub(bookClub)
		if err != nil {
			return nil, err
		}
	}

	return bookClub, nil
}

func (bookclub *Bookclub) Save() error {
	return updateBookclub(bookclub)
}

func newBookclub(chatId int64) *Bookclub {
	return &Bookclub{
		ChatId:              chatId,
		Hosts:               make(map[int64]*Host),
		Queue:               make([]int64, 0),
		Books:               make(map[string]*Book),
		StartHostQueueIndex: 0,
		StartTime:           time.Now(),
	}
}

func (b *Bookclub) AddHost(telegramId int64, username string) error {
	if b.Hosts[telegramId] != nil {
		return errors.New("Host already exists.")
	}
	b.Hosts[telegramId] = &Host{TelegramId: telegramId, Username: username}
	return nil
}

func (b *Bookclub) RemoveHost(telegramId int64) error {
	if b.Hosts[telegramId] == nil {
		return errors.New("Host does not exist.")
	}
	delete(b.Hosts, telegramId)
	return nil
}

func (b *Bookclub) SetNextBook(telegramId int64, bookId string) error {
	if b.Books[bookId] == nil {
		return errors.New("Book does not exist.")
	}
	b.Hosts[telegramId].NextBookId = bookId
	return nil
}

func (b *Bookclub) SetQueue(telegramIds []int64) error {
	for _, id := range telegramIds {
		if b.Hosts[id] == nil {
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
	b.Books[uuid] = &Book{Id: uuid, Author: author, Title: title}
}

func (b *Bookclub) GetBook(bookId string) *Book {
	return b.Books[bookId]
}

func (b *Bookclub) GetBooks() map[string]*Book {
	return b.Books
}

func (b *Bookclub) GetHosts() map[int64]*Host {
	return b.Hosts
}

func (b *Bookclub) DeleteBook(bookId string) error {
	if b.Books[bookId] == nil {
		return errors.New("Book does not exist.")
	}
	delete(b.Books, bookId)
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
	host := b.Hosts[b.Queue[queueIndex]]

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
