package snapshot

import (
	"github.com/megamsys/opennebula-go/api"
)

type Snapshot struct {
	Name            string `xml:"NAME"`
	VMId            int    `xml:"VMID"`
	DiskId          int    `xml:"DISK_ID"`
	SnapId          int    `xml:"SNAP_ID"`
	DiskDiscription string `xml:"DISK_DISC"`
	T               *api.Rpc
}

func (s *Snapshot) CreateSnapshot() (interface{}, error) {
	args := []interface{}{s.T.Key, s.VMId, s.DiskId, s.DiskDiscription}
	res, err := s.T.Call(api.DISK_SNAPSHOT_CREATE, args)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Snapshot) DeleteSnapshot() (interface{}, error) {
	args := []interface{}{s.T.Key, s.VMId, s.DiskId, s.SnapId}
	res, err := s.T.Call(api.DISK_SNAPSHOT_DELETE, args)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Snapshot) SnapshotSaveAs() (interface{}, error) {
	args := []interface{}{s.T.Key, s.VMId, s.DiskId, s.SnapId}
	res, err := s.T.Call(api.DISK_SNAPSHOT_DELETE, args)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Snapshot) RevertSnapshot() (interface{}, error) {
	args := []interface{}{s.T.Key, s.VMId, s.DiskId, s.SnapId}
	res, err := s.T.Call(api.DISK_SNAPSHOT_REVERT, args)
	if err != nil {
		return nil, err
	}
	return res, nil
}
