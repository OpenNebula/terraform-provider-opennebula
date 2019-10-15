package os

var HostOS = hostOS // for monkey patching

type OSType int

const (
	Unknown OSType = iota
	Ubuntu
	Windows
	OSX
	CentOS
	Debian
	Arch
	FreeBSD
)

func (t OSType) String() string {
	switch t {
	case Ubuntu:
		return "Ubuntu"
	case Windows:
		return "Windows"
	case OSX:
		return "OSX"
	case CentOS:
		return "CentOS"
	case Debian:
		return "Debian"
	case Arch:
		return "Arch"
	case FreeBSD:
		return "FreeBSD"
	}
	return "Unknown"
}
