# VanGO

## [VanJS](https://github.com/vanjs-org/van) but in golang.


<img src="https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcTtHIZV8z8Devr_vFmzk8EoIkX8g3C4pgoeLg&s" height="150px"/>

## Installation
```sh
go get github.com/danielberigoi/van-go
```

## Usage
```go
import (
    "fmt"
    "log"
    "net/http"

    . "github.com/danielberigoi/van-go"
)

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprint(w, Html()(
            Head()(),
            Body()(
                H1()("Hello, World!"),
            ),
        ))
    })

    log.Fatal(http.ListenAndServe(":8080", nil))
}
```
