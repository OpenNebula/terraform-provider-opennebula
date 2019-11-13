package bills

import (
	"github.com/megamsys/libgo/utils"
	"strconv"
	"strings"
)

func init() {
	Register(utils.SCYLLAMGR, scylladbManager{})
}

type scylladbManager struct{}

func (m scylladbManager) IsEnabled() bool {
	return true
}

func (m scylladbManager) Onboard(o *BillOpts, mi map[string]string) error {
	return nil
}

func (m scylladbManager) Deduct(o *BillOpts, mi map[string]string) error {
	b, err := NewBalances(o.AccountId, mi)
	if err != nil {
		return err
	}

	if err = b.Deduct(&BalanceOpts{
		Id:       o.AccountId,
		Consumed: o.Consumed,
	}, mi); err != nil {
		return err
	}
	return nil
}

func (m scylladbManager) Transaction(o *BillOpts, mi map[string]string) error {
	//bt, err := carton.NewBillTransaction(o.AccountsId)
	bt, err := NewBilledHistories(o)
	if err != nil {
		return err
	}
	//if err = bt.Transact(&carton.BillTransactionOpts{
	if err = bt.BilledHistories(mi); err != nil {
		return err
	}
	return nil
}

func (m scylladbManager) AuditUnpaid(o *BillOpts, mi map[string]string) error {
	sk := &EventsSkews{
		AccountId: o.AccountId,
		CatId:     o.AssemblyId,
		EventType: o.SkewsType,
	}

	if strings.Split(o.SkewsType, ".")[1] == "quota" {
		return sk.SkewsQuotaUnpaid(o, mi)
	}

	b, err := NewBalances(o.AccountId, mi)
	if err != nil {
		return err
	}
	cb, _ := strconv.ParseFloat(b.Credit, 64)
	if cb <= 0 {
		return sk.ActionEvents(o, b.Credit, mi)
	}

	return sk.DeactiveEvents(o, mi)
}

func (m scylladbManager) Invoice(o *BillOpts) error {
	return nil
}

func (m scylladbManager) Nuke(o *BillOpts) error {
	return nil
}

func (m scylladbManager) Suspend(o *BillOpts) error {
	return nil
}

func (m scylladbManager) Notify(o *BillOpts) error {

	return nil
}
