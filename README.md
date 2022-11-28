# clcnt

Calories counter.

## init & build

See `make`.

## run

### dev

`./clcnt`

### production

`GIN_MODE=release ./clcnt`

## usage (localhost)

| Use case | Verb | URL |
| --- | --- | --- |
| Add breakfast entry with 350 calories | POST | localhost:8080/api/v1/entry/Breakfast/350 |
| Retrieve all entries | GET | localhost:8080/api/v1/entry |
| Get today's calories in total | GET | localhost:8080/api/v1/calories |
| Get 3 days calories average | GET | localhost:8080/api/v1/calories?days=3 |

## backup & restore

To back up and restore, simply copy `clcnt.db`.

## bookmarks

- https://www.allhandsontech.com/programming/golang/how-to-use-sqlite-with-go/
- https://www.allhandsontech.com/programming/golang/web-app-sqlite-go/
- https://sqlitebrowser.org/
- https://www.epochconverter.com/

### further refs

- https://www.golang.dk/articles/go-and-sqlite-in-the-cloud
- https://github.com/maragudk/sqlite-app
- https://github.com/maragudk/sqlite-app/blob/main/cmd/server/main.go

## to do

- Test boundaries
- readinessHandler
- TODOs

## backlog

- Update entries
- Delete specific entries
- Delete old entries