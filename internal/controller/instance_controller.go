/*
Copyright 2024.

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

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"

	homeassistantv1alpha1 "github.com/lmatfy/home-assistant-operator/api/v1alpha1"
)

// InstanceReconciler reconciles a Instance object
type InstanceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=home-assistant.lmatfy.io,resources=instances,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=home-assistant.lmatfy.io,resources=instances/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=home-assistant.lmatfy.io,resources=instances/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=pod,verbs=get;watch;create;update;delete
//+kubebuilder:rbac:groups=apps,resources=ingress,verbs=get;watch;create;update;delete
//+kubebuilder:rbac:groups=apps,resources=service,verbs=get;watch;create;update;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Instance object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.3/pkg/reconcile
func (r *InstanceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)

	instance := &homeassistantv1alpha1.Instance{}
	if err := r.Get(ctx, req.NamespacedName, instance); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, r.cleanupResources(ctx, req)
		}
		l.Error(err, "Failed to get Instance.")
		return ctrl.Result{}, nil
	}

	r.reconcilePersistentVolumeClaim(ctx, req, instance)
	r.reconcilePod(ctx, req, instance)
	r.reconcileService(ctx, req, instance)
	r.reconcileIngress(ctx, req, instance)

	return ctrl.Result{}, nil
}

func (r *InstanceReconciler) cleanupResources(ctx context.Context, req ctrl.Request) error {
	l := log.FromContext(ctx)

	l.Info("Cleanup resources.")

	ingress := &networkingv1.Ingress{}
	if err := r.Get(ctx, req.NamespacedName, ingress); err != nil {
		if !errors.IsNotFound(err) {
			l.Error(err, "Failed to delete ingress resource: fetch failed.")
		}
	} else {
		if err := r.Delete(ctx, ingress); err != nil {
			l.Error(err, "Failed to delete ingress resource: delete failed.")
			return err
		}
		l.Info("Ingress deleted.")
	}

	service := &corev1.Service{}
	if err := r.Get(ctx, req.NamespacedName, service); err != nil {
		if !errors.IsNotFound(err) {
			l.Error(err, "Failed to delete service resource: fetch failed.")
		}
	} else {
		if err := r.Delete(ctx, service); err != nil {
			l.Error(err, "Failed to delete service resource: delete failed.")
			return err
		}
		l.Info("Service deleted.")
	}

	pod := &corev1.Pod{}
	if err := r.Get(ctx, req.NamespacedName, pod); err != nil {
		if !errors.IsNotFound(err) {
			l.Error(err, "Failed to delete pod resource: fetch failed.")
		}
	} else {
		if err := r.Delete(ctx, pod); err != nil {
			l.Error(err, "Failed to delete pod resource: delete failed.")
			return err
		}
		l.Info("Pod deleted.")
	}

	claim := &corev1.PersistentVolumeClaim{}
	name := req.NamespacedName
	name.Name = name.Name + "-config"
	if err := r.Get(ctx, name, claim); err != nil {
		if !errors.IsNotFound(err) {
			l.Error(err, "Failed to delete persistentvolumeclaim resource: fetch failed.")
		}
	} else {
		if err := r.Delete(ctx, claim); err != nil {
			l.Error(err, "Failed to delete persistentvolumeclaim resource: delete failed.")
			return err
		}
		l.Info("PersistentVolumeClaim deleted.")
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *InstanceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&homeassistantv1alpha1.Instance{}).
		Owns(&corev1.PersistentVolumeClaim{}).
		Owns(&corev1.Pod{}).
		Owns(&corev1.Service{}).
		Owns(&networkingv1.Ingress{}).
		Complete(r)
}

func (r *InstanceReconciler) reconcileIngress(ctx context.Context, req ctrl.Request, instance *homeassistantv1alpha1.Instance) {
	ingress := &networkingv1.Ingress{}
	isIngressAvailable := true
	if err := r.Get(ctx, req.NamespacedName, ingress); err != nil {
		// TODO: add log
		isIngressAvailable = false
	}

	if !instance.Spec.Ingress.Enabled && !isIngressAvailable {
		return
	}

	if !instance.Spec.Ingress.Enabled && isIngressAvailable {
		if err := r.Delete(ctx, ingress); err != nil {
			// TODO: add log
			return
		}
	}

	ingress.Name = req.Name
	ingress.Namespace = instance.Namespace
	pathType := networkingv1.PathTypePrefix
	ingress.Spec.Rules = []networkingv1.IngressRule{
		{
			Host: instance.Spec.Ingress.Host,
			IngressRuleValue: networkingv1.IngressRuleValue{
				HTTP: &networkingv1.HTTPIngressRuleValue{
					Paths: []networkingv1.HTTPIngressPath{
						{
							Path:     "/",
							PathType: &pathType,
							Backend: networkingv1.IngressBackend{
								Service: &networkingv1.IngressServiceBackend{
									Name: instance.Name,
									Port: networkingv1.ServiceBackendPort{
										Number: 8123,
									},
								},
							},
						},
					},
				},
			},
		},
	}
	ingress.Spec.TLS = []networkingv1.IngressTLS{
		{
			Hosts: []string{
				instance.Spec.Ingress.Host,
			},
			SecretName: instance.Spec.Ingress.SecretName,
		},
	}
	applyLabels(&ingress.ObjectMeta, instance)
	applyAnnotations(&ingress.ObjectMeta, instance)

	if !isIngressAvailable {
		if err := r.Create(ctx, ingress); err != nil {
			// TODO: add log
			return
		}
	} else {
		if err := r.Update(ctx, ingress); err != nil {
			// TODO: add log
			return
		}
	}
}

func (r *InstanceReconciler) reconcilePersistentVolumeClaim(ctx context.Context, req ctrl.Request, instance *homeassistantv1alpha1.Instance) {
	claim := &corev1.PersistentVolumeClaim{}
	if err := r.Get(ctx, req.NamespacedName, claim); err == nil {
		// TODO: add log
		return
	}

	storageSize, err := resource.ParseQuantity(instance.Spec.Persistence.Size)
	if err != nil {
		// TODO: add log and return error
		return
	}

	if instance.Spec.Persistence.Size == "" {
		instance.Spec.Persistence.Size = "1Gi"
	}

	claim.Name = instance.Name + "-config" // TODO: use name function
	claim.Namespace = instance.Namespace
	claim.Spec.AccessModes = []corev1.PersistentVolumeAccessMode{
		corev1.ReadWriteOnce,
	}
	claim.Spec.StorageClassName = instance.Spec.Persistence.StorageClassName
	claim.Spec.Resources = corev1.VolumeResourceRequirements{
		Requests: corev1.ResourceList{
			"storage": storageSize,
		},
	}

	if err := r.Create(ctx, claim); err != nil {
		// TODO: add log
		return
	}
}

func (r *InstanceReconciler) reconcilePod(ctx context.Context, req ctrl.Request, instance *homeassistantv1alpha1.Instance) {
	pod := &corev1.Pod{}
	isPodAvailable := true
	if err := r.Get(ctx, req.NamespacedName, pod); err != nil {
		// TODO: add log
		isPodAvailable = false
	}

	pod.Name = instance.Name
	pod.Namespace = instance.Namespace
	isPriviledged := true
	var tolerationsSec int64 = 300
	pod.Spec.Affinity = &instance.Spec.Affinity
	pod.Spec.HostNetwork = instance.Spec.HostNetwork
	pod.Spec.Tolerations = []corev1.Toleration{
		{
			Effect:            corev1.TaintEffectNoExecute,
			Key:               "node.kubernetes.io/not-ready",
			Operator:          corev1.TolerationOpExists,
			TolerationSeconds: &tolerationsSec,
		},
		{
			Effect:            corev1.TaintEffectNoExecute,
			Key:               "node.kubernetes.io/unreachable",
			Operator:          corev1.TolerationOpExists,
			TolerationSeconds: &tolerationsSec,
		},
	}
	pod.Spec.Volumes = []corev1.Volume{
		{
			Name: "config",
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: "/run/dbus",
				},
			},
		},
		{
			Name: "dbus",
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: instance.Name + "-config", // TODO: use name function
				},
			},
		},
	}
	pod.Spec.Containers = []corev1.Container{
		{
			Name:            instance.Name,
			Env:             instance.Spec.Env,
			Image:           "homeassistant/home-assistant:" + instance.Spec.Version,
			ImagePullPolicy: corev1.PullIfNotPresent, // TODO: make configurable
			LivenessProbe: &corev1.Probe{
				FailureThreshold: 3,
				PeriodSeconds:    10,
				SuccessThreshold: 1,
				TimeoutSeconds:   1,
				ProbeHandler: corev1.ProbeHandler{
					TCPSocket: &corev1.TCPSocketAction{
						Port: intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: 8123,
						},
					},
				},
			},
			ReadinessProbe: &corev1.Probe{
				FailureThreshold: 3,
				PeriodSeconds:    10,
				SuccessThreshold: 1,
				TimeoutSeconds:   1,
				ProbeHandler: corev1.ProbeHandler{
					TCPSocket: &corev1.TCPSocketAction{
						Port: intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: 8123,
						},
					},
				},
			},
			StartupProbe: &corev1.Probe{
				FailureThreshold: 30,
				PeriodSeconds:    5,
				SuccessThreshold: 1,
				TimeoutSeconds:   1,
				ProbeHandler: corev1.ProbeHandler{
					TCPSocket: &corev1.TCPSocketAction{
						Port: intstr.IntOrString{
							Type:   intstr.Int,
							IntVal: 8123,
						},
					},
				},
			},
			Ports: []corev1.ContainerPort{
				{
					Name:          "http",
					ContainerPort: 8123,
					HostPort:      8123,
					Protocol:      corev1.ProtocolTCP,
				},
			},
			SecurityContext: &corev1.SecurityContext{
				Privileged: &isPriviledged,
			},
			VolumeMounts: []corev1.VolumeMount{
				{
					MountPath: "/config",
					Name:      "config",
				},
				{
					MountPath: "/run/dbus",
					Name:      "dbus",
				},
			},
		},
	}
	applyLabels(&pod.ObjectMeta, instance)
	applyAnnotations(&pod.ObjectMeta, instance)

	if !isPodAvailable {
		if err := r.Create(ctx, pod); err != nil {
			// TODO: add log
			return
		}
	} else {
		if err := r.Update(ctx, pod); err != nil {
			// TODO: add log
			return
		}
	}
}

func (r *InstanceReconciler) reconcileService(ctx context.Context, req ctrl.Request, instance *homeassistantv1alpha1.Instance) {
	service := &corev1.Service{}
	isServiceAvailable := true
	if err := r.Get(ctx, req.NamespacedName, service); err != nil {
		// TODO: add log
		isServiceAvailable = false
	}

	service.Name = instance.Name
	service.Namespace = instance.Namespace
	service.Spec.Type = corev1.ServiceTypeClusterIP
	service.Spec.Ports = []corev1.ServicePort{
		{
			Name:     "http",
			Port:     8123,
			Protocol: "TCP",
			TargetPort: intstr.IntOrString{
				Type:   intstr.String,
				StrVal: "http",
			},
		},
	}
	service.Spec.Selector = map[string]string{
		"app.kubernetes.io/instance": instance.Name,
		"app.kubernetes.io/name":     instance.Name,
	}
	applyLabels(&service.ObjectMeta, instance)
	applyAnnotations(&service.ObjectMeta, instance)

	if !isServiceAvailable {
		if err := r.Create(ctx, service); err != nil {
			// TODO: add log
			return
		}
	} else {
		if err := r.Update(ctx, service); err != nil {
			// TODO: add log
			return
		}
	}
}

func applyAnnotations(obj *metav1.ObjectMeta, instance *homeassistantv1alpha1.Instance) {
	if obj.Annotations == nil {
		obj.Annotations = make(map[string]string)
	}

	for name, value := range instance.Spec.Annotations {
		obj.Annotations[name] = value
	}
}

func applyLabels(obj *metav1.ObjectMeta, instance *homeassistantv1alpha1.Instance) {
	if obj.Labels == nil {
		obj.Labels = make(map[string]string)
	}
	for name, value := range instance.Spec.Labels {
		obj.Labels[name] = value
	}
	obj.Labels["app.kubernetes.io/instance"] = instance.Name
	obj.Labels["app.kubernetes.io/managed-by"] = "home-assistant-operator" // TODO: get operator name
	obj.Labels["app.kubernetes.io/name"] = instance.Name
	obj.Labels["app.kubernetes.io/version"] = instance.Spec.Version
}
