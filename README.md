# Book Club
Telegram bot for book clubs.

## Disclaimers
- Currently only supports Russian language.
- Gatherings can only happen on the 2nd Tuesday of every Month.


# Usage
```
export BOOKCLUB_BOTTOKEN=[your-token]
go get
go run main.go
```

### With docker
```
export BOOKCLUB_BOTTOKEN=[your-token]
docker-compose up
```

### Local development
For convenience you can copy `.env-template` to `.env` and populate your token there instead of exporting the variable.
