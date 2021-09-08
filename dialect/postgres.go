package dialect

type Postgres string

func (x Postgres) Type(name string, size int) string {
	return ``
}
