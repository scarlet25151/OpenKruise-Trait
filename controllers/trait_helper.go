package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/crossplane/oam-kubernetes-runtime/apis/core/v1alpha2"
	"github.com/crossplane/oam-kubernetes-runtime/pkg/oam"
	"github.com/crossplane/oam-kubernetes-runtime/pkg/oam/util"
	"github.com/go-logr/logr"
	kruise "github.com/openkruise/kruise-api/apps/v1alpha1"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"kruise_trait/api/v1alpha1"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
)

const (
	cloneSet            = "CloneSet"
	advancedStatefulSet = "StatefulSet"
	unitedDeployment    = "UnitedDeployment"

	updateStrategy = "updateStrategy"
	scaleStrategy  = "scaleStrategy"
)

var (
	workloadAPIVersion = v1alpha2.SchemeGroupVersion.String()
	appsAPIVersion     = v1alpha2.SchemeGroupVersion.String()

	configMapKind       = reflect.TypeOf(corev1.ConfigMap{}).Name()
	configMapAPIVersion = corev1.SchemeGroupVersion.String()
)

func DetermineWorkloadType(ctx context.Context, logger logr.Logger,
	workload *unstructured.Unstructured, r *TraitReconciler) ([]*unstructured.Unstructured, error) {
	apiVersion := workload.GetAPIVersion()
	switch apiVersion {
	case workloadAPIVersion:
		return util.FetchWorkloadChildResources(ctx, logger, r, r.dm, workload)
	case appsAPIVersion:
		logger.Info("workload is K8S native resources", "APIVersion", apiVersion)
		return []*unstructured.Unstructured{workload}, nil
	case "":
		return nil, errors.Errorf(fmt.Sprint("failed to get the workload apiVersion"))
	default:
		return nil, errors.Errorf(fmt.Sprint("This trait doesn't support the type", apiVersion))

	}
}

func (r *TraitReconciler) renderConfigMaps(tr v1alpha1.Trait, obj oam.Object) (*corev1.ConfigMap, error) {

	var (
		jsonProperties      = []byte(nil)
		configMapBinaryData = make(map[string][]byte, 0)
		err                 error
		newConfigMap        *corev1.ConfigMap
		bts                 []byte
	)

	switch obj.GetObjectKind().GroupVersionKind().Kind {
	case cloneSet:
		var cs kruise.CloneSet
		bts, _ = json.Marshal(obj)
		if err = json.Unmarshal(bts, &cs); err != nil {
			return nil, err
		}
		scaleProperties, _ := json.Marshal(tr.Spec.CloneSetScaleStrategy)
		jsonProperties, _ := json.Marshal(tr.Spec.CloneSetUpdateStrategy)
		configMapBinaryData[scaleStrategy] = scaleProperties
		configMapBinaryData[updateStrategy] = jsonProperties

		cm := &corev1.ConfigMap{

			TypeMeta: metav1.TypeMeta{
				Kind:       configMapKind,
				APIVersion: configMapAPIVersion,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "trait-data",
				Namespace: cs.GetNamespace(),
			},
			BinaryData: configMapBinaryData,
		}
		util.PassLabelAndAnnotation(&tr, cm)
		newConfigMap = cm

	case advancedStatefulSet:
		var ss kruise.StatefulSet
		bts, _ = json.Marshal(obj)
		if err = json.Unmarshal(bts, &ss); err != nil {
			return nil, err
		}
		jsonProperties, err = json.Marshal(tr.Spec.StatefulSetUpdateStrategy)
		configMapBinaryData[scaleStrategy] = jsonProperties
		cm := &corev1.ConfigMap{

			TypeMeta: metav1.TypeMeta{
				Kind:       configMapKind,
				APIVersion: configMapAPIVersion,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "trait-data",
				Namespace: ss.GetNamespace(),
			},
			BinaryData: configMapBinaryData,
		}
		util.PassLabelAndAnnotation(&tr, cm)
		newConfigMap = cm
	case unitedDeployment:
		var ud kruise.UnitedDeployment
		bts, _ = json.Marshal(obj)
		if err = json.Unmarshal(bts, &ud); err != nil {
			return nil, err
		}
		jsonProperties, err = json.Marshal(tr.Spec.UnitedDeploymentUpdateStrategy)
		configMapBinaryData[updateStrategy] = jsonProperties
		cm := &corev1.ConfigMap{

			TypeMeta: metav1.TypeMeta{
				Kind:       configMapKind,
				APIVersion: configMapAPIVersion,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "trait-data",
				Namespace: ud.GetNamespace(),
			},
			BinaryData: configMapBinaryData,
		}
		util.PassLabelAndAnnotation(&tr, cm)
		newConfigMap = cm
	}

	if err = ctrl.SetControllerReference(&tr, newConfigMap, r.Scheme); err != nil {
		return nil, err
	}

	return newConfigMap, nil
}
