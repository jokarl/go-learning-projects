package tasks

type Task struct {
	Id          int       `json:"id"`
	Description string    `json:"description"`
	CreatedAt   string    `json:"createdAt"`
	Status      TaskState `json:"status"`
}

type TaskState int

const (
	Completed TaskState = iota
	Pending
)

var stateName = map[TaskState]string{
	Completed: "completed",
	Pending:   "pending",
}

func (ts TaskState) MarshalJSON() ([]byte, error) {
	name, ok := stateName[ts]
	if !ok {
		name = "unknown"
	}
	return []byte(`"` + name + `"`), nil
}

func (ts TaskState) String() string {
	name, ok := stateName[ts]
	if !ok {
		name = "unknown"
	}
	return name
}

func (ts TaskState) Emoji() string {
	switch ts {
	case Completed:
		return "\u2713" // ✓
	case Pending:
		return "\U0001F551" // 🕑
	default:
		return "unknown"
	}
}
