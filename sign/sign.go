package sign

type Sign string

func (s Sign) String() string {
	return string(s)
}

const (
	LOGGER         Sign = "logger"
	TRACE_ID       Sign = "trace_id"
	Error          Sign = "error"
	TRANSACTION    Sign = "trans"
	DIS_LOCK       Sign = "distribute_lock"
	IS_GLOBAL_LOCK Sign = "is_global_lock"
)
