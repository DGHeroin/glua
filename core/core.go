package core

func Filter(ch chan Mail, id int, args ...interface{}) {
    switch id {
    case -100:
        StartHttp(ch, args...)
    }
}
