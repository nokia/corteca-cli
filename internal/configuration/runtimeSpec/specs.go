package specs

import "os"

// Spec is the base configuration for the container.
type Spec struct {
	// Version of the Open Container Initiative Runtime Specification with which the bundle complies.
	Version string `json:"ociVersion" yaml:"ociVersion"`
	// Process configures the container process.
	Process *Process `json:"process,omitempty" yaml:"process,omitempty"`
	// Root configures the container's root filesystem.
	Root *Root `json:"root,omitempty" yaml:"root,omitempty"`
	// Hostname configures the container's hostname.
	Hostname string `json:"hostname,omitempty" yaml:"hostname"`
	// Domainname configures the container's domainname.
	Domainname string `json:"domainname,omitempty" yaml:"domainname,omitempty"`
	// Mounts configures additional mounts (on top of Root).
	Mounts []Mount `json:"mounts,omitempty" yaml:"mounts,omitempty"`
	// Hooks configures callbacks for container lifecycle events.
	Hooks *Hooks `json:"hooks,omitempty" yaml:"hooks,omitempty" platform:"linux,solaris,zos"`
	// Annotations contains arbitrary metadata for the container.
	Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`

	// Linux is platform-specific configuration for Linux based containers.
	Linux *Linux `json:"linux,omitempty" yaml:"linux,omitempty" platform:"linux"`
	// Solaris is platform-specific configuration for Solaris based containers.
	Solaris *Solaris `json:"solaris,omitempty" yaml:"solaris,omitempty" platform:"solaris"`
	// Windows is platform-specific configuration for Windows based containers.
	Windows *Windows `json:"windows,omitempty" yaml:"windows,omitempty" platform:"windows"`
	// VM specifies configuration for virtual-machine-based containers.
	VM *VM `json:"vm,omitempty" yaml:"vm,omitempty" platform:"vm"`
	// ZOS is platform-specific configuration for z/OS based containers.
	ZOS *ZOS `json:"zos,omitempty" yaml:"zos,omitempty" platform:"zos"`
}

// Scheduler represents the scheduling attributes for a process. It is based on
// the Linux sched_setattr(2) syscall.
type Scheduler struct {
	// Policy represents the scheduling policy (e.g., SCHED_FIFO, SCHED_RR, SCHED_OTHER).
	Policy LinuxSchedulerPolicy `json:"policy" yaml:"policy"`

	// Nice is the nice value for the process, which affects its priority.
	Nice int32 `json:"nice,omitempty" yaml:"nice,omitempty"`

	// Priority represents the static priority of the process.
	Priority int32 `json:"priority,omitempty" yaml:"priority,omitempty"`

	// Flags is an array of scheduling flags.
	Flags []LinuxSchedulerFlag `json:"flags,omitempty" yaml:"flags,omitempty"`

	// The following ones are used by the DEADLINE scheduler.

	// Runtime is the amount of time in nanoseconds during which the process
	// is allowed to run in a given period.
	Runtime uint64 `json:"runtime,omitempty" yaml:"runtime,omitempty"`

	// Deadline is the absolute deadline for the process to complete its execution.
	Deadline uint64 `json:"deadline,omitempty" yaml:"deadline,omitempty"`

	// Period is the length of the period in nanoseconds used for determining the process runtime.
	Period uint64 `json:"period,omitempty" yaml:"period,omitempty"`
}

// Process contains information to start a specific application inside the container.
type Process struct {
	// Terminal creates an interactive terminal for the container.
	Terminal bool `json:"terminal,omitempty" yaml:"terminal,omitempty"`
	// ConsoleSize specifies the size of the console.
	ConsoleSize *Box `json:"consoleSize,omitempty" yaml:"consoleSize,omitempty"`
	// User specifies user information for the process.
	User User `json:"user" yaml:"user"`
	// Args specifies the binary and arguments for the application to execute.
	Args []string `json:"args,omitempty" yaml:"args,omitempty"`
	// CommandLine specifies the full command line for the application to execute on Windows.
	CommandLine string `json:"commandLine,omitempty" yaml:"commandLine,omitempty" platform:"windows"`
	// Env populates the process environment for the process.
	Env []string `json:"env,omitempty" yaml:"env,omitempty"`
	// Cwd is the current working directory for the process and must be
	// relative to the container's root.
	Cwd string `json:"cwd" yaml:"cwd"`
	// Capabilities are Linux capabilities that are kept for the process.
	Capabilities *LinuxCapabilities `json:"capabilities,omitempty" yaml:"capabilities,omitempty" platform:"linux"`
	// Rlimits specifies rlimit options to apply to the process.
	Rlimits []POSIXRlimit `json:"rlimits,omitempty" yaml:"rlimits,omitempty" platform:"linux,solaris,zos"`
	// NoNewPrivileges controls whether additional privileges could be gained by processes in the container.
	NoNewPrivileges bool `json:"noNewPrivileges,omitempty" yaml:"noNewPrivileges,omitempty" platform:"linux"`
	// ApparmorProfile specifies the apparmor profile for the container.
	ApparmorProfile string `json:"apparmorProfile,omitempty" yaml:"apparmorProfile,omitempty" platform:"linux"`
	// Specify an oom_score_adj for the container.
	OOMScoreAdj *int `json:"oomScoreAdj,omitempty" yaml:"oomScoreAdj,omitempty" platform:"linux"`
	// Scheduler specifies the scheduling attributes for a process.
	Scheduler *Scheduler `json:"scheduler,omitempty" yaml:"scheduler,omitempty" platform:"linux"`
	// SelinuxLabel specifies the selinux context that the container process is run as.
	SelinuxLabel string `json:"selinuxLabel,omitempty" yaml:"selinuxLabel,omitempty" platform:"linux"`
	// IOPriority contains the I/O priority settings for the cgroup.
	IOPriority *LinuxIOPriority `json:"ioPriority,omitempty" yaml:"ioPriority,omitempty" platform:"linux"`
	// ExecCPUAffinity specifies CPU affinity for exec processes.
	ExecCPUAffinity *CPUAffinity `json:"execCPUAffinity,omitempty" yaml:"execCPUAffinity,omitempty" platform:"linux"`
}

// LinuxCapabilities specifies the list of allowed capabilities that are kept for a process.
// http://man7.org/linux/man-pages/man7/capabilities.7.html
type LinuxCapabilities struct {
	// Bounding is the set of capabilities checked by the kernel.
	Bounding []string `json:"bounding,omitempty" yaml:"bounding,omitempty" platform:"linux"`
	// Effective is the set of capabilities checked by the kernel.
	Effective []string `json:"effective,omitempty" yaml:"effective,omitempty" platform:"linux"`
	// Inheritable is the capabilities preserved across execve.
	Inheritable []string `json:"inheritable,omitempty" yaml:"inheritable,omitempty" platform:"linux"`
	// Permitted is the limiting superset for effective capabilities.
	Permitted []string `json:"permitted,omitempty" yaml:"permitted,omitempty" platform:"linux"`
	// Ambient is the ambient set of capabilities that are kept.
	Ambient []string `json:"ambient,omitempty" yaml:"ambient,omitempty" platform:"linux"`
}

// IOPriority represents I/O priority settings for the container's processes within the process group.
type LinuxIOPriority struct {
	Class    IOPriorityClass `json:"class" yaml:"class"`
	Priority int             `json:"priority" yaml:"priority"`
}

// IOPriorityClass represents an I/O scheduling class.
type IOPriorityClass string

// Possible values for IOPriorityClass.
const (
	IOPRIO_CLASS_RT   IOPriorityClass = "IOPRIO_CLASS_RT"
	IOPRIO_CLASS_BE   IOPriorityClass = "IOPRIO_CLASS_BE"
	IOPRIO_CLASS_IDLE IOPriorityClass = "IOPRIO_CLASS_IDLE"
)

// CPUAffinity specifies process' CPU affinity.
type CPUAffinity struct {
	Initial string `json:"initial,omitempty" yaml:"initial,omitempty"`
	Final   string `json:"final,omitempty" yaml:"final,omitempty"`
}

// Box specifies dimensions of a rectangle. Used for specifying the size of a console.
// Box specifies dimensions of a rectangle. Used for specifying the size of a console.
type Box struct {
	// Height is the vertical dimension of a box.
	Height uint `json:"height" yaml:"height"`
	// Width is the horizontal dimension of a box.
	Width uint `json:"width" yaml:"width"`
}

// User specifies specific user (and group) information for the container process.
type User struct {
	// UID is the user id.
	UID uint32 `json:"uid" yaml:"uid" platform:"linux,solaris,zos"`
	// GID is the group id.
	GID uint32 `json:"gid" yaml:"gid" platform:"linux,solaris,zos"`
	// Umask is the umask for the init process.
	Umask *uint32 `json:"umask,omitempty" yaml:"umask,omitempty" platform:"linux,solaris,zos"`
	// AdditionalGids are additional group ids set for the container's process.
	AdditionalGids []uint32 `json:"additionalGids,omitempty" yaml:"additionalGids,omitempty" platform:"linux,solaris"`
	// Username is the user name.
	Username string `json:"username,omitempty" yaml:"username,omitempty" platform:"windows"`
}

// Root contains information about the container's root filesystem on the host.
type Root struct {
	// Path is the absolute path to the container's root filesystem.
	Path string `json:"path" yaml:"path"`
	// Readonly makes the root filesystem for the container readonly before the process is executed.
	Readonly bool `json:"readonly,omitempty" yaml:"readonly,omitempty"`
}

// Mount specifies a mount for a container.
type Mount struct {
	// Destination is the absolute path where the mount will be placed in the container.
	Destination string `json:"destination" yaml:"destination"`
	// Type specifies the mount kind.
	Type string `json:"type,omitempty" yaml:"type,omitempty" platform:"linux,solaris,zos"`
	// Source specifies the source path of the mount.
	Source string `json:"source,omitempty" yaml:"source,omitempty"`
	// Options are fstab style mount options.
	Options []string `json:"options,omitempty" yaml:"options,omitempty"`

	// UID/GID mappings used for changing file owners w/o calling chown, fs should support it.
	// Every mount point could have its own mapping.
	UIDMappings []LinuxIDMapping `json:"uidMappings,omitempty" yaml:"uidMappings,omitempty" platform:"linux"`
	GIDMappings []LinuxIDMapping `json:"gidMappings,omitempty" yaml:"gidMappings,omitempty" platform:"linux"`
}

// Hook specifies a command that is run at a particular event in the lifecycle of a container
type Hook struct {
	Path    string   `json:"path,omitempty" yaml:"path"`
	Args    []string `json:"args,omitempty" yaml:"args,omitempty"`
	Env     []string `json:"env,omitempty" yaml:"env,omitempty"`
	Timeout *int     `json:"timeout,omitempty" yaml:"timeout,omitempty"`
}

// Hooks specifies a command that is run in the container at a particular event in the lifecycle of a container
// Hooks for container setup and teardown
type Hooks struct {
	// Prestart is Deprecated. Prestart is a list of hooks to be run before the container process is executed.
	// It is called in the Runtime Namespace
	//
	// Deprecated: use [Hooks.CreateRuntime], [Hooks.CreateContainer], and
	// [Hooks.StartContainer] instead, which allow more granular hook control
	// during the create and start phase.
	Prestart []Hook `json:"prestart,omitempty" yaml:"prestart"`
	// CreateRuntime is a list of hooks to be run after the container has been created but before pivot_root or any equivalent operation has been called
	// It is called in the Runtime Namespace
	CreateRuntime []Hook `json:"createRuntime,omitempty" yaml:"createRuntime,omitempty"`
	// CreateContainer is a list of hooks to be run after the container has been created but before pivot_root or any equivalent operation has been called
	// It is called in the Container Namespace
	CreateContainer []Hook `json:"createContainer,omitempty" yaml:"createContainer,omitempty"`
	// StartContainer is a list of hooks to be run after the start operation is called but before the container process is started
	// It is called in the Container Namespace
	StartContainer []Hook `json:"startContainer,omitempty" yaml:"startContainer,omitempty"`
	// Poststart is a list of hooks to be run after the container process is started.
	// It is called in the Runtime Namespace
	Poststart []Hook `json:"poststart,omitempty" yaml:"poststart,omitempty"`
	// Poststop is a list of hooks to be run after the container process exits.
	// It is called in the Runtime Namespace
	Poststop []Hook `json:"poststop,omitempty" yaml:"poststop"`
}

// Linux contains platform-specific configuration for Linux based containers.
type Linux struct {
	// UIDMapping specifies user mappings for supporting user namespaces.
	UIDMappings []LinuxIDMapping `json:"uidMappings,omitempty" yaml:"uidMappings,omitempty"`
	// GIDMapping specifies group mappings for supporting user namespaces.
	GIDMappings []LinuxIDMapping `json:"gidMappings,omitempty" yaml:"gidMappings,omitempty"`
	// Sysctl are a set of key value pairs that are set for the container on start
	Sysctl map[string]string `json:"sysctl,omitempty" yaml:"sysctl,omitempty"`
	// Resources contain cgroup information for handling resource constraints
	// for the container
	Resources *LinuxResources `json:"resources" yaml:"resources"`
	// CgroupsPath specifies the path to cgroups that are created and/or joined by the container.
	// The path is expected to be relative to the cgroups mountpoint.
	// If resources are specified, the cgroups at CgroupsPath will be updated based on resources.
	CgroupsPath string `json:"cgroupsPath,omitempty" yaml:"cgroupsPath,omitempty"`
	// Namespaces contains the namespaces that are created and/or joined by the container
	Namespaces []LinuxNamespace `json:"namespaces,omitempty" yaml:"namespaces,omitempty"`
	// Devices are a list of device nodes that are created for the container
	Devices []LinuxDevice `json:"devices,omitempty" yaml:"devices,omitempty"`
	// Seccomp specifies the seccomp security settings for the container.
	Seccomp *LinuxSeccomp `json:"seccomp,omitempty" yaml:"seccomp,omitempty"`
	// RootfsPropagation is the rootfs mount propagation mode for the container.
	RootfsPropagation string `json:"rootfsPropagation,omitempty" yaml:"rootfsPropagation,omitempty"`
	// MaskedPaths masks over the provided paths inside the container.
	MaskedPaths []string `json:"maskedPaths,omitempty" yaml:"maskedPaths,omitempty"`
	// ReadonlyPaths sets the provided paths as RO inside the container.
	ReadonlyPaths []string `json:"readonlyPaths,omitempty" yaml:"readonlyPaths,omitempty"`
	// MountLabel specifies the selinux context for the mounts in the container.
	MountLabel string `json:"mountLabel,omitempty" yaml:"mountLabel,omitempty"`
	// IntelRdt contains Intel Resource Director Technology (RDT) information for
	// handling resource constraints and monitoring metrics (e.g., L3 cache, memory bandwidth) for the container
	IntelRdt *LinuxIntelRdt `json:"intelRdt,omitempty" yaml:"intelRdt,omitempty"`
	// Personality contains configuration for the Linux personality syscall
	Personality *LinuxPersonality `json:"personality,omitempty" yaml:"personality,omitempty"`
	// TimeOffsets specifies the offset for supporting time namespaces.
	TimeOffsets map[string]LinuxTimeOffset `json:"timeOffsets,omitempty" yaml:"timeOffsets,omitempty"`
}

// LinuxNamespace is the configuration for a Linux namespace
type LinuxNamespace struct {
	// Type is the type of namespace
	Type LinuxNamespaceType `json:"type" yaml:"type"`
	// Path is a path to an existing namespace persisted on disk that can be joined
	// and is of the same type
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
}

// LinuxNamespaceType is one of the Linux namespaces
type LinuxNamespaceType string

const (
	// PIDNamespace for isolating process IDs
	PIDNamespace LinuxNamespaceType = "pid"
	// NetworkNamespace for isolating network devices, stacks, ports, etc
	NetworkNamespace LinuxNamespaceType = "network"
	// MountNamespace for isolating mount points
	MountNamespace LinuxNamespaceType = "mount"
	// IPCNamespace for isolating System V IPC, POSIX message queues
	IPCNamespace LinuxNamespaceType = "ipc"
	// UTSNamespace for isolating hostname and NIS domain name
	UTSNamespace LinuxNamespaceType = "uts"
	// UserNamespace for isolating user and group IDs
	UserNamespace LinuxNamespaceType = "user"
	// CgroupNamespace for isolating cgroup hierarchies
	CgroupNamespace LinuxNamespaceType = "cgroup"
	// TimeNamespace for isolating the clocks
	TimeNamespace LinuxNamespaceType = "time"
)

// LinuxIDMapping specifies UID/GID mappings
type LinuxIDMapping struct {
	// ContainerID is the starting UID/GID in the container
	ContainerID uint32 `json:"containerID" yaml:"containerID"`
	// HostID is the starting UID/GID on the host to be mapped to 'ContainerID'
	HostID uint32 `json:"hostID" yaml:"hostID"`
	// Size is the number of IDs to be mapped
	Size uint32 `json:"size" yaml:"size"`
}

// LinuxTimeOffset specifies the offset for Time Namespace
type LinuxTimeOffset struct {
	// Secs is the offset of clock (in secs) in the container
	Secs int64 `json:"secs,omitempty" yaml:"secs,omitempty"`
	// Nanosecs is the additional offset for Secs (in nanosecs)
	Nanosecs uint32 `json:"nanosecs,omitempty" yaml:"nanosecs,omitempty"`
}

// POSIXRlimit type and restrictions
type POSIXRlimit struct {
	// Type of the rlimit to set
	Type string `json:"type" yaml:"type"`
	// Hard is the hard limit for the specified type
	Hard uint64 `json:"hard" yaml:"hard"`
	// Soft is the soft limit for the specified type
	Soft uint64 `json:"soft" yaml:"soft"`
}

// LinuxHugepageLimit structure corresponds to limiting kernel hugepages.
// Default to reservation limits if supported. Otherwise fallback to page fault limits.
type LinuxHugepageLimit struct {
	// Pagesize is the hugepage size.
	// Format: "<size><unit-prefix>B' (e.g. 64KB, 2MB, 1GB, etc.).
	Pagesize string `json:"pageSize" yaml:"pageSize"`
	// Limit is the limit of "hugepagesize" hugetlb reservations (if supported) or usage.
	Limit uint64 `json:"limit" yaml:"limit"`
}

// LinuxInterfacePriority for network interfaces
type LinuxInterfacePriority struct {
	// Name is the name of the network interface
	Name string `json:"name" yaml:"name"`
	// Priority for the interface
	Priority uint32 `json:"priority" yaml:"priority"`
}

// LinuxBlockIODevice holds major:minor format supported in blkio cgroup
type LinuxBlockIODevice struct {
	// Major is the device's major number.
	Major int64 `json:"major" yaml:"major"`
	// Minor is the device's minor number.
	Minor int64 `json:"minor" yaml:"minor"`
}

// LinuxWeightDevice struct holds a `major:minor weight` pair for weightDevice
type LinuxWeightDevice struct {
	LinuxBlockIODevice `yaml:",inline"`
	// Weight is the bandwidth rate for the device.
	Weight *uint16 `json:"weight,omitempty" yaml:"weight,omitempty"`
	// LeafWeight is the bandwidth rate for the device while competing with the cgroup's child cgroups, CFQ scheduler only
	LeafWeight *uint16 `json:"leafWeight,omitempty" yaml:"leafWeight,omitempty"`
}

// LinuxThrottleDevice struct holds a `major:minor rate_per_second` pair
type LinuxThrottleDevice struct {
	LinuxBlockIODevice `yaml:",inline"`
	// Rate is the IO rate limit per cgroup per device
	Rate uint64 `json:"rate" yaml:"rate"`
}

// LinuxBlockIO for Linux cgroup 'blkio' resource management
type LinuxBlockIO struct {
	// Specifies per cgroup weight
	Weight *uint16 `json:"weight,omitempty" yaml:"weight,omitempty"`
	// Specifies tasks' weight in the given cgroup while competing with the cgroup's child cgroups, CFQ scheduler only
	LeafWeight *uint16 `json:"leafWeight,omitempty" yaml:"leafWeight,omitempty"`
	// Weight per cgroup per device, can override BlkioWeight
	WeightDevice []LinuxWeightDevice `json:"weightDevice,omitempty" yaml:"weightDevice,omitempty"`
	// IO read rate limit per cgroup per device, bytes per second
	ThrottleReadBpsDevice []LinuxThrottleDevice `json:"throttleReadBpsDevice,omitempty" yaml:"throttleReadBpsDevice,omitempty"`
	// IO write rate limit per cgroup per device, bytes per second
	ThrottleWriteBpsDevice []LinuxThrottleDevice `json:"throttleWriteBpsDevice,omitempty" yaml:"throttleWriteBpsDevice,omitempty"`
	// IO read rate limit per cgroup per device, IO per second
	ThrottleReadIOPSDevice []LinuxThrottleDevice `json:"throttleReadIOPSDevice,omitempty" yaml:"throttleReadIOPSDevice,omitempty"`
	// IO write rate limit per cgroup per device, IO per second
	ThrottleWriteIOPSDevice []LinuxThrottleDevice `json:"throttleWriteIOPSDevice,omitempty" yaml:"throttleWriteIOPSDevice,omitempty"`
}

// LinuxMemory for Linux cgroup 'memory' resource management
type LinuxMemory struct {
	// Memory limit (in bytes).
	Limit *int64 `json:"limit,omitempty" yaml:"limit,omitempty"`
	// Memory reservation or soft_limit (in bytes).
	Reservation *int64 `json:"reservation,omitempty" yaml:"reservation,omitempty"`
	// Total memory limit (memory + swap).
	Swap *int64 `json:"swap,omitempty" yaml:"swap,omitempty"`
	// Kernel memory limit (in bytes).
	//
	// Deprecated: kernel-memory limits are not supported in cgroups v2, and
	// were obsoleted in [kernel v5.4]. This field should no longer be used,
	// as it may be ignored by runtimes.
	//
	// [kernel v5.4]: https://github.com/torvalds/linux/commit/0158115f702b0ba208ab0
	Kernel *int64 `json:"kernel,omitempty" yaml:"kernel,omitempty"`
	// Kernel memory limit for tcp (in bytes)
	KernelTCP *int64 `json:"kernelTCP,omitempty" yaml:"kernelTCP,omitempty"`
	// How aggressive the kernel will swap memory pages.
	Swappiness *uint64 `json:"swappiness,omitempty" yaml:"swappiness,omitempty"`
	// DisableOOMKiller disables the OOM killer for out of memory conditions
	DisableOOMKiller *bool `json:"disableOOMKiller,omitempty" yaml:"disableOOMKiller,omitempty"`
	// Enables hierarchical memory accounting
	UseHierarchy *bool `json:"useHierarchy,omitempty" yaml:"useHierarchy,omitempty"`
	// CheckBeforeUpdate enables checking if a new memory limit is lower
	// than the current usage during update, and if so, rejecting the new
	// limit.
	CheckBeforeUpdate *bool `json:"checkBeforeUpdate,omitempty" yaml:"checkBeforeUpdate,omitempty"`
}

// LinuxCPU for Linux cgroup 'cpu' resource management
type LinuxCPU struct {
	// CPU shares (relative weight (ratio) vs. other cgroups with cpu shares).
	Shares *uint64 `json:"shares,omitempty" yaml:"shares"`
	// CPU hardcap limit (in usecs). Allowed cpu time in a given period.
	Quota *int64 `json:"quota,omitempty" yaml:"quota"`
	// CPU hardcap burst limit (in usecs). Allowed accumulated cpu time additionally for burst in a
	// given period.
	Burst *uint64 `json:"burst,omitempty" yaml:"burst,omitempty"`
	// CPU period to be used for hardcapping (in usecs).
	Period *uint64 `json:"period,omitempty" yaml:"period"`
	// How much time realtime scheduling may use (in usecs).
	RealtimeRuntime *int64 `json:"realtimeRuntime,omitempty" yaml:"realtimeRuntime,omitempty"`
	// CPU period to be used for realtime scheduling (in usecs).
	RealtimePeriod *uint64 `json:"realtimePeriod,omitempty" yaml:"realtimePeriod,omitempty"`
	// CPUs to use within the cpuset. Default is to use any CPU available.
	Cpus string `json:"cpus,omitempty" yaml:"cpus,omitempty"`
	// List of memory nodes in the cpuset. Default is to use any available memory node.
	Mems string `json:"mems,omitempty" yaml:"mems,omitempty"`
	// cgroups are configured with minimum weight, 0: default behavior, 1: SCHED_IDLE.
	Idle *int64 `json:"idle,omitempty" yaml:"idle,omitempty"`
}

// LinuxPids for Linux cgroup 'pids' resource management (Linux 4.3)
type LinuxPids struct {
	// Maximum number of PIDs. Default is "no limit".
	Limit int64 `json:"limit" yaml:"limit"`
}

// LinuxNetwork identification and priority configuration
type LinuxNetwork struct {
	// Set class identifier for container's network packets
	ClassID *uint32 `json:"classID,omitempty" yaml:"classID,omitempty"`
	// Set priority of network traffic for container
	Priorities []LinuxInterfacePriority `json:"priorities,omitempty" yaml:"priorities,omitempty"`
}

// LinuxRdma for Linux cgroup 'rdma' resource management (Linux 4.11)
type LinuxRdma struct {
	// Maximum number of HCA handles that can be opened. Default is "no limit".
	HcaHandles *uint32 `json:"hcaHandles,omitempty" yaml:"hcaHandles,omitempty"`
	// Maximum number of HCA objects that can be created. Default is "no limit".
	HcaObjects *uint32 `json:"hcaObjects,omitempty" yaml:"hcaObjects,omitempty"`
}

// LinuxResources has container runtime resource constraints
type LinuxResources struct {
	// Devices configures the device allowlist.
	Devices []LinuxDeviceCgroup `json:"devices,omitempty" yaml:"devices,omitempty"`
	// Memory restriction configuration
	Memory *LinuxMemory `json:"memory,omitempty" yaml:"memory"`
	// CPU resource restriction configuration
	CPU *LinuxCPU `json:"cpu,omitempty" yaml:"cpu"`
	// Task resource restriction configuration.
	Pids *LinuxPids `json:"pids,omitempty" yaml:"pids,omitempty"`
	// BlockIO restriction configuration
	BlockIO *LinuxBlockIO `json:"blockIO,omitempty" yaml:"blockIO,omitempty"`
	// Hugetlb limits (in bytes). Default to reservation limits if supported.
	HugepageLimits []LinuxHugepageLimit `json:"hugepageLimits,omitempty" yaml:"hugepageLimits,omitempty"`
	// Network restriction configuration
	Network *LinuxNetwork `json:"network,omitempty" yaml:"network,omitempty"`
	// Rdma resource restriction configuration.
	// Limits are a set of key value pairs that define RDMA resource limits,
	// where the key is device name and value is resource limits.
	Rdma map[string]LinuxRdma `json:"rdma,omitempty" yaml:"rdma,omitempty"`
	// Unified resources.
	Unified map[string]string `json:"unified,omitempty" yaml:"unified,omitempty"`
}

// LinuxDevice represents the mknod information for a Linux special device file
type LinuxDevice struct {
	// Path to the device.
	Path string `json:"path" yaml:"path"`
	// Device type, block, char, etc.
	Type string `json:"type" yaml:"type"`
	// Major is the device's major number.
	Major int64 `json:"major" yaml:"major"`
	// Minor is the device's minor number.
	Minor int64 `json:"minor" yaml:"minor"`
	// FileMode permission bits for the device.
	FileMode *os.FileMode `json:"fileMode,omitempty" yaml:"fileMode,omitempty"`
	// UID of the device.
	UID *uint32 `json:"uid,omitempty" yaml:"uid,omitempty"`
	// Gid of the device.
	GID *uint32 `json:"gid,omitempty" yaml:"gid,omitempty"`
}

// LinuxDeviceCgroup represents a device rule for the devices specified to
// the device controller
type LinuxDeviceCgroup struct {
	// Allow or deny
	Allow bool `json:"allow" yaml:"allow"`
	// Device type, block, char, etc.
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
	// Major is the device's major number.
	Major *int64 `json:"major,omitempty" yaml:"major,omitempty"`
	// Minor is the device's minor number.
	Minor *int64 `json:"minor,omitempty" yaml:"minor,omitempty"`
	// Cgroup access permissions format, rwm.
	Access string `json:"access,omitempty" yaml:"access,omitempty"`
}

// LinuxPersonalityDomain refers to a personality domain.
type LinuxPersonalityDomain string

// LinuxPersonalityFlag refers to an additional personality flag. None are currently defined.
type LinuxPersonalityFlag string

// Define domain and flags for Personality
const (
	// PerLinux is the standard Linux personality
	PerLinux LinuxPersonalityDomain = "LINUX"
	// PerLinux32 sets personality to 32 bit
	PerLinux32 LinuxPersonalityDomain = "LINUX32"
)

// LinuxPersonality represents the Linux personality syscall input
type LinuxPersonality struct {
	// Domain for the personality
	Domain LinuxPersonalityDomain `json:"domain" yaml:"domain"`
	// Additional flags
	Flags []LinuxPersonalityFlag `json:"flags,omitempty" yaml:"flags,omitempty"`
}

// Solaris contains platform-specific configuration for Solaris application containers.
type Solaris struct {
	// SMF FMRI which should go "online" before we start the container process.
	Milestone string `json:"milestone,omitempty" yaml:"milestone,omitempty"`
	// Maximum set of privileges any process in this container can obtain.
	LimitPriv string `json:"limitpriv,omitempty" yaml:"limitpriv,omitempty"`
	// The maximum amount of shared memory allowed for this container.
	MaxShmMemory string `json:"maxShmMemory,omitempty" yaml:"maxShmMemory,omitempty"`
	// Specification for automatic creation of network resources for this container.
	Anet []SolarisAnet `json:"anet,omitempty" yaml:"anet,omitempty"`
	// Set limit on the amount of CPU time that can be used by container.
	CappedCPU *SolarisCappedCPU `json:"cappedCPU,omitempty" yaml:"cappedCPU,omitempty"`
	// The physical and swap caps on the memory that can be used by this container.
	CappedMemory *SolarisCappedMemory `json:"cappedMemory,omitempty" yaml:"cappedMemory,omitempty"`
}

// SolarisCappedCPU allows users to set limit on the amount of CPU time that can be used by container.
type SolarisCappedCPU struct {
	Ncpus string `json:"ncpus,omitempty" yaml:"ncpus,omitempty"`
}

// SolarisCappedMemory allows users to set the physical and swap caps on the memory that can be used by this container.
type SolarisCappedMemory struct {
	Physical string `json:"physical,omitempty" yaml:"physical,omitempty"`
	Swap     string `json:"swap,omitempty" yaml:"swap,omitempty"`
}

// SolarisAnet provides the specification for automatic creation of network resources for this container.
type SolarisAnet struct {
	// Specify a name for the automatically created VNIC datalink.
	Linkname string `json:"linkname,omitempty" yaml:"linkname,omitempty"`
	// Specify the link over which the VNIC will be created.
	Lowerlink string `json:"lowerLink,omitempty" yaml:"lowerLink,omitempty"`
	// The set of IP addresses that the container can use.
	Allowedaddr string `json:"allowedAddress,omitempty" yaml:"allowedAddress,omitempty"`
	// Specifies whether allowedAddress limitation is to be applied to the VNIC.
	Configallowedaddr string `json:"configureAllowedAddress,omitempty" yaml:"configureAllowedAddress,omitempty"`
	// The value of the optional default router.
	Defrouter string `json:"defrouter,omitempty" yaml:"defrouter,omitempty"`
	// Enable one or more types of link protection.
	Linkprotection string `json:"linkProtection,omitempty" yaml:"linkProtection,omitempty"`
	// Set the VNIC's macAddress
	Macaddress string `json:"macAddress,omitempty" yaml:"macAddress,omitempty"`
}

// Windows defines the runtime configuration for Windows based containers, including Hyper-V containers.
type Windows struct {
	// LayerFolders contains a list of absolute paths to directories containing image layers.
	LayerFolders []string `json:"layerFolders" yaml:"layerFolders"`
	// Devices are the list of devices to be mapped into the container.
	Devices []WindowsDevice `json:"devices,omitempty" yaml:"devices,omitempty"`
	// Resources contains information for handling resource constraints for the container.
	Resources *WindowsResources `json:"resources,omitempty" yaml:"resources,omitempty"`
	// CredentialSpec contains a JSON object describing a group Managed Service Account (gMSA) specification.
	CredentialSpec interface{} `json:"credentialSpec,omitempty" yaml:"credentialSpec,omitempty"`
	// Servicing indicates if the container is being started in a mode to apply a Windows Update servicing operation.
	Servicing bool `json:"servicing,omitempty" yaml:"servicing,omitempty"`
	// IgnoreFlushesDuringBoot indicates if the container is being started in a mode where disk writes are not flushed during its boot process.
	IgnoreFlushesDuringBoot bool `json:"ignoreFlushesDuringBoot,omitempty" yaml:"ignoreFlushesDuringBoot,omitempty"`
	// HyperV contains information for running a container with Hyper-V isolation.
	HyperV *WindowsHyperV `json:"hyperv,omitempty" yaml:"hyperv,omitempty"`
	// Network restriction configuration.
	Network *WindowsNetwork `json:"network,omitempty" yaml:"network,omitempty"`
}

// WindowsDevice represents information about a host device to be mapped into the container.
type WindowsDevice struct {
	// Device identifier: interface class GUID, etc.
	ID string `json:"id" yaml:"id"`
	// Device identifier type: "class", etc.
	IDType string `json:"idType" yaml:"idType"`
}

// WindowsResources has container runtime resource constraints for containers running on Windows.
type WindowsResources struct {
	// Memory restriction configuration.
	Memory *WindowsMemoryResources `json:"memory,omitempty" yaml:"memory,omitempty"`
	// CPU resource restriction configuration.
	CPU *WindowsCPUResources `json:"cpu,omitempty" yaml:"cpu,omitempty"`
	// Storage restriction configuration.
	Storage *WindowsStorageResources `json:"storage,omitempty" yaml:"storage,omitempty"`
}

// WindowsMemoryResources contains memory resource management settings.
type WindowsMemoryResources struct {
	// Memory limit in bytes.
	Limit *uint64 `json:"limit,omitempty" yaml:"limit,omitempty"`
}

// WindowsCPUResources contains CPU resource management settings.
type WindowsCPUResources struct {
	// Count is the number of CPUs available to the container. It represents the
	// fraction of the configured processor `count` in a container in relation
	// to the processors available in the host. The fraction ultimately
	// determines the portion of processor cycles that the threads in a
	// container can use during each scheduling interval, as the number of
	// cycles per 10,000 cycles.
	Count *uint64 `json:"count,omitempty" yaml:"count,omitempty"`
	// Shares limits the share of processor time given to the container relative
	// to other workloads on the processor. The processor `shares` (`weight` at
	// the platform level) is a value between 0 and 10000.
	Shares *uint16 `json:"shares,omitempty" yaml:"shares,omitempty"`
	// Maximum determines the portion of processor cycles that the threads in a
	// container can use during each scheduling interval, as the number of
	// cycles per 10,000 cycles. Set processor `maximum` to a percentage times
	// 100.
	Maximum *uint16 `json:"maximum,omitempty" yaml:"maximum,omitempty"`
}

// WindowsStorageResources contains storage resource management settings.
type WindowsStorageResources struct {
	// Specifies maximum Iops for the system drive.
	Iops *uint64 `json:"iops,omitempty" yaml:"iops,omitempty"`
	// Specifies maximum bytes per second for the system drive.
	Bps *uint64 `json:"bps,omitempty" yaml:"bps,omitempty"`
	// Sandbox size specifies the minimum size of the system drive in bytes.
	SandboxSize *uint64 `json:"sandboxSize,omitempty" yaml:"sandboxSize,omitempty"`
}

// WindowsNetwork contains network settings for Windows containers.
type WindowsNetwork struct {
	// List of HNS endpoints that the container should connect to.
	EndpointList []string `json:"endpointList,omitempty" yaml:"endpointList,omitempty"`
	// Specifies if unqualified DNS name resolution is allowed.
	AllowUnqualifiedDNSQuery bool `json:"allowUnqualifiedDNSQuery,omitempty" yaml:"allowUnqualifiedDNSQuery,omitempty"`
	// Comma separated list of DNS suffixes to use for name resolution.
	DNSSearchList []string `json:"DNSSearchList,omitempty" yaml:"DNSSearchList,omitempty"`
	// Name (ID) of the container that we will share with the network stack.
	NetworkSharedContainerName string `json:"networkSharedContainerName,omitempty" yaml:"networkSharedContainerName,omitempty"`
	// name (ID) of the network namespace that will be used for the container.
	NetworkNamespace string `json:"networkNamespace,omitempty" yaml:"networkNamespace,omitempty"`
}

// WindowsHyperV contains information for configuring a container to run with Hyper-V isolation.
type WindowsHyperV struct {
	// UtilityVMPath is an optional path to the image used for the Utility VM.
	UtilityVMPath string `json:"utilityVMPath,omitempty" yaml:"utilityVMPath,omitempty"`
}

// VM contains information for virtual-machine-based containers.
type VM struct {
	// Hypervisor specifies hypervisor-related configuration for virtual-machine-based containers.
	Hypervisor VMHypervisor `json:"hypervisor,omitempty" yaml:"hypervisor,omitempty"`
	// Kernel specifies kernel-related configuration for virtual-machine-based containers.
	Kernel VMKernel `json:"kernel" yaml:"kernel"`
	// Image specifies guest image related configuration for virtual-machine-based containers.
	Image VMImage `json:"image,omitempty" yaml:"image,omitempty"`
}

// VMHypervisor contains information about the hypervisor to use for a virtual machine.
type VMHypervisor struct {
	// Path is the host path to the hypervisor used to manage the virtual machine.
	Path string `json:"path" yaml:"path"`
	// Parameters specifies parameters to pass to the hypervisor.
	Parameters []string `json:"parameters,omitempty" yaml:"parameters,omitempty"`
}

// VMKernel contains information about the kernel to use for a virtual machine.
type VMKernel struct {
	// Path is the host path to the kernel used to boot the virtual machine.
	Path string `json:"path" yaml:"path"`
	// Parameters specifies parameters to pass to the kernel.
	Parameters []string `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	// InitRD is the host path to an initial ramdisk to be used by the kernel.
	InitRD string `json:"initrd,omitempty" yaml:"initrd,omitempty"`
}

// VMImage contains information about the virtual machine root image.
type VMImage struct {
	// Path is the host path to the root image that the VM kernel would boot into.
	Path string `json:"path" yaml:"path"`
	// Format is the root image format type (e.g. "qcow2", "raw", "vhd", etc).
	Format string `json:"format" yaml:"format"`
}

// LinuxSeccomp represents syscall restrictions
type LinuxSeccomp struct {
	DefaultAction    LinuxSeccompAction `json:"defaultAction" yaml:"defaultAction"`
	DefaultErrnoRet  *uint              `json:"defaultErrnoRet,omitempty" yaml:"defaultErrnoRet,omitempty"`
	Architectures    []Arch             `json:"architectures,omitempty" yaml:"architectures,omitempty"`
	Flags            []LinuxSeccompFlag `json:"flags,omitempty" yaml:"flags,omitempty"`
	ListenerPath     string             `json:"listenerPath,omitempty" yaml:"listenerPath,omitempty"`
	ListenerMetadata string             `json:"listenerMetadata,omitempty" yaml:"listenerMetadata,omitempty"`
	Syscalls         []LinuxSyscall     `json:"syscalls,omitempty" yaml:"syscalls,omitempty"`
}

// Arch used for additional architectures
type Arch string

// LinuxSeccompFlag is a flag to pass to seccomp(2).
type LinuxSeccompFlag string

const (
	// LinuxSeccompFlagLog is a seccomp flag to request all returned
	// actions except SECCOMP_RET_ALLOW to be logged. An administrator may
	// override this filter flag by preventing specific actions from being
	// logged via the /proc/sys/kernel/seccomp/actions_logged file. (since
	// Linux 4.14)
	LinuxSeccompFlagLog LinuxSeccompFlag = "SECCOMP_FILTER_FLAG_LOG"

	// LinuxSeccompFlagSpecAllow can be used to disable Speculative Store
	// Bypass mitigation. (since Linux 4.17)
	LinuxSeccompFlagSpecAllow LinuxSeccompFlag = "SECCOMP_FILTER_FLAG_SPEC_ALLOW"

	// LinuxSeccompFlagWaitKillableRecv can be used to switch to the wait
	// killable semantics. (since Linux 5.19)
	LinuxSeccompFlagWaitKillableRecv LinuxSeccompFlag = "SECCOMP_FILTER_FLAG_WAIT_KILLABLE_RECV"
)

// Additional architectures permitted to be used for system calls
// By default only the native architecture of the kernel is permitted
const (
	ArchX86         Arch = "SCMP_ARCH_X86"
	ArchX86_64      Arch = "SCMP_ARCH_X86_64"
	ArchX32         Arch = "SCMP_ARCH_X32"
	ArchARM         Arch = "SCMP_ARCH_ARM"
	ArchAARCH64     Arch = "SCMP_ARCH_AARCH64"
	ArchMIPS        Arch = "SCMP_ARCH_MIPS"
	ArchMIPS64      Arch = "SCMP_ARCH_MIPS64"
	ArchMIPS64N32   Arch = "SCMP_ARCH_MIPS64N32"
	ArchMIPSEL      Arch = "SCMP_ARCH_MIPSEL"
	ArchMIPSEL64    Arch = "SCMP_ARCH_MIPSEL64"
	ArchMIPSEL64N32 Arch = "SCMP_ARCH_MIPSEL64N32"
	ArchPPC         Arch = "SCMP_ARCH_PPC"
	ArchPPC64       Arch = "SCMP_ARCH_PPC64"
	ArchPPC64LE     Arch = "SCMP_ARCH_PPC64LE"
	ArchS390        Arch = "SCMP_ARCH_S390"
	ArchS390X       Arch = "SCMP_ARCH_S390X"
	ArchPARISC      Arch = "SCMP_ARCH_PARISC"
	ArchPARISC64    Arch = "SCMP_ARCH_PARISC64"
	ArchRISCV64     Arch = "SCMP_ARCH_RISCV64"
)

// LinuxSeccompAction taken upon Seccomp rule match
type LinuxSeccompAction string

// Define actions for Seccomp rules
const (
	ActKill        LinuxSeccompAction = "SCMP_ACT_KILL"
	ActKillProcess LinuxSeccompAction = "SCMP_ACT_KILL_PROCESS"
	ActKillThread  LinuxSeccompAction = "SCMP_ACT_KILL_THREAD"
	ActTrap        LinuxSeccompAction = "SCMP_ACT_TRAP"
	ActErrno       LinuxSeccompAction = "SCMP_ACT_ERRNO"
	ActTrace       LinuxSeccompAction = "SCMP_ACT_TRACE"
	ActAllow       LinuxSeccompAction = "SCMP_ACT_ALLOW"
	ActLog         LinuxSeccompAction = "SCMP_ACT_LOG"
	ActNotify      LinuxSeccompAction = "SCMP_ACT_NOTIFY"
)

// LinuxSeccompOperator used to match syscall arguments in Seccomp
type LinuxSeccompOperator string

// Define operators for syscall arguments in Seccomp
const (
	OpNotEqual     LinuxSeccompOperator = "SCMP_CMP_NE"
	OpLessThan     LinuxSeccompOperator = "SCMP_CMP_LT"
	OpLessEqual    LinuxSeccompOperator = "SCMP_CMP_LE"
	OpEqualTo      LinuxSeccompOperator = "SCMP_CMP_EQ"
	OpGreaterEqual LinuxSeccompOperator = "SCMP_CMP_GE"
	OpGreaterThan  LinuxSeccompOperator = "SCMP_CMP_GT"
	OpMaskedEqual  LinuxSeccompOperator = "SCMP_CMP_MASKED_EQ"
)

// LinuxSeccompArg used for matching specific syscall arguments in Seccomp
type LinuxSeccompArg struct {
	Index    uint                 `json:"index" yaml:"index"`
	Value    uint64               `json:"value" yaml:"value"`
	ValueTwo uint64               `json:"valueTwo,omitempty" yaml:"valueTwo,omitempty"`
	Op       LinuxSeccompOperator `json:"op" yaml:"op"`
}

// LinuxSyscall is used to match a syscall in Seccomp
type LinuxSyscall struct {
	Names    []string           `json:"names" yaml:"names"`
	Action   LinuxSeccompAction `json:"action" yaml:"action"`
	ErrnoRet *uint              `json:"errnoRet,omitempty" yaml:"errnoRet,omitempty"`
	Args     []LinuxSeccompArg  `json:"args,omitempty" yaml:"args,omitempty"`
}

// LinuxIntelRdt has container runtime resource constraints for Intel RDT CAT and MBA
// features and flags enabling Intel RDT CMT and MBM features.
// Intel RDT features are available in Linux 4.14 and newer kernel versions.
type LinuxIntelRdt struct {
	// The identity for RDT Class of Service
	ClosID string `json:"closID,omitempty" yaml:"closID,omitempty"`
	// The schema for L3 cache id and capacity bitmask (CBM)
	// Format: "L3:<cache_id0>=<cbm0>;<cache_id1>=<cbm1>;..."
	L3CacheSchema string `json:"l3CacheSchema,omitempty" yaml:"l3CacheSchema,omitempty"`

	// The schema of memory bandwidth per L3 cache id
	// Format: "MB:<cache_id0>=bandwidth0;<cache_id1>=bandwidth1;..."
	// The unit of memory bandwidth is specified in "percentages" by
	// default, and in "MBps" if MBA Software Controller is enabled.
	MemBwSchema string `json:"memBwSchema,omitempty" yaml:"memBwSchema,omitempty"`

	// EnableCMT is the flag to indicate if the Intel RDT CMT is enabled. CMT (Cache Monitoring Technology) supports monitoring of
	// the last-level cache (LLC) occupancy for the container.
	EnableCMT bool `json:"enableCMT,omitempty" yaml:"enableCMT,omitempty"`

	// EnableMBM is the flag to indicate if the Intel RDT MBM is enabled. MBM (Memory Bandwidth Monitoring) supports monitoring of
	// total and local memory bandwidth for the container.
	EnableMBM bool `json:"enableMBM,omitempty" yaml:"enableMBM,omitempty"`
}

// ZOS contains platform-specific configuration for z/OS based containers.
type ZOS struct {
	// Devices are a list of device nodes that are created for the container
	Devices []ZOSDevice `json:"devices,omitempty" yaml:"devices,omitempty"`
}

// ZOSDevice represents the mknod information for a z/OS special device file
type ZOSDevice struct {
	// Path to the device.
	Path string `json:"path" yaml:"path"`
	// Device type, block, char, etc.
	Type string `json:"type" yaml:"type"`
	// Major is the device's major number.
	Major int64 `json:"major" yaml:"major"`
	// Minor is the device's minor number.
	Minor int64 `json:"minor" yaml:"minor"`
	// FileMode permission bits for the device.
	FileMode *os.FileMode `json:"fileMode,omitempty" yaml:"fileMode,omitempty"`
	// UID of the device.
	UID *uint32 `json:"uid,omitempty" yaml:"uid,omitempty"`
	// Gid of the device.
	GID *uint32 `json:"gid,omitempty" yaml:"gid,omitempty"`
}

// LinuxSchedulerPolicy represents different scheduling policies used with the Linux Scheduler
type LinuxSchedulerPolicy string

const (
	// SchedOther is the default scheduling policy
	SchedOther LinuxSchedulerPolicy = "SCHED_OTHER"
	// SchedFIFO is the First-In-First-Out scheduling policy
	SchedFIFO LinuxSchedulerPolicy = "SCHED_FIFO"
	// SchedRR is the Round-Robin scheduling policy
	SchedRR LinuxSchedulerPolicy = "SCHED_RR"
	// SchedBatch is the Batch scheduling policy
	SchedBatch LinuxSchedulerPolicy = "SCHED_BATCH"
	// SchedISO is the Isolation scheduling policy
	SchedISO LinuxSchedulerPolicy = "SCHED_ISO"
	// SchedIdle is the Idle scheduling policy
	SchedIdle LinuxSchedulerPolicy = "SCHED_IDLE"
	// SchedDeadline is the Deadline scheduling policy
	SchedDeadline LinuxSchedulerPolicy = "SCHED_DEADLINE"
)

// LinuxSchedulerFlag represents the flags used by the Linux Scheduler.
type LinuxSchedulerFlag string

const (
	// SchedFlagResetOnFork represents the reset on fork scheduling flag
	SchedFlagResetOnFork LinuxSchedulerFlag = "SCHED_FLAG_RESET_ON_FORK"
	// SchedFlagReclaim represents the reclaim scheduling flag
	SchedFlagReclaim LinuxSchedulerFlag = "SCHED_FLAG_RECLAIM"
	// SchedFlagDLOverrun represents the deadline overrun scheduling flag
	SchedFlagDLOverrun LinuxSchedulerFlag = "SCHED_FLAG_DL_OVERRUN"
	// SchedFlagKeepPolicy represents the keep policy scheduling flag
	SchedFlagKeepPolicy LinuxSchedulerFlag = "SCHED_FLAG_KEEP_POLICY"
	// SchedFlagKeepParams represents the keep parameters scheduling flag
	SchedFlagKeepParams LinuxSchedulerFlag = "SCHED_FLAG_KEEP_PARAMS"
	// SchedFlagUtilClampMin represents the utilization clamp minimum scheduling flag
	SchedFlagUtilClampMin LinuxSchedulerFlag = "SCHED_FLAG_UTIL_CLAMP_MIN"
	// SchedFlagUtilClampMin represents the utilization clamp maximum scheduling flag
	SchedFlagUtilClampMax LinuxSchedulerFlag = "SCHED_FLAG_UTIL_CLAMP_MAX"
)