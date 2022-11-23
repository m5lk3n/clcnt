# clcnt

calories counter

## init

```bash
go mod init lttl.dev/clcnt
go get -u github.com/sirupsen/logrus
go get -u github.com/mattn/go-sqlite3
go get -u github.com/gin-gonic/gin
```

## build

`go build`

## run

### dev

`./clcnt`

### production

`GIN_MODE=release ./clcnt`

## bookmarks

- https://www.allhandsontech.com/programming/golang/how-to-use-sqlite-with-go/
- https://www.allhandsontech.com/programming/golang/web-app-sqlite-go/
- https://www.epochconverter.com/

### further refs

- https://www.golang.dk/articles/go-and-sqlite-in-the-cloud
- https://github.com/maragudk/sqlite-app
- https://github.com/maragudk/sqlite-app/blob/main/cmd/server/main.go

## to do

- Sum current day
- Sum last 7 days
- Sum overall