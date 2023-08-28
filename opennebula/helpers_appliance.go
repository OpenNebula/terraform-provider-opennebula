package opennebula

type ApplianceType string

const (
	AppTypeImage   = "IMAGE"
	AppTypeVM      = "VMTEMPLATE"
	AppTypeService = "SERVICE_TEMPLATE"
)

func ApplianceTypeToString(appType int) string {
	switch appType {
	case 0:
		return "UNKNOWN"
	case 1:
		return "IMAGE"
	case 2:
		return "VMTEMPLATE"
	case 3:
		return "SERVICE_TEMPLATE"
	default:
		return ""
	}
}
