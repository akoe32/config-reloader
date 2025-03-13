package controller

import (
	"context"
	"time"

	opsv1 "github.com/akoe32/config-reloader/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// ConfigReloaderReconciler reconciles a ConfigReloader object
type ConfigReloaderReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Reconcile detects changes and restarts workloads
func (r *ConfigReloaderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx)
	reloader := &opsv1.ConfigReloader{}
	if err := r.Get(ctx, req.NamespacedName, reloader); err != nil {
		if errors.IsNotFound(err) {
			log.Info("ConfigReloader resource not found")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	log.Info("Restarting workload", "WorkloadType", reloader.Spec.WorkloadType, "WorkloadName", reloader.Spec.WorkloadName)

	// Restart workload based on type
	switch reloader.Spec.WorkloadType {
	case "Deployment":
		if err := r.restartDeployment(ctx, reloader); err != nil {
			return ctrl.Result{}, err
		}
	case "StatefulSet":
		if err := r.restartStatefulSet(ctx, reloader); err != nil {
			return ctrl.Result{}, err
		}
	case "DaemonSet":
		if err := r.restartDaemonSet(ctx, reloader); err != nil {
			return ctrl.Result{}, err
		}
	default:
		log.Info("Unknown workload type, skipping reload...")
		return ctrl.Result{}, nil
	}

	reloader.Status.LastReloadTime = v1.Now()
	if err := r.Status().Update(ctx, reloader); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: time.Minute * 5}, nil
}

// Ensure annotations exist and restart workload
func ensureAnnotations(annotations map[string]string) map[string]string {
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)
	return annotations
}

// Restart functions for workloads
func (r *ConfigReloaderReconciler) restartDeployment(ctx context.Context, reloader *opsv1.ConfigReloader) error {
	deployment := &appsv1.Deployment{}
	if err := r.Get(ctx, types.NamespacedName{Name: reloader.Spec.WorkloadName, Namespace: reloader.Namespace}, deployment); err != nil {
		return err
	}
	deployment.Spec.Template.Annotations = ensureAnnotations(deployment.Spec.Template.Annotations)
	return r.Update(ctx, deployment)
}

func (r *ConfigReloaderReconciler) restartStatefulSet(ctx context.Context, reloader *opsv1.ConfigReloader) error {
	statefulSet := &appsv1.StatefulSet{}
	if err := r.Get(ctx, types.NamespacedName{Name: reloader.Spec.WorkloadName, Namespace: reloader.Namespace}, statefulSet); err != nil {
		return err
	}
	statefulSet.Spec.Template.Annotations = ensureAnnotations(statefulSet.Spec.Template.Annotations)
	return r.Update(ctx, statefulSet)
}

func (r *ConfigReloaderReconciler) restartDaemonSet(ctx context.Context, reloader *opsv1.ConfigReloader) error {
	daemonSet := &appsv1.DaemonSet{}
	if err := r.Get(ctx, types.NamespacedName{Name: reloader.Spec.WorkloadName, Namespace: reloader.Namespace}, daemonSet); err != nil {
		return err
	}
	daemonSet.Spec.Template.Annotations = ensureAnnotations(daemonSet.Spec.Template.Annotations)
	return r.Update(ctx, daemonSet)
}

func (r *ConfigReloaderReconciler) findAffectedConfigReloaders(ctx context.Context, obj client.Object) []reconcile.Request {
	var reloaderList opsv1.ConfigReloaderList
	if err := r.List(ctx, &reloaderList, client.InNamespace(obj.GetNamespace())); err != nil {
		return nil
	}

	var requests []reconcile.Request
	for _, reloader := range reloaderList.Items {
		if reloader.Spec.ConfigmapName == obj.GetName() {
			requests = append(requests, reconcile.Request{
				NamespacedName: types.NamespacedName{Name: reloader.Name, Namespace: reloader.Namespace},
			})
		}
	}
	return requests
}

func (r *ConfigReloaderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&opsv1.ConfigReloader{}).
		Named("configreloader").
		Watches(
			&corev1.ConfigMap{}, // Langsung menggunakan tipe ConfigMap
			handler.EnqueueRequestsFromMapFunc(r.findAffectedConfigReloaders),
		).
		Complete(r)
}
