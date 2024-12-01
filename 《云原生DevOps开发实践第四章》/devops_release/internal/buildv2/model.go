package buildv2

type K8s_cluster_info struct {
	Name      string `yaml:"cluster_name"`
	Url       string
	Token     string
	Ca        string
	TLK_ID    string
	Namespace []string `yaml:"namespaces"`
}

type OpsInfo struct {
	ProjectName   string         `json:"project_name"`
	Type          string         `json:"type"`
	OpsProResults []OpsProResult `json:"ops_pro_results"`
	Channel       string         `json:"channel"`
	Status        string         `json:"status"`
}
type OpsProResult struct {
	Namespace      string   `json:"namespace"`
	ClusterName    string   `json:"cluster_name"`
	Status         string   `json:"status"`
	Message        string   `json:"message"`
	DeploymentErrs []string `json:"deployment_errs"`
	ServiceErrs    []string `json:"service_errs"`
	ConfigMapErrs  []string `json:"configmap_errs"`
}
