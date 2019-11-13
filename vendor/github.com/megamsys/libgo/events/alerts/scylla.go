package alerts

import (
	constants "github.com/megamsys/libgo/utils"
	"strings"
)

type Scylla struct {
	Scylla_host     []string
	Scylla_keyspace string
	Scylla_username string
	Scylla_password string
}

func NewScylla(m map[string]string) Notifier {
	return &Scylla{
		Scylla_host:     strings.Split(m[constants.SCYLLAHOST], ","),
		Scylla_keyspace: m[constants.SCYLLAKEYSPACE],
		Scylla_username: m[constants.SCYLLAUSERNAME],
		Scylla_password: m[constants.SCYLLAPASSWORD],
	}
}

func (s *Scylla) satisfied(eva EventAction) bool {
	if eva == STATUS {
		return true
	}
	return false
}

func (s *Scylla) Notify(eva EventAction, edata EventData) error {
	value := edata.M[constants.EVENT_TYPE]
	et := strings.Split(value, ".")
	if et[0] == "obc" {
		return s.NotifyOBC(eva, edata)
	}
	return nil
}
