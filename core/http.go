package core

import (
    "fmt"
    "net/http"
    "strconv"
    "strings"
)

func StartHttp(ch chan Mail, args ...interface{}) {
    addr := args[0].(string)
    mux := http.NewServeMux()
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        idStr := r.URL.Query().Get("id")
        argsStr := r.URL.Query().Get("args")
        id, _ := strconv.Atoi(idStr)
        var args []interface{}
        for _, s := range strings.Split(argsStr, ",") {
            args = append(args, s)
        }
        fmt.Println(args)
        ch <- Mail{
            Id:   id,
            Args: args,
        }
    })

    go http.ListenAndServe(addr, mux)
}
