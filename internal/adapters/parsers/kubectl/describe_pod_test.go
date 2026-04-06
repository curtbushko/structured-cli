package kubectl_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/curtbushko/structured-cli/internal/adapters/parsers/kubectl"
)

func TestDescribePodParser_ParseCompletePod(t *testing.T) {
	input := `Name:             nginx-deployment-abc123
Namespace:        default
Priority:         0
Node:             worker-1/10.0.0.5
Start Time:       Mon, 01 Jan 2024 10:00:00 +0000
Labels:           app=nginx
                  pod-template-hash=abc123
Annotations:      kubernetes.io/created-by=controller
Status:           Running
IP:               10.244.0.5
IPs:
  IP:  10.244.0.5
Containers:
  nginx:
    Container ID:   docker://abc123
    Image:          nginx:1.21
    Image ID:       docker-pullable://nginx@sha256:abc123
    Port:           80/TCP
    Host Port:      0/TCP
    State:          Running
      Started:      Mon, 01 Jan 2024 10:00:05 +0000
    Ready:          True
    Restart Count:  0
    Limits:
      cpu:     500m
      memory:  128Mi
    Requests:
      cpu:        250m
      memory:     64Mi
    Environment:  <none>
    Mounts:
      /var/run/secrets/kubernetes.io/serviceaccount from default-token (ro)
Conditions:
  Type              Status
  Initialized       True
  Ready             True
  ContainersReady   True
  PodScheduled      True
Events:
  Type    Reason     Age   From               Message
  ----    ------     ----  ----               -------
  Normal  Scheduled  5m    default-scheduler  Successfully assigned default/nginx-deployment-abc123 to worker-1
  Normal  Pulled     5m    kubelet            Container image "nginx:1.21" already present on machine
  Normal  Created    5m    kubelet            Created container nginx
  Normal  Started    5m    kubelet            Started container nginx`

	parser := kubectl.NewDescribePodParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	podResult, ok := result.Data.(*kubectl.DescribePodResult)
	require.True(t, ok, "result.Data should be *kubectl.DescribePodResult")

	// Basic pod info
	assert.Equal(t, "nginx-deployment-abc123", podResult.Name)
	assert.Equal(t, "default", podResult.Namespace)
	assert.Equal(t, "worker-1", podResult.Node)
	assert.Equal(t, "Mon, 01 Jan 2024 10:00:00 +0000", podResult.StartTime)
	assert.Equal(t, "Running", podResult.Status)
	assert.Equal(t, "10.244.0.5", podResult.IP)

	// Labels
	require.Len(t, podResult.Labels, 2)
	assert.Equal(t, "nginx", podResult.Labels["app"])
	assert.Equal(t, "abc123", podResult.Labels["pod-template-hash"])

	// Containers
	require.Len(t, podResult.Containers, 1)
	container := podResult.Containers[0]
	assert.Equal(t, "nginx", container.Name)
	assert.Equal(t, "nginx:1.21", container.Image)
	assert.Equal(t, "Running", container.State)
	assert.True(t, container.Ready)
	assert.Equal(t, 0, container.RestartCount)
	assert.Equal(t, "500m", container.Limits["cpu"])
	assert.Equal(t, "128Mi", container.Limits["memory"])
	assert.Equal(t, "250m", container.Requests["cpu"])
	assert.Equal(t, "64Mi", container.Requests["memory"])

	// Conditions
	require.Len(t, podResult.Conditions, 4)
	assert.Equal(t, "Initialized", podResult.Conditions[0].Type)
	assert.Equal(t, "True", podResult.Conditions[0].Status)
	assert.Equal(t, "Ready", podResult.Conditions[1].Type)

	// Events
	require.Len(t, podResult.Events, 4)
	assert.Equal(t, "Normal", podResult.Events[0].Type)
	assert.Equal(t, "Scheduled", podResult.Events[0].Reason)
	assert.Equal(t, "5m", podResult.Events[0].Age)
	assert.Equal(t, "default-scheduler", podResult.Events[0].From)
	assert.Contains(t, podResult.Events[0].Message, "Successfully assigned")
}

func TestDescribePodParser_ParseMultipleContainers(t *testing.T) {
	input := `Name:             multi-container-pod
Namespace:        default
Node:             worker-1/10.0.0.5
Start Time:       Mon, 01 Jan 2024 10:00:00 +0000
Labels:           app=multi
Status:           Running
IP:               10.244.0.10
Containers:
  app:
    Image:          myapp:latest
    State:          Running
    Ready:          True
    Restart Count:  0
  sidecar:
    Image:          sidecar:v1
    State:          Running
    Ready:          True
    Restart Count:  2
    Limits:
      cpu:     100m
      memory:  64Mi
Conditions:
  Type    Status
  Ready   True
Events:
  Type    Reason   Age   From     Message
  ----    ------   ----  ----     -------
  Normal  Started  1m    kubelet  Started container app`

	parser := kubectl.NewDescribePodParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	podResult, ok := result.Data.(*kubectl.DescribePodResult)
	require.True(t, ok)

	// Should have two containers
	require.Len(t, podResult.Containers, 2)

	assert.Equal(t, "app", podResult.Containers[0].Name)
	assert.Equal(t, "myapp:latest", podResult.Containers[0].Image)
	assert.Equal(t, 0, podResult.Containers[0].RestartCount)

	assert.Equal(t, "sidecar", podResult.Containers[1].Name)
	assert.Equal(t, "sidecar:v1", podResult.Containers[1].Image)
	assert.Equal(t, 2, podResult.Containers[1].RestartCount)
	assert.Equal(t, "100m", podResult.Containers[1].Limits["cpu"])
}

func TestDescribePodParser_ParsePodWithEvents(t *testing.T) {
	input := `Name:             event-pod
Namespace:        default
Node:             worker-1/10.0.0.5
Start Time:       Mon, 01 Jan 2024 10:00:00 +0000
Labels:           <none>
Status:           Running
IP:               10.244.0.15
Containers:
  app:
    Image:          nginx:latest
    State:          Running
    Ready:          True
    Restart Count:  0
Conditions:
  Type    Status
  Ready   True
Events:
  Type     Reason     Age   From               Message
  ----     ------     ----  ----               -------
  Normal   Scheduled  10m   default-scheduler  Successfully assigned
  Normal   Pulling    10m   kubelet            Pulling image "nginx:latest"
  Normal   Pulled     9m    kubelet            Successfully pulled image
  Normal   Created    9m    kubelet            Created container app
  Normal   Started    9m    kubelet            Started container app
  Warning  Unhealthy  5m    kubelet            Readiness probe failed`

	parser := kubectl.NewDescribePodParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	podResult, ok := result.Data.(*kubectl.DescribePodResult)
	require.True(t, ok)

	// Should have 6 events
	require.Len(t, podResult.Events, 6)

	// Check warning event
	warningEvent := podResult.Events[5]
	assert.Equal(t, "Warning", warningEvent.Type)
	assert.Equal(t, "Unhealthy", warningEvent.Reason)
	assert.Contains(t, warningEvent.Message, "Readiness probe failed")
}

func TestDescribePodParser_ParsePodWithConditions(t *testing.T) {
	input := `Name:             condition-pod
Namespace:        default
Node:             worker-1/10.0.0.5
Start Time:       Mon, 01 Jan 2024 10:00:00 +0000
Labels:           <none>
Status:           Pending
IP:               <none>
Containers:
  app:
    Image:          nginx:latest
    State:          Waiting
    Ready:          False
    Restart Count:  0
Conditions:
  Type              Status
  Initialized       True
  Ready             False
  ContainersReady   False
  PodScheduled      True
Events:              <none>`

	parser := kubectl.NewDescribePodParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	podResult, ok := result.Data.(*kubectl.DescribePodResult)
	require.True(t, ok)

	// Should have 4 conditions
	require.Len(t, podResult.Conditions, 4)

	// Check Ready condition is False
	readyCondition := podResult.Conditions[1]
	assert.Equal(t, "Ready", readyCondition.Type)
	assert.Equal(t, "False", readyCondition.Status)
}

func TestDescribePodParser_HandleNoEvents(t *testing.T) {
	input := `Name:             no-events-pod
Namespace:        default
Node:             worker-1/10.0.0.5
Start Time:       Mon, 01 Jan 2024 10:00:00 +0000
Labels:           <none>
Status:           Running
IP:               10.244.0.20
Containers:
  app:
    Image:          nginx:latest
    State:          Running
    Ready:          True
    Restart Count:  0
Conditions:
  Type    Status
  Ready   True
Events:   <none>`

	parser := kubectl.NewDescribePodParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)
	require.Nil(t, result.Error)

	podResult, ok := result.Data.(*kubectl.DescribePodResult)
	require.True(t, ok)

	// Events should be empty, not nil
	assert.NotNil(t, podResult.Events)
	assert.Empty(t, podResult.Events)
}

func TestDescribePodParser_Matches(t *testing.T) {
	parser := kubectl.NewDescribePodParser()

	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{"kubectl describe pod", "kubectl", []string{"describe", "pod"}, true},
		{"kubectl describe pod with name", "kubectl", []string{"describe", "pod", "my-pod"}, true},
		{"kubectl describe pod with flags", "kubectl", []string{"describe", "pod", "-n", "default"}, true},
		{"kubectl describe pods", "kubectl", []string{"describe", "pods"}, true},
		{"kubectl describe po", "kubectl", []string{"describe", "po"}, true},
		{"kubectl describe node", "kubectl", []string{"describe", "node"}, false},
		{"kubectl describe service", "kubectl", []string{"describe", "service"}, false},
		{"kubectl get pods", "kubectl", []string{"get", "pods"}, false},
		{"kubectl only", "kubectl", []string{}, false},
		{"kubectl describe only", "kubectl", []string{"describe"}, false},
		{"docker describe pod", "docker", []string{"describe", "pod"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDescribePodParser_Schema(t *testing.T) {
	parser := kubectl.NewDescribePodParser()
	schema := parser.Schema()

	assert.NotEmpty(t, schema.ID)
	assert.NotEmpty(t, schema.Title)
	assert.Equal(t, "object", schema.Type)
	assert.Contains(t, schema.Properties, "name")
	assert.Contains(t, schema.Properties, "namespace")
	assert.Contains(t, schema.Properties, "containers")
	assert.Contains(t, schema.Properties, "conditions")
	assert.Contains(t, schema.Properties, "events")
}
