# clcnt

A **c**a**l**ories **c**ou**nt**er with a `go` backend and a web frontend, intended to run locally on a (Pixel) smartphone.

## prerequisites

AT WORK!

### Runtime

- A Pixel (6) device with `termux` installed.
- golang (1.19) installed on `termux`.

## init & build

Run `make`.

## run

### dev

`./clcnt -debug` or `go run main.go -debug`

### production

`./clcnt`

## usage (localhost)

### frontend

Browse to [localhost](http://localhost:8080).

### backend

| Use case | Verb | URL |
| --- | --- | --- |
| Add breakfast entry with 350 calories | POST | localhost:8080/api/v1/entry/Breakfast/350 |
| Retrieve all entries | GET | localhost:8080/api/v1/entry |
| Get today's calories in total | GET | localhost:8080/api/v1/calories |
| Get 3 days calories average | GET | localhost:8080/api/v1/calories?days=3 |

## backup & restore

To back up and restore, simply copy `clcnt.db`. **Backup prior to every update!**

## bookmarks

- https://www.allhandsontech.com/programming/golang/how-to-use-sqlite-with-go/
- https://www.allhandsontech.com/programming/golang/web-app-sqlite-go/
- https://sqlitebrowser.org/
- https://www.epochconverter.com/
- https://developers.google.com/chart/interactive/docs/gallery/gauge

### further refs

- https://www.golang.dk/articles/go-and-sqlite-in-the-cloud
- https://github.com/maragudk/sqlite-app
- https://github.com/maragudk/sqlite-app/blob/main/cmd/server/main.go
- https://github.com/gin-gonic/gin#serving-static-files

## known issues

- `make build` and `make build-dist` don't seem to work on a Macbook. I didn't investigate.

## to do

- Rename to calcnt
- Test boundaries (backend and frontend)
- Move code into packages/includes (backend and frontend)
- TODOs in code
- Add disclaimer, mention sources, esp. font awesome
- Documentation (Swagger, deployment)

## backlog

- Update entries
- Delete specific entries
- Delete old entries
