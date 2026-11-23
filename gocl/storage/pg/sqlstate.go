package pg

type SQLState string

const (
	// SQLState23505 unique constraint violation.
	SQLState23505 = SQLState("23505")
	// SQLState23514 check constraint violation.
	SQLState23514 = SQLState("23514")
	// SqlState40001 serialization failure due concurrent requests.
	SqlState40001 = SQLState("40001")
)

func (v SQLState) String() string {
	return string(v)
}
