/*
Copyright 2025.

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

package controller

import (
	"context"
	"time"

	opsv1 "github.com/akoe32/config-reloader/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ConfigReloaderReconciler reconciles a ConfigReloader object
type ConfigReloaderReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=ops.nendeskombet.com,resources=configreloaders,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=ops.nendeskombet.com,resources=configreloaders/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=ops.nendeskombet.com,resources=configreloaders/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ConfigReloader object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.0/pkg/reconcile
func (r *ConfigReloaderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	log := ctrl.LoggerFrom(ctx)
	reloader := &opsv1.ConfigReloader{}

	if err := r.Get(ctx, req.NamespacedName, reloader); err != nil {
		if errors.IsNotFound(err) {
			log.Info("ConfigReloader resource not found")
			return ctrl.Result{}, nil
		}
	}

	// Check changing on configmap or secret
	if reloader.Spec.WorkloadType == "ConfigMap" {
		configMap := &corev1.ConfigMap{}
		if err := r.Get(ctx, types.NamespacedName{Name: reloader.Spec.ConfigmapName, Namespace: reloader.Namespace}, configMap); err != nil {
			if errors.IsNotFound(err) {
				log.Info("ConfigMap not found, skipping reload...")
				return ctrl.Result{}, nil
			}
			return ctrl.Result{}, nil

		}
	} else if reloader.Spec.ResourceType == "Secret" {
		secret := &corev1.Secret{}
		if err := r.Get(ctx, types.NamespacedName{Name: reloader.Spec.SecretName, Namespace: reloader.Namespace}, secret); err != nil {
			if errors.IsNotFound(err) {
				log.Info("Secret not found, skipping reload...")
				return ctrl.Result{}, nil
			}
			return ctrl.Result{}, nil
		}
	} else {
		log.Info("Unknown resource type, pskipping reload...")
	}

	switch reloader.Spec.WorkloadType {
	case "Deployment":
		err := r.restartDeployment(ctx, reloader)
		if err != nil {
			return ctrl.Result{}, nil
		}
	case "Statefulset":
		err := r.restartStatefulset(ctx, reloader)
		if err != nil {
			return ctrl.Result{}, nil
		}
	case "Daemonset":
		err := r.Daemonset(ctx, reloader)
		if err != nil {
			return ctrl.Result{}, nil
		}
	default:
		log.Info("Unknown workload type, skipping reload...")
		return ctrl.Result{}, nil
	}

	reloader.Status.LastReloadTime = metav1.Now()
	if err := r.Status().Update(ctx, reloader); err != nil {
		return ctrl.Result{}, nil
	}

	return ctrl.Result{RequeueAfter: time.Minute * 5}, nil
}

func (r *ConfigReloaderReconciler) Daemonset(ctx context.Context, reloader *opsv1.ConfigReloader) error {
	daemonset := &appsv1.DaemonSet{}
	err := r.Get(ctx, types.NamespacedName{Name: reloader.Spec.WorkloadName, Namespace: reloader.Namespace}, daemonset)
	if err != nil {
		return err
	}
	daemonset.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)
	return r.Update(ctx, daemonset)
}

func (r *ConfigReloaderReconciler) restartStatefulset(ctx context.Context, reloader *opsv1.ConfigReloader) error {
	statefulset := &appsv1.StatefulSet{}
	err := r.Get(ctx, types.NamespacedName{Name: reloader.Spec.WorkloadName, Namespace: reloader.Namespace}, statefulset)
	if err != nil {
		return err
	}
	statefulset.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)
	return r.Update(ctx, statefulset)
}

func (r *ConfigReloaderReconciler) restartDeployment(ctx context.Context, reloader *opsv1.ConfigReloader) error {
	deployment := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: reloader.Spec.WorkloadName, Namespace: reloader.Namespace}, deployment)
	if err != nil {
		return err
	}
	deployment.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)
	return r.Update(ctx, deployment)
}

// SetupWithManager sets up the controller with the Manager.
func (r *ConfigReloaderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&opsv1.ConfigReloader{}).
		Named("configreloader").
		Complete(r)
}
