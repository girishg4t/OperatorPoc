This poc is show, if any thing is changed in configmap then the corresponding pod should restart

it identify the pod by Labels and delete that so the replicaSet will again create it

references: https://github.com/NautiluX/presentation-example-operator/blob/master/pkg/controller/presentation/presentation_controller.go
https://www.magalix.com/blog/creating-custom-kubernetes-operators