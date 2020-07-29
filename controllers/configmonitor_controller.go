/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	cachev1alpha1 "github.com/example-inc/configmap-operator/api/v1alpha1"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// ConfigMonitorReconciler reconciles a ConfigMonitor object
type ConfigMonitorReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=cache.example.com,resources=configmonitors,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cache.example.com,resources=configmonitors/status,verbs=get;update;patch

func (r *ConfigMonitorReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("configmonitor", req.NamespacedName)

	if req.Namespace != "default" {
		return ctrl.Result{}, nil
	}

	configMap := &corev1.ConfigMap{}
	errCm := r.Get(ctx, req.NamespacedName, configMap)

	if errCm == nil {
		var configMapList cachev1alpha1.ConfigMonitorList
		err := r.List(context.Background(), &configMapList,
			client.InNamespace(req.Namespace))
		if err == nil {
			for _, app := range configMapList.Items {
				var podSelector = app.Spec.PodSelector
				var podList corev1.PodList
				errPod := r.List(context.Background(), &podList,
					client.InNamespace(req.Namespace),
					client.MatchingLabels(map[string]string{"app": podSelector.App}))
				if errPod == nil {
					for _, pod := range podList.Items {
						err = r.Delete(ctx, &pod)
					}
				}
			}
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("Memcached resource not found. Ignoring since object must be deleted")
			if errors.IsNotFound(errCm) {
				log.Info("Memcached resource not found. Ignoring since object must be deleted")
			}
		}
	}
	return ctrl.Result{}, nil
}

func (r *ConfigMonitorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.ConfigMap{}).
		Owns(&cachev1alpha1.ConfigMonitor{}).
		Complete(r)
}

func ignoreDeletionPredicate() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			// Ignore updates to CR status in which case metadata.Generation does not change
			return e.MetaOld.GetGeneration() != e.MetaNew.GetGeneration()
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			// Evaluates to false if the object has been confirmed deleted.
			return false
		},
	}
}
