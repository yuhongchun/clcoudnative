package k8sresource

var (
	ImageMergePatchJsonTemp = `{"spec":{"template":{"spec":{"containers":[{"name":"${projectName}","image":"repourl:version"}]}}}}`
)

type ImageMergePatch struct {
	Spec struct {
		Template struct {
			Spec struct {
				Containers []Container `json:"containers"`
			} `json:"spec"`
		} `json:"template"`
	} `json:"spec"`
}

type Container struct {
	Name  string `json:"name"`
	Image string `json:"image"`
}
type ImageJsonPatch struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value string `json:"value"`
}

type JsonPatch struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value"`
}

type LocalObjectMeta struct {
	Name   string            `json:"name"`
	Labels map[string]string `json:"labels"`
}

type LocalLabelSelector struct {
	MatchLabels map[string]string `json:"matchLabels"`
}
type LocalRollingUpdateDeployment struct {
	MaxUnavailable string `json:"maxUnavailable"`
	MaxSurge       string `json:"maxSurge"`
}
type LocalDeploymentStrategy struct {
	Type          string                        `json:"type"`
	RollingUpdate *LocalRollingUpdateDeployment `json:"rollingUpdate"`
}
type LocalEnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
type LocalResourceRequirements struct {
	Limits   map[string]string `json:"limits"`
	Requests map[string]string `json:"requests"`
}
type LocalSecurityContext struct {
	Privileged *bool `json:"privileged"`
}
type LocalVolumeMount struct {
	Name      string `json:"name"`
	MountPath string `json:"mountPath"`
	SubPath   string `json:"subPath"`
}
type LocalContainer struct {
	Name                     string                    `json:"name"`
	Image                    string                    `json:"image"`
	Command                  []string                  `json:"command"`
	Args                     []string                  `json:"args"`
	WorkingDir               string                    `json:"workingDir"`
	Env                      []LocalEnvVar             `json:"env"`
	ImagePullPolicy          string                    `json:"imagePullPolicy"`
	Resources                LocalResourceRequirements `json:"resources"`
	SecurityContext          *LocalSecurityContext     `json:"securityContext"`
	TerminationMessagePath   string                    `json:"terminationMessagePath"`
	TerminationMessagePolicy string                    `json:"terminationMessagePolicy"`
	VolumeMounts             []LocalVolumeMount        `json:"volumeMounts"`
}
type LocalObjectReference struct {
	Name string `json:"name"`
}
type LocalPodSecurityContext struct {
}
type LocalVolume struct {
	Name string `json:"name"`
	//	VolumeSource `json:""`
}
type LocalPodSpec struct {
	Containers                    []LocalContainer         `json:"containers"`
	DNSPolicy                     string                   `json:"dnsPolicy"`
	ImagePullSecrets              []LocalObjectReference   `json:"imagePullSecrets"`
	RestartPolicy                 string                   `json:"restartPolicy"`
	SchedulerName                 string                   `json:"schedulerName"`
	SecurityContext               *LocalPodSecurityContext `json:"securityContext"`
	TerminationGracePeriodSeconds *int64                   `json:"terminationGracePeriodSeconds"`
	Volumes                       []LocalVolume            `json:"volumes"`
}
type LocalPodTemplateSpec struct {
	Metadata LocalObjectMeta `json:"metadata"`
	Spec     LocalPodSpec    `json:"spec"`
}
type LoaclDeploymentSpec struct {
	ProgressDeadlineSeconds *int32                  `json:"progressDeadlineSeconds"`
	Replicas                *int32                  `json:"replicas"`
	RevisionHistoryLimit    *int32                  `json:"revisionHistoryLimit"`
	Selector                LocalLabelSelector      `json:"selector"`
	Strategy                LocalDeploymentStrategy `json:"strategy"`
	Template                LocalPodTemplateSpec    `json:"template"`
}

type LocalDeployment struct {
	ApiVersion string              `json:"api_version"`
	Kind       string              `json:"kind"`
	Metadata   LocalObjectMeta     `json:"metadata"`
	Spec       LoaclDeploymentSpec `json:"spec"`
}

func test() {
	//d := appsv1.Deployment{}
}
