package opennebula

import (
	"strconv"
	"strings"

	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/shared"
)

func permissionsUnixString(p shared.Permissions) string {
	owner := p.OwnerU<<2 | p.OwnerM<<1 | p.OwnerA
	group := p.GroupU<<2 | p.GroupM<<1 | p.GroupA
	other := p.OtherU<<2 | p.OtherM<<1 | p.OtherA
	return strconv.Itoa(int(owner)*100 + int(group)*10 + int(other))
}

func permissionUnix(p string) shared.Permissions {

	perms := strings.Split(p, "")
	owner, _ := strconv.Atoi(perms[0])
	group, _ := strconv.Atoi(perms[1])
	other, _ := strconv.Atoi(perms[2])

	return shared.Permissions{
		OwnerU: int8(owner & 4 >> 2),
		OwnerM: int8(owner & 2 >> 1),
		OwnerA: int8(owner & 1),
		GroupU: int8(group & 4 >> 2),
		GroupM: int8(group & 2 >> 1),
		GroupA: int8(group & 1),
		OtherU: int8(other & 4 >> 2),
		OtherM: int8(other & 2 >> 1),
		OtherA: int8(other & 1),
	}
}
