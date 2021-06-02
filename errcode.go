package summer

type errCode int

// ensure "ErrCode" implements "error"
var _ error = (*errCode)(nil)

const (
	StatusOK errCode = 200
)

var errCodeLookup = map[errCode]string{
	StatusOK: "Ok",
}

func (e errCode) String() string {
	if val, ok := errCodeLookup[e]; ok {
		return val
	} else {
		return "Something wrong!"
	}
}

func (e errCode) Error() string {
	return e.String()
}
