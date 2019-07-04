package opennebula

import (
	"strconv"
	"strings"

	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/shared"
)

func permissionsUnixString(p *shared.Permissions) string {
	owner := p.OwnerU<<2 | p.OwnerM<<1 | p.OwnerA
	group := p.GroupU<<2 | p.GroupM<<1 | p.GroupA
	other := p.OtherU<<2 | p.OtherM<<1 | p.OtherA
	return strconv.Itoa(owner*100 + group*10 + other)
}

func permissionUnix(p string) *shared.Permissions {
	perms := strings.Split(p, "")
	owner, _ := strconv.Atoi(perms[0])
	group, _ := strconv.Atoi(perms[1])
	other, _ := strconv.Atoi(perms[2])

	return &shared.Permissions{
		OwnerU: owner & 4 >> 2,
		OwnerM: owner & 2 >> 1,
		OwnerA: owner & 1,
		GroupU: group & 4 >> 2,
		GroupM: group & 2 >> 1,
		GroupA: group & 1,
		OtherU: other & 4 >> 2,
		OtherM: other & 2 >> 1,
		OtherA: other & 1,
	}
}
