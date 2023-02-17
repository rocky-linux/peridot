// Copyright (c) All respective contributors to the Peridot Project. All rights reserved.
// Copyright (c) 2021-2022 Rocky Enterprise Software Foundation, Inc. All rights reserved.
// Copyright (c) 2021-2022 Ctrl IQ, Inc. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
// 1. Redistributions of source code must retain the above copyright notice,
// this list of conditions and the following disclaimer.
//
// 2. Redistributions in binary form must reproduce the above copyright notice,
// this list of conditions and the following disclaimer in the documentation
// and/or other materials provided with the distribution.
//
// 3. Neither the name of the copyright holder nor the names of its contributors
// may be used to endorse or promote products derived from this software without
// specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
// AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
// IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
// ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
// LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
// CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
// SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
// INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
// CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
// ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
// POSSIBILITY OF SUCH DAMAGE.

package workflow

import (
	"bufio"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"peridot.resf.org/peridot/db/models"
	peridotpb "peridot.resf.org/peridot/pb"
	"peridot.resf.org/utils"
)

type ProvisionWorkerRequest struct {
	TaskId        string         `json:"taskId"`
	ParentTaskId  sql.NullString `json:"parentTaskId"`
	Purpose       string         `json:"purpose"`
	Arch          string         `json:"arch"`
	ImageArch     string         `json:"imageArch"`
	BuildPoolType string         `json:"buildPoolType"`
	ProjectId     string         `json:"projectId"`
	HighResource  bool           `json:"highResource"`
	Privileged    bool           `json:"privileged"`
}

func archToGoArch(arch string) string {
	if arch == "aarch64" {
		arch = "arm64"
	} else if arch == "i686" || arch == "386" || arch == "x86_64" {
		arch = "amd64"
	}

	return arch
}

func goArchToArch(arch string) string {
	switch arch {
	case "arm64":
		return "aarch64"
	case "ppc64le":
		return "ppc64le"
	case "s390x":
		return "s390x"
	default:
		return "x86_64"
	}
}

func buildPoolArch(goArch string, req *ProvisionWorkerRequest) (result string) {
	result = goArch
	if len(req.BuildPoolType) > 0 {
		result = result + "-" + req.BuildPoolType
	}
	return result
}

func (c *Controller) genNameWorker(buildID, purpose string) string {
	return strings.ReplaceAll(fmt.Sprintf("pb-%s-%s", buildID, purpose), "_", "-")
}

// provisionWorker provisions a new ephemeral worker
// to do so it utilizes a child workflow and returns a function that should
// be deferred from the build workflow to ensure that the provisioned worker
// is only active for the time of the parent workflow
func (c *Controller) provisionWorker(ctx workflow.Context, req *ProvisionWorkerRequest) (string, func(), error) {
	queue := c.mainQueue

	var project *models.Project
	if req.ProjectId != "" {
		projects, err := c.db.ListProjects(&peridotpb.ProjectFilters{
			Id: wrapperspb.String(req.ProjectId),
		})
		if err != nil {
			return "", nil, fmt.Errorf("could not list projects: %v", err)
		}
		project = &projects[0]
	} else {
		project = &models.Project{
			Archs: []string{goArchToArch(runtime.GOARCH)},
		}
	}

	// Normalize arch string
	imageArch := req.Arch
	if imageArch == "noarch" {
		// For now limit where we schedule noarch
		// todo(mustafa): Make this configurable from project settings
		var filteredArches []string
		for _, arch := range project.Archs {
			if arch == "x86_64" || arch == "aarch64" {
				filteredArches = append(filteredArches, arch)
			}
		}
		if len(filteredArches) == 0 {
			filteredArches = project.Archs
		}
		imageArch = filteredArches[rand.Intn(len(filteredArches))]
	}

	// Since we assign a random image for noarch, just make sure correct imageArch
	// is given for i686
	if imageArch == "i686" {
		imageArch = "x86_64"
	}

	// s390x and ppc64le is considered extarches and runs in provision only mode
	switch imageArch {
	case "s390x",
		"ppc64le":
		queue = "peridot-provision-only-extarches"
	}

	var podName string
	isKilled := false
	deletePod := func() {
		// since the context may be terminated by this point, creating a disconnected
		// context is recommended by the temporal team
		if !isKilled {
			ctx, _ := workflow.NewDisconnectedContext(ctx)
			workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
				ParentClosePolicy: enums.PARENT_CLOSE_POLICY_ABANDON,
				TaskQueue:         queue,
			})
			err := workflow.ExecuteChildWorkflow(ctx, c.DestroyWorkerWorkflow, req).Get(ctx, nil)
			if err != nil {
				logrus.Errorf("could not destroy ephemeral worker, %s may linger around until peridot-janitor picks it up", podName)
			}
			isKilled = true
		}
	}

	ctx = workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		TaskQueue: queue,
	})
	req.ImageArch = imageArch
	if project.BuildPoolType.Valid {
		req.BuildPoolType = project.BuildPoolType.String
	} else {
		req.BuildPoolType = ""
	}
	err := workflow.ExecuteChildWorkflow(ctx, c.ProvisionWorkerWorkflow, req, queue).Get(ctx, &podName)
	if err != nil {
		var applicationErr *temporal.ApplicationError
		if errors.As(err, &applicationErr) {
			if applicationErr.Error() == "pod failed" {
				_ = applicationErr.Details(&podName)
				deletePod()
			}
		}
		return "", nil, err
	}

	workflow.Go(ctx, func(ctx workflow.Context) {
		newCtx, _ := workflow.NewDisconnectedContext(ctx)
		newCtx = workflow.WithActivityOptions(newCtx, workflow.ActivityOptions{
			ScheduleToStartTimeout: 10 * time.Hour,
			StartToCloseTimeout:    30 * time.Hour,
			HeartbeatTimeout:       15 * time.Second,
			TaskQueue:              queue,
		})
		_ = workflow.ExecuteActivity(newCtx, c.IngestLogsActivity, podName, req.TaskId, req.ParentTaskId).Get(ctx, nil)
	})

	return podName, deletePod, nil
}

// ProvisionWorkerWorkflow provisions a new job specific container using
// the provided ephemeral provisioner.
// Returns an identifier
func (c *Controller) ProvisionWorkerWorkflow(ctx workflow.Context, req *ProvisionWorkerRequest, queue string) (string, error) {
	var task models.Task
	taskSideEffect := workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
		var projectId *string
		if req.ProjectId != "" {
			projectId = &req.ProjectId
		}
		task, err := c.db.CreateTask(nil, "noarch", peridotpb.TaskType_TASK_TYPE_WORKER_PROVISION, projectId, &req.TaskId)
		if err != nil {
			return nil
		}

		return task
	})
	err := taskSideEffect.Get(&task)
	if err != nil {
		return "", err
	}
	if task.ID.String() == "" {
		return "", fmt.Errorf("could not create task")
	}

	err = c.db.SetTaskStatus(task.ID.String(), peridotpb.TaskStatus_TASK_STATUS_RUNNING)
	if err != nil {
		return "", err
	}

	defer func() {
		err := c.db.SetTaskStatus(task.ID.String(), task.Status)
		if err != nil {
			c.log.Errorf("could not set task status in ProvisionWorkerWorkflow: %v", err)
		}
	}()

	// should fall back to FAILED in case it actually fails before we
	// can set it to SUCCEEDED
	task.Status = peridotpb.TaskStatus_TASK_STATUS_FAILED

	options := workflow.ActivityOptions{
		ScheduleToStartTimeout: 10 * time.Hour,
		StartToCloseTimeout:    10 * time.Hour,
		HeartbeatTimeout:       10 * time.Second,
		TaskQueue:              queue,
	}
	ctx = workflow.WithActivityOptions(ctx, options)

	var podName string
	if err := workflow.ExecuteActivity(ctx, c.CreateK8sPodActivity, req, task).Get(ctx, &podName); err != nil {
		return "", err
	}

	task.Status = peridotpb.TaskStatus_TASK_STATUS_SUCCEEDED

	return podName, nil
}

// DestroyWorkerWorkflow destroys the created ephemeral worker
func (c *Controller) DestroyWorkerWorkflow(ctx workflow.Context, req *ProvisionWorkerRequest) error {
	var task models.Task
	taskSideEffect := workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
		var projectId *string
		if req.ProjectId != "" {
			projectId = &req.ProjectId
		}
		task, err := c.db.CreateTask(nil, "noarch", peridotpb.TaskType_TASK_TYPE_WORKER_DESTROY, projectId, &req.TaskId)
		if err != nil {
			return nil
		}

		return task
	})
	err := taskSideEffect.Get(&task)
	if err != nil {
		return err
	}
	if task.ID.String() == "" {
		return fmt.Errorf("could not create task")
	}

	err = c.db.SetTaskStatus(task.ID.String(), peridotpb.TaskStatus_TASK_STATUS_RUNNING)
	if err != nil {
		return err
	}

	defer func() {
		err := c.db.SetTaskStatus(task.ID.String(), task.Status)
		if err != nil {
			c.log.Errorf("could not set task status in ProvisionWorkerWorkflow: %v", err)
		}
	}()

	// should fall back to FAILED in case it actually fails before we
	// can set it to SUCCEEDED
	task.Status = peridotpb.TaskStatus_TASK_STATUS_FAILED

	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToStartTimeout: 10 * time.Hour,
		StartToCloseTimeout:    10 * time.Hour,
		HeartbeatTimeout:       10 * time.Second,
		TaskQueue:              c.mainQueue,
	})
	_ = workflow.ExecuteActivity(ctx, c.DeleteK8sPodActivity, req, task).Get(ctx, nil)

	task.Status = peridotpb.TaskStatus_TASK_STATUS_SUCCEEDED

	return nil
}

// DeleteK8sPodActivity deletes the pod that hosts the ephemeral worker
func (c *Controller) DeleteK8sPodActivity(ctx context.Context, req *ProvisionWorkerRequest, task *models.Task) error {
	stopChan := makeHeartbeat(ctx, 3*time.Second)
	defer func() { stopChan <- true }()

	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	kctx := context.TODO()
	name := c.genNameWorker(req.TaskId, req.Purpose)

	_ = c.logToMon(
		[]string{fmt.Sprintf("Deleting pod %s", name)},
		task.ID.String(),
		req.ParentTaskId.String,
	)

	podInterface := clientSet.CoreV1().Pods(utils.GetKubeNS())
	err = podInterface.Delete(kctx, name, metav1.DeleteOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			_ = c.logToMon(
				[]string{fmt.Sprintf("Pod %s already deleted, skipping", name)},
				task.ID.String(),
				req.ParentTaskId.String,
			)
			return nil
		}
		return err
	}

	return nil
}

// IngestLogsActivity ingests the logs of the ephemeral worker
// This is NOT running on the ephemeral worker itself, but the peridotephemeral service
func (c *Controller) IngestLogsActivity(ctx context.Context, podName string, taskId string, parentTaskId sql.NullString) error {
	var sinceTime metav1.Time
	go func() {
		for {
			activity.RecordHeartbeat(ctx, sinceTime)
			time.Sleep(4 * time.Second)
		}
	}()
	if activity.HasHeartbeatDetails(ctx) {
		_ = activity.GetHeartbeatDetails(ctx, &sinceTime)
	}

	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		logrus.Errorf("failed to create in-cluster config: %v", err)
		return err
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		logrus.Errorf("failed to create k8s client: %v", err)
	}

	podInterface := clientSet.CoreV1().Pods(utils.GetKubeNS())

	var parentTask string
	if parentTaskId.Valid {
		parentTask = parentTaskId.String
	} else {
		parentTask = taskId
	}

	for {
		_, err := podInterface.Get(context.TODO(), podName, metav1.GetOptions{})
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				break
			}
			time.Sleep(time.Second)
			continue
		}

		logReq := podInterface.GetLogs(podName, &v1.PodLogOptions{
			Container: "builder",
			Follow:    true,
			SinceTime: &sinceTime,
		})
		kctx := context.TODO()
		podLogs, err := logReq.Stream(kctx)
		if err != nil {
			logrus.Errorf("failed to get logs for pod %s: %v", podName, err)
			return err
		}

		scanner := bufio.NewScanner(podLogs)
		var lines []string
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "unable to retrieve container logs for") {
				continue
			}
			if strings.Contains(line, "Failed to poll for task") {
				continue
			}
			if strings.HasPrefix(line, "failed to try resolving symlinks in path") {
				continue
			}
			lines = append(lines, line)
			if len(lines) > 5 {
				if err := c.logToMon(lines, taskId, parentTask); err != nil {
					c.log.Errorf("could not log to mon: %v", err)
				}
				lines = []string{}
			}
		}
		if len(lines) > 0 {
			if err := c.logToMon(lines, taskId, parentTask); err != nil {
				c.log.Errorf("could not log to mon: %v", err)
			}
		}
		sinceTime = metav1.Time{Time: time.Now()}

		err = podLogs.Close()
		if err != nil {
			logrus.Errorf("failed to close pod logs: %v", err)
		}
	}

	return nil
}

// CreateK8sPodActivity creates a new pod in the same namespace
// with the specified container (the container has to contain peridotbuilder)
func (c *Controller) CreateK8sPodActivity(ctx context.Context, req *ProvisionWorkerRequest, task *models.Task) (string, error) {
	stopChan := makeHeartbeat(ctx, 3*time.Second)
	defer func() { stopChan <- true }()

	task.Status = peridotpb.TaskStatus_TASK_STATUS_FAILED

	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return "", err
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return "", err
	}

	production := "false"
	if viper.GetBool("production") {
		production = "true"
	}

	imageArch := req.ImageArch
	goArch := archToGoArch(imageArch)
	nodePoolArch := buildPoolArch(goArch, req)

	_ = c.logToMon([]string{
		fmt.Sprintf("Creating worker for purpose %s", req.Purpose),
		fmt.Sprintf("Image arch: %s", imageArch),
		fmt.Sprintf("Go arch: %s", goArch),
	}, task.ID.String(), req.ParentTaskId.String)

	image := viper.GetString(fmt.Sprintf("builder-oci-image-%s", imageArch))
	if image == "" {
		return "", fmt.Errorf("could not find builder-oci-image for arch %s", imageArch)
	}

	imagePullPolicy := v1.PullIfNotPresent
	if !viper.GetBool("production") {
		imagePullPolicy = v1.PullAlways
	}

	podInterface := clientSet.CoreV1().Pods(utils.GetKubeNS())
	name := c.genNameWorker(req.TaskId, req.Purpose)

	args := []string{
		"DATABASE_URL=" + viper.GetString("database.url"),
		"PRODUCTION=" + production,
		"TASK_QUEUE=" + name,
		"PROJECT_ID=" + req.ProjectId,
		"TASK_ID=" + req.TaskId,
		"TEMPORAL_HOSTPORT=" + viper.GetString("temporal.hostport"),
	}

	if req.ParentTaskId.Valid {
		args = append(args, "PARENT_TASK_ID="+req.ParentTaskId.String)
	}

	if accessKey := viper.GetString("s3-access-key"); accessKey != "" {
		args = append(args, "S3_ACCESS_KEY="+accessKey, "S3_SECRET_KEY="+viper.GetString("s3-secret-key"))
	}

	if endpoint := viper.GetString("s3-endpoint"); endpoint != "" {
		args = append(args, "S3_ENDPOINT="+endpoint)
	}

	if region := viper.GetString("s3-region"); region != "" {
		args = append(args, "S3_REGION="+region)
	}

	if bucket := viper.GetString("s3-bucket"); bucket != "" {
		args = append(args, "S3_BUCKET="+bucket)
	}

	if disableSsl := viper.GetBool("s3-disable-ssl"); disableSsl {
		args = append(args, "S3_DISABLE_SSL=true")
	}

	if forcePathStyle := viper.GetBool("s3-force-path-style"); forcePathStyle {
		args = append(args, "S3_FORCE_PATH_STYLE=true")
	}

	var imagePullSecrets []v1.LocalObjectReference
	if secret := os.Getenv("IMAGE_PULL_SECRET"); secret != "" {
		imagePullSecrets = []v1.LocalObjectReference{
			{
				Name: secret,
			},
		}
	}

	memoryQuantity, err := resource.ParseQuantity("4Gi")
	if err != nil {
		return "", err
	}
	ephemeralQuantity, err := resource.ParseQuantity("10Gi")
	if err != nil {
		return "", err
	}
	cpuQuantity, err := resource.ParseQuantity("1")
	if err != nil {
		return "", err
	}
	resourceRequirements := v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceCPU:              cpuQuantity,
			v1.ResourceMemory:           memoryQuantity,
			v1.ResourceEphemeralStorage: ephemeralQuantity,
		},
	}

	// For now disable resource requirements in dev
	if os.Getenv("LOCALSTACK_ENDPOINT") != "" || !req.HighResource {
		resourceRequirements = v1.ResourceRequirements{}
	}

	command := fmt.Sprintf("/bundle/peridotbuilder_%s", goArch)
	podConfig := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: utils.GetKubeNS(),
			Labels: map[string]string{
				"peridot.rockylinux.org/managed-by":              "peridotephemeral",
				"peridot.rockylinux.org/task-id":                 req.TaskId,
				"peridot.rockylinux.org/name":                    name,
				"peridot.rockylinux.org/workflow-tolerates-arch": nodePoolArch,
				// todo(mustafa): Implement janitor (cron workflow?)
				"janitor.peridot.rockylinux.org/allow-cleanup":   "yes",
				"janitor.peridot.rockylinux.org/cleanup-timeout": "864000s",
			},
			Annotations: map[string]string{
				"cluster-autoscaler.kubernetes.io/safe-to-evict": "false",
			},
		},
		Spec: v1.PodSpec{
			ServiceAccountName: os.Getenv("RESF_SERVICE_ACCOUNT"),
			ImagePullSecrets:   imagePullSecrets,
			Containers: []v1.Container{
				{
					Name:  "builder",
					Image: image,
					Args:  []string{command},
					Env: []v1.EnvVar{
						{
							Name:  "RESF_ENV",
							Value: os.Getenv("RESF_ENV"),
						},
						{
							Name:  "RESF_NS",
							Value: os.Getenv("RESF_NS"),
						},
						{
							Name:  "RESF_FORCE_NS",
							Value: os.Getenv("RESF_FORCE_NS"),
						},
						{
							Name:  "LOCALSTACK_ENDPOINT",
							Value: os.Getenv("LOCALSTACK_ENDPOINT"),
						},
						{
							Name:  "REAL_BUILD_ARCH",
							Value: imageArch,
						},
						{
							Name:  "TEMPORAL_NAMESPACE",
							Value: viper.GetString("temporal.namespace"),
						},
						{
							Name:  "KEYKEEPER_GRPC_ENDPOINT_OVERRIDE",
							Value: os.Getenv("KEYKEEPER_GRPC_ENDPOINT_OVERRIDE"),
						},
						{
							Name:  "YUMREPOFS_HTTP_ENDPOINT_OVERRIDE",
							Value: os.Getenv("YUMREPOFS_HTTP_ENDPOINT_OVERRIDE"),
						},
					},
					// todo(mustafa): Figure out good limitations
					// todo(mustafa): We will probably generate analysis based upper limits
					// todo(mustafa): Defaulting to 1cpu/4GiB for packages without metrics
					Resources:       resourceRequirements,
					ImagePullPolicy: imagePullPolicy,
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      "urandom",
							MountPath: "/dev/random",
						},
						{
							Name:      "security-limits",
							MountPath: "/etc/security/limits.conf",
						},
					},
				},
			},
			Volumes: []v1.Volume{
				{
					Name: "urandom",
					VolumeSource: v1.VolumeSource{
						HostPath: &v1.HostPathVolumeSource{
							Path: "/dev/urandom",
						},
					},
				},
				{
					Name: "security-limits",
					VolumeSource: v1.VolumeSource{
						HostPath: &v1.HostPathVolumeSource{
							Path: "/etc/security/limits.conf",
						},
					},
				},
			},
			Tolerations: []v1.Toleration{
				{
					Key:      "peridot.rockylinux.org/workflow-tolerates-arch",
					Operator: v1.TolerationOpEqual,
					Value:    nodePoolArch,
					Effect:   v1.TaintEffectNoSchedule,
				},
			},
			RestartPolicy: v1.RestartPolicyOnFailure,
		},
	}
	for _, arg := range args {
		envSplit := strings.SplitN(arg, "=", 2)
		if len(envSplit) != 2 {
			continue
		}
		podConfig.Spec.Containers[0].Env = append(podConfig.Spec.Containers[0].Env, v1.EnvVar{
			Name:  fmt.Sprintf("PERIDOTBUILDER_%s", envSplit[0]),
			Value: envSplit[1],
		})
	}
	if aki := os.Getenv("AWS_ACCESS_KEY_ID"); aki != "" {
		podConfig.Spec.Containers[0].Env = append(podConfig.Spec.Containers[0].Env, v1.EnvVar{
			Name:  "AWS_ACCESS_KEY_ID",
			Value: aki,
		})
		podConfig.Spec.Containers[0].Env = append(podConfig.Spec.Containers[0].Env, v1.EnvVar{
			Name:  "AWS_SECRET_ACCESS_KEY",
			Value: os.Getenv("AWS_SECRET_ACCESS_KEY"),
		})
	}
	if os.Getenv("PERIDOT_SITE") == "extarches" {
		podConfig.Spec.DNSPolicy = v1.DNSNone
		podConfig.Spec.DNSConfig = &v1.PodDNSConfig{
			Nameservers: []string{
				"1.1.1.1",
				"1.0.0.1",
			},
			Searches: []string{
				"peridotephemeral.svc.cluster.local",
			},
		}
	}

	// Build containers may require a privileged container
	// to do necessary mounting within the chroot.
	// The privileges are dropped soon after
	if req.Privileged {
		podConfig.Spec.Containers[0].SecurityContext = &v1.SecurityContext{
			RunAsUser:                utils.Pointer[int64](0),
			RunAsGroup:               utils.Pointer[int64](0),
			RunAsNonRoot:             utils.Pointer[bool](false),
			ReadOnlyRootFilesystem:   utils.Pointer[bool](false),
			AllowPrivilegeEscalation: utils.Pointer[bool](true),
			Privileged:               utils.Pointer[bool](true),
		}
	}

	// Set node toleration to the specific arch
	if !viper.GetBool("k8s-supports-cross-platform-no-affinity") {
		podConfig.Spec.Affinity = &v1.Affinity{
			NodeAffinity: &v1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
					NodeSelectorTerms: []v1.NodeSelectorTerm{
						{
							MatchExpressions: []v1.NodeSelectorRequirement{
								{
									Key:      "peridot.rockylinux.org/workflow-tolerates-arch",
									Operator: "In",
									Values:   []string{nodePoolArch},
								},
							},
						},
					},
				},
			},
		}
	}

	metadata := &peridotpb.ProvisionWorkerMetadata{
		Name:    name,
		Purpose: req.Purpose,
		TaskId:  req.TaskId,
	}
	metadataAny, err := anypb.New(metadata)
	if err != nil {
		return "", fmt.Errorf("could not create metadata any: %v", err)
	}
	err = c.db.SetTaskMetadata(task.ID.String(), metadataAny)
	if err != nil {
		return "", fmt.Errorf("could not set task metadata: %v", err)
	}

	pod, err := podInterface.Create(ctx, podConfig, metav1.CreateOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			return name, nil
		}
		return "", fmt.Errorf("could not create pod: %v", err)
	}

	runningCount := 0
	for {
		if runningCount >= 3 {
			break
		}
		pod, err := podInterface.Get(ctx, pod.Name, metav1.GetOptions{})
		if err != nil {
			return "", fmt.Errorf("could not get pod: %v", err)
		}

		if pod.Status.Phase == v1.PodFailed {
			return "", temporal.NewNonRetryableApplicationError("pod failed", "Failed pod", nil, pod.Name)
		}
		if pod.Status.Phase == v1.PodRunning {
			runningCount += 1
			break
		}

		time.Sleep(time.Second)
	}

	pod, err = podInterface.Get(ctx, pod.Name, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("could not get pod: %v", err)
	}
	if pod.Status.Phase == v1.PodFailed {
		return "", fmt.Errorf("pod failed")
	}

	_ = c.logToMon(
		[]string{fmt.Sprintf("Created worker %s", pod.Name)},
		task.ID.String(),
		req.ParentTaskId.String,
	)

	return pod.Name, nil
}
