package db

import (
	log "github.com/Sirupsen/logrus"
	"github.com/megamsys/gocassa"
	"github.com/megamsys/libgo/cmd"
)

type RelationsFunc func() gocassa.Relation

type ScyllaDB struct {
	NodeIps []string
	KS      gocassa.KeySpace
}

type ScyllaTable struct {
	T gocassa.MultimapMkTable
}

type ScyllaWhere struct {
	Clauses map[string]string
}

type ScyllaDBOpts struct {
	KeySpaceName string
	NodeIps      []string
	Username     string
	Password     string
	Debug        bool
}

func newScyllaDB(opts ScyllaDBOpts) (*ScyllaDB, error) {
	ks, err := connectToKeySpace(opts.KeySpaceName, opts.NodeIps, opts.Username, opts.Password)
	if err != nil {
		return nil, err
	}
	ks.DebugMode(opts.Debug)

	return &ScyllaDB{
		NodeIps: opts.NodeIps,
		KS:      ks,
	}, nil
}

// Connect to a certain keyspace directly. Same as using Connect().KeySpace(keySpaceName)
func connectToKeySpace(keySpace string, nodeIps []string, username, password string) (gocassa.KeySpace, error) {
	c, err := gocassa.Connect(nodeIps, username, password)
	if err != nil {
		return nil, err
	}
	log.Debugf(cmd.Colorfy("  > [scylla] keyspace "+keySpace, "blue", "", "bold"))
	return c.KeySpace(keySpace), nil
}

func (sy *ScyllaDB) table(name string, pks []string, ccms []string, out interface{}) *ScyllaTable {
	log.Debugf(cmd.Colorfy("  > [scylla] table "+name, "blue", "", "bold"))
	return &ScyllaTable{T: sy.KS.MultimapMultiKeyTable(name, pks, ccms, out)}
}

func (sy *ScyllaDB) Close() {
	log.Debugf(cmd.Colorfy("  > [scylla] Connection close", "blue", "", "bold"))
	sy.KS.Close()
}

func (st *ScyllaTable) read(fields, ids map[string]interface{}, out interface{}) error {
	log.Debugf(cmd.Colorfy("  > [scylla] read", "blue", "", "bold"))
	op := gocassa.Options{AllowFiltering: true}
	return st.T.Read(fields, ids, out).WithOptions(op).Run()
}

func (st *ScyllaTable) readMulti(fields map[string]interface{}, limit int, out interface{}) error {
	log.Debugf(cmd.Colorfy("  > [scylla] read", "blue", "", "bold"))
	op := gocassa.Options{AllowFiltering: true}
	return st.T.List(fields, nil, limit, out).WithOptions(op).Run()
}

func (st *ScyllaTable) insert(data interface{}) error {
	log.Debugf(cmd.Colorfy("  > [scylla] insert", "blue", "", "bold"))
	return st.T.Set(data).Run()
}

func (st *ScyllaTable) update(tinfo Options, data map[string]interface{}) error {
	log.Debugf(cmd.Colorfy("  > [scylla] update", "blue", "", "bold"))
	return st.T.Update(tinfo.PksClauses, tinfo.CcmsClauses, data).Run()
}

func (st *ScyllaTable) deleterow(tinfo Options) error {
	log.Debugf(cmd.Colorfy("  > [scylla] delete", "blue", "", "bold"))
	return st.T.Delete(tinfo.PksClauses, tinfo.CcmsClauses).Run()
}
