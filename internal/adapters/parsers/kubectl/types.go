// Package kubectl provides parsers for kubectl command output.
// This package is in the adapters layer and implements parsers for
// converting raw kubectl command output into structured domain types.
package kubectl

// Pod represents a Kubernetes pod from kubectl get pods output.
type Pod struct {
	// Name is the pod name.
	Name string `json:"name"`

	// Namespace is the pod's namespace (when -A or --all-namespaces is used).
	Namespace string `json:"namespace,omitempty"`

	// Ready is the ready status (e.g., "1/1", "0/1").
	Ready string `json:"ready"`

	// Status is the pod status (e.g., "Running", "Pending", "CrashLoopBackOff").
	Status string `json:"status"`

	// Restarts is the number of container restarts.
	Restarts int `json:"restarts"`

	// Age is the pod age (e.g., "5d", "10m", "1h").
	Age string `json:"age"`

	// IP is the pod IP address (shown with -o wide).
	IP string `json:"ip,omitempty"`

	// Node is the node the pod is running on (shown with -o wide).
	Node string `json:"node,omitempty"`
}

// GetPodsResult represents the structured output of 'kubectl get pods'.
type GetPodsResult struct {
	// Pods is the list of pods.
	Pods []Pod `json:"pods"`
}

// Port represents a service port from kubectl get services output.
type Port struct {
	// Port is the service port number.
	Port int `json:"port"`

	// NodePort is the node port number (for NodePort and LoadBalancer services).
	NodePort int `json:"node_port,omitempty"`

	// Protocol is the protocol (TCP, UDP, SCTP).
	Protocol string `json:"protocol"`
}

// Service represents a Kubernetes service from kubectl get services output.
type Service struct {
	// Name is the service name.
	Name string `json:"name"`

	// Namespace is the service's namespace (when -A or --all-namespaces is used).
	Namespace string `json:"namespace,omitempty"`

	// Type is the service type (ClusterIP, NodePort, LoadBalancer, ExternalName).
	Type string `json:"type"`

	// ClusterIP is the cluster-internal IP address.
	ClusterIP string `json:"cluster_ip"`

	// ExternalIP is the external IP address (for LoadBalancer services).
	ExternalIP string `json:"external_ip,omitempty"`

	// Ports is the list of exposed ports.
	Ports []Port `json:"ports"`

	// Age is the service age (e.g., "5d", "10m", "1h").
	Age string `json:"age"`
}

// GetServicesResult represents the structured output of 'kubectl get services'.
type GetServicesResult struct {
	// Services is the list of services.
	Services []Service `json:"services"`
}

// Deployment represents a Kubernetes deployment from kubectl get deployments output.
type Deployment struct {
	// Name is the deployment name.
	Name string `json:"name"`

	// Namespace is the deployment's namespace (when -A or --all-namespaces is used).
	Namespace string `json:"namespace,omitempty"`

	// Ready is the ready status (e.g., "3/3", "0/1").
	Ready string `json:"ready"`

	// ReadyCount is the number of ready replicas.
	ReadyCount int `json:"ready_count"`

	// DesiredCount is the number of desired replicas.
	DesiredCount int `json:"desired_count"`

	// UpToDate is the number of up-to-date replicas.
	UpToDate int `json:"up_to_date"`

	// Available is the number of available replicas.
	Available int `json:"available"`

	// Age is the deployment age (e.g., "5d", "10m", "1h").
	Age string `json:"age"`
}

// GetDeploymentsResult represents the structured output of 'kubectl get deployments'.
type GetDeploymentsResult struct {
	// Deployments is the list of deployments.
	Deployments []Deployment `json:"deployments"`
}

// Node represents a Kubernetes node from kubectl get nodes output.
type Node struct {
	// Name is the node name.
	Name string `json:"name"`

	// Status is the node status (e.g., "Ready", "NotReady").
	Status string `json:"status"`

	// Roles is the list of node roles (e.g., ["control-plane", "master"]).
	Roles []string `json:"roles"`

	// Age is the node age (e.g., "5d", "10m", "1h").
	Age string `json:"age"`

	// Version is the Kubernetes version (e.g., "v1.28.0").
	Version string `json:"version"`

	// InternalIP is the internal IP address (shown with -o wide).
	InternalIP string `json:"internal_ip,omitempty"`

	// ExternalIP is the external IP address (shown with -o wide).
	ExternalIP string `json:"external_ip,omitempty"`

	// OSImage is the OS image (shown with -o wide).
	OSImage string `json:"os_image,omitempty"`

	// KernelVersion is the kernel version (shown with -o wide).
	KernelVersion string `json:"kernel_version,omitempty"`

	// ContainerRuntime is the container runtime (shown with -o wide).
	ContainerRuntime string `json:"container_runtime,omitempty"`
}

// GetNodesResult represents the structured output of 'kubectl get nodes'.
type GetNodesResult struct {
	// Nodes is the list of nodes.
	Nodes []Node `json:"nodes"`
}
