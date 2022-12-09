package core

type Mail struct {
    Id   int
    Args []interface{}
}

func (m *Mail) Get() []interface{} {
    var result []interface{}
    result = append(result, m.Id)
    return append(result, m.Args...)
}
