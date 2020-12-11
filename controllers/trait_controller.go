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
	"fmt"
	"github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	cpv1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/oam-kubernetes-runtime/pkg/oam/discoverymapper"
	"github.com/crossplane/oam-kubernetes-runtime/pkg/oam/util"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kruisetraitv1 "kruise_trait/api/v1alpha1"
	ingresstraitv1 "kruise_trait/api/v1alpha2"
)

const (
	errNotTrait       = "object is not a trait"
	errNotWorkload    = "workload is not containerized workload"
	errRenderWorkload = "cannot render workload"
	errRenderTrait    = "cannot render trait"
	errApplyConfigMap = "cannot apply configmap"
)

// TraitReconciler reconciles a KruiseTrait object
type TraitReconciler struct {
	client.Client
	discovery.DiscoveryClient
	Log    logr.Logger
	record event.Recorder
	Scheme *runtime.Scheme
	dm     discoverymapper.DiscoveryMapper
}

//func Setup(mgr ctrl.Manager) error {
//	dm, err := discoverymapper.New(mgr.GetConfig())
//	if err != nil {
//		return err
//	}
//	reconciler := TraitReconciler{
//		Client:          mgr.GetClient(),
//		DiscoveryClient: *discovery.NewDiscoveryClientForConfigOrDie(mgr.GetConfig()),
//		log:             ctrl.Log.WithName("KruiseTrait"),
//		record:          event.NewAPIRecorder(mgr.GetEventRecorderFor("KruiseTrait")),
//		Scheme:          mgr.GetScheme(),
//		dm:              dm,
//	}
//	return reconciler.SetupWithManager(mgr)
//}

// +kubebuilder:rbac:groups=core.oam.dev,resources=kruisetraits,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.oam.dev.v1alpha1,resources=kruisetraits/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps.kruise.io,resources=clonesets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.kruise.io,resources=clonesets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps.kruise.io,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.kruise.io,resources=statefulsets/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps.kruise.io,resources=uniteddeployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.kruise.io,resources=uniteddeployments/status,verbs=get;update;patch
func (r *TraitReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	mLog := r.Log.WithValues("trait", req.NamespacedName)

	mLog.Info("reconcile to configmap")
	var ingressTrait ingresstraitv1.IngressTrait
	if err := r.Get(ctx, req.NamespacedName, &ingressTrait); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	mLog.Info("Get the ingress trait", "WorkloadReference", ingressTrait.Spec.WorkloadReference)

	var trait kruisetraitv1.KruiseTrait
	if err := r.Get(ctx, req.NamespacedName, &trait); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Fetch applicationConfiguration
	eventObj, err := util.LocateParentAppConfig(ctx, r.Client, &trait)
	if eventObj == nil {
		mLog.Error(err, "Failed to find parent resource", trait.Name)
		eventObj = &trait
	}

	// Fetch the workload instance this trait is refer to
	workload, err := util.FetchWorkload(ctx, r, mLog, &trait)
	if err != nil {
		r.record.Event(eventObj, event.Warning(util.ErrLocateWorkload, err))
		return util.ReconcileWaitResult, util.PatchCondition(
			ctx, r, &trait, v1alpha1.ReconcileError(errors.Wrap(err, util.ErrLocateWorkload)))
	}
	resources, err := DetermineWorkloadType(ctx, mLog, workload, r)
	if err != nil {
		r.record.Event(eventObj, event.Warning(util.ErrLocateWorkload, err))
		return util.ReconcileWaitResult, util.PatchCondition(
			ctx, r, &trait, cpv1alpha1.ReconcileError(errors.Wrap(err, util.ErrLocateAppConfig)))
	}
	if len(resources) == 0 {
		resources = append(resources, workload)
	}
	configMapApplyOpts := []client.PatchOption{client.ForceOwnership, client.FieldOwner(workload.GetUID())}
	configMaps, err := r.createConfigmap(ctx, trait, resources)
	if err != nil {
		mLog.Error(err, "Failed to render configmaps")
		r.record.Event(eventObj, event.Warning(errRenderWorkload, err))
		return util.ReconcileWaitResult,
			util.PatchCondition(ctx, r, &trait, cpv1alpha1.ReconcileError(errors.Wrap(err, errNotWorkload)))
	}
	for _, cm := range configMaps {
		if err := r.Patch(ctx, cm, client.Apply, configMapApplyOpts...); err != nil {
			mLog.Error(err, "Failed to apply a configmap")
			r.record.Event(eventObj, event.Warning(errApplyConfigMap, err))
			return util.ReconcileWaitResult,
				util.PatchCondition(ctx, r, &trait, cpv1alpha1.ReconcileError(errors.Wrap(err, errApplyConfigMap)))
		}
		r.record.Event(eventObj, event.Normal("ConfigMap created",
			fmt.Sprintf("Workload `%s` successfully server side patched a configmap `%s`",
				workload.GetName(), cm.Name)))
	}
	// record the configmaps
	trait.Status.PatchConfigMap = configMaps

	trait.Status.Resources = nil
	for _, cm := range configMaps {
		trait.Status.Resources = append(trait.Status.Resources, cpv1alpha1.TypedReference{
			APIVersion: cm.GetObjectKind().GroupVersionKind().GroupVersion().String(),
			Kind:       cm.GetObjectKind().GroupVersionKind().Kind,
			Name:       cm.GetName(),
			UID:        cm.GetUID(),
		})
	}

	if err := r.Status().Update(ctx, &trait); err != nil {
		return util.ReconcileWaitResult, err
	}

	return ctrl.Result{}, util.PatchCondition(ctx, r, &trait, cpv1alpha1.ReconcileSuccess())
}

func (r *TraitReconciler) createConfigmap(ctx context.Context, tr kruisetraitv1.KruiseTrait,
	resources []*unstructured.Unstructured) ([]*corev1.ConfigMap, error) {
	var newConfigMap []*corev1.ConfigMap

	for _, res := range resources {
		if res.GetAPIVersion() == appsv1.SchemeGroupVersion.String() {
			configMap, err := r.renderConfigMaps(tr, res)
			if err != nil {
				r.Log.Error(err, "Failed to render configmap")
				return nil, util.PatchCondition(ctx, r, &tr, cpv1alpha1.ReconcileError(errors.Wrap(err, errRenderTrait)))
			}
			newConfigMap = append(newConfigMap, configMap)
		}
	}
	return newConfigMap, nil
}

func (r *TraitReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kruisetraitv1.KruiseTrait{}).
		Complete(r)
}
