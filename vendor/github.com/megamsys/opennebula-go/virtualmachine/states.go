package virtualmachine

import ()

type VmState int
type LcmState int

var LcmStateString = map[LcmState]string{LCM_INIT: "lcm_init", PROLOG: "prolog", BOOT: "boot", RUNNING: "running",
	MIGRATE: "migrate", SAVE_STOP: "save_stop", SAVE_SUSPEND: "save_suspend", SAVE_MIGRATE: "save_migrate",
	PROLOG_MIGRATE: "prolog_migrate", PROLOG_RESUME: "prolog_resume", EPILOG_STOP: "epilog_stop", EPILOG: "epilog",
	SHUTDOWN: "shutdown", CLEANUP_RESUBMIT: "cleanup_resubmit", UNKNOWN: "unknown", HOTPLUG: "hotplug",
	SHUTDOWN_POWEROFF: "shutdown_poweroff", BOOT_UNKNOWN: "boot_unknown", BOOT_POWEROFF: "boot_poweroff",
	BOOT_SUSPENDED: "boot_suspended", BOOT_STOPPED: "boot_stopped", CLEANUP_DELETE: "cleanup_delete",
	HOTPLUG_SNAPSHOT: "hotplug_snapshot", HOTPLUG_NIC: "hotplug_nic", HOTPLUG_SAVEAS: "hotplug_saveas",
	HOTPLUG_SAVEAS_POWEROFF: "hotplug_saveas_poweroff", HOTPLUG_SAVEAS_SUSPENDED: "hotplug_saveas_suspended",
	SHUTDOWN_UNDEPLOY: "shutdown_undeploy", EPILOG_UNDEPLOY: "epilog_undeploy", PROLOG_UNDEPLOY: "prolog_undeploy",
	BOOT_UNDEPLOY: "boot_undeploy", HOTPLUG_PROLOG_POWEROFF: "hotplug_prolog_poweroff",
	HOTPLUG_EPILOG_POWEROFF: "hotplug_epilog_poweroff", BOOT_MIGRATE: "boot_migrate", BOOT_FAILURE: "boot_failure",
	BOOT_MIGRATE_FAILURE: "boot_migrate_failure", PROLOG_MIGRATE_FAILURE: "prolog_migrate_failure",
	PROLOG_FAILURE: "prolog_failure", EPILOG_FAILURE: "epilog_failure", EPILOG_STOP_FAILURE: "epilog_stop_failure",
	EPILOG_UNDEPLOY_FAILURE: "epilog_undeploy_failure", PROLOG_MIGRATE_POWEROFF: "prolog_migrate_poweroff",
	PROLOG_MIGRATE_POWEROFF_FAILURE: "prolog_migrate_poweroff_failure", PROLOG_MIGRATE_SUSPEND: "prolog_migrate_suspend",
	PROLOG_MIGRATE_SUSPEND_FAILURE: "prolog_migrate_suspend_failure", BOOT_UNDEPLOY_FAILURE: "boot_undeploy_failure",
	BOOT_STOPPED_FAILURE: "boot_stopped_failure", PROLOG_RESUME_FAILURE: "prolog_resume_failure",
	PROLOG_UNDEPLOY_FAILURE: "prolog_undeploy_failure", DISK_SNAPSHOT_POWEROFF: "disk_snapshot_poweroff",
	DISK_SNAPSHOT_REVERT_POWEROFF: "disk_snapshot_revert_poweroff", DISK_SNAPSHOT_DELETE_POWEROFF: "disk_snapshot_delete_poweroff",
	DISK_SNAPSHOT_SUSPENDED: "disk_snapshot_suspended", DISK_SNAPSHOT_REVERT_SUSPENDED: "disk_snapshot_revert_suspended",
	DISK_SNAPSHOT_DELETE_SUSPENDED: "disk_snapshot_delete_suspended", DISK_SNAPSHOT: "disk_snapshot",
	DISK_SNAPSHOT_DELETE: "disk_snapshot_delete", PROLOG_MIGRATE_UNKNOWN: "prolog_migrate_unknown",
	PROLOG_MIGRATE_UNKNOWN_FAILURE: "prolog_migrate_unknown_failure"}

var VmStateString = map[VmState]string{INIT: "init", PENDING: "pending", HOLD: "hold", ACTIVE: "active", STOPPED: "stopped", SUSPENDED: "suspended", DONE: "done", UNKNOWNSTATE: "unknown", POWEROFF: "poweroff", UNDEPLOYED: "undeployed"}

const (

	//LcmState starts at 0
	LCM_INIT LcmState = iota
	PROLOG
	BOOT
	RUNNING
	MIGRATE
	SAVE_STOP
	SAVE_SUSPEND
	SAVE_MIGRATE
	PROLOG_MIGRATE
	PROLOG_RESUME
	EPILOG_STOP
	EPILOG
	SHUTDOWN
	CLEANUP_RESUBMIT
	UNKNOWN
	HOTPLUG
	SHUTDOWN_POWEROFF
	BOOT_UNKNOWN
	BOOT_POWEROFF
	BOOT_SUSPENDED
	BOOT_STOPPED
	CLEANUP_DELETE
	HOTPLUG_SNAPSHOT
	HOTPLUG_NIC
	HOTPLUG_SAVEAS
	HOTPLUG_SAVEAS_POWEROFF
	HOTPLUG_SAVEAS_SUSPENDED
	SHUTDOWN_UNDEPLOY
	EPILOG_UNDEPLOY
	PROLOG_UNDEPLOY
	BOOT_UNDEPLOY
	HOTPLUG_PROLOG_POWEROFF
	HOTPLUG_EPILOG_POWEROFF
	BOOT_MIGRATE
	BOOT_FAILURE
	BOOT_MIGRATE_FAILURE
	PROLOG_MIGRATE_FAILURE
	PROLOG_FAILURE
	EPILOG_FAILURE
	EPILOG_STOP_FAILURE
	EPILOG_UNDEPLOY_FAILURE
	PROLOG_MIGRATE_POWEROFF
	PROLOG_MIGRATE_POWEROFF_FAILURE
	PROLOG_MIGRATE_SUSPEND
	PROLOG_MIGRATE_SUSPEND_FAILURE
	BOOT_UNDEPLOY_FAILURE
	BOOT_STOPPED_FAILURE
	PROLOG_RESUME_FAILURE
	PROLOG_UNDEPLOY_FAILURE
	DISK_SNAPSHOT_POWEROFF
	DISK_SNAPSHOT_REVERT_POWEROFF
	DISK_SNAPSHOT_DELETE_POWEROFF
	DISK_SNAPSHOT_SUSPENDED
	DISK_SNAPSHOT_REVERT_SUSPENDED
	DISK_SNAPSHOT_DELETE_SUSPENDED
	DISK_SNAPSHOT
	DISK_SNAPSHOT_DELETE
	PROLOG_MIGRATE_UNKNOWN
	PROLOG_MIGRATE_UNKNOWN_FAILURE
)
