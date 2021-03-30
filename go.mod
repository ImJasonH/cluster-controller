module github.com/imjasonh/cluster-controller

go 1.15

require (
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.16.0
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	k8s.io/api v0.20.2
	k8s.io/apiextensions-apiserver v0.20.2
	k8s.io/apimachinery v0.20.2
	k8s.io/client-go v0.20.2
	k8s.io/code-generator v0.20.2
	k8s.io/kube-openapi v0.0.0-20201113171705-d219536bb9fd
	knative.dev/hack v0.0.0-20210317214554-58edbdc42966
	knative.dev/pkg v0.0.0-20210318052054-dfeeb1817679
	sigs.k8s.io/controller-runtime v0.8.3
	sigs.k8s.io/controller-tools v0.5.0 // indirect
)
