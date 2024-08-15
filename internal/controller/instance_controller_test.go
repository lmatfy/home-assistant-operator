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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	homeassistantv1alpha1 "github.com/lmatfy/home-assistant-operator/api/v1alpha1"
)

var _ = Describe("Instance Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()

		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default", // TODO(user):Modify as needed
		}
		instance := &homeassistantv1alpha1.Instance{}

		BeforeEach(func() {
			By("creating the custom resource for the Kind Instance")
			err := k8sClient.Get(ctx, typeNamespacedName, instance)
			if err != nil && errors.IsNotFound(err) {
				resource := &homeassistantv1alpha1.Instance{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: homeassistantv1alpha1.InstanceSpec{
						HostNetwork: true,
						Env: []corev1.EnvVar{
							{
								Name:  "TZ",
								Value: "UTC",
							},
						},
						Ingress: homeassistantv1alpha1.Ingress{
							Enabled:    true,
							Host:       "home-assistant.example.com",
							SecretName: "wildcard-example-com-tls",
						},
						Persistence: homeassistantv1alpha1.Persistence{
							Size: "2Gi",
						},
					},
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			// TODO(user): Cleanup logic after each test, like removing the resource instance.
			resource := &homeassistantv1alpha1.Instance{}
			err := k8sClient.Get(ctx, typeNamespacedName, resource)
			Expect(err).NotTo(HaveOccurred())

			By("Cleanup the specific resource instance Instance")
			Expect(k8sClient.Delete(ctx, resource)).To(Succeed())
		})
		It("should successfully reconcile the resource", func() {
			By("Reconciling the created resource")
			controllerReconciler := &InstanceReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			By("Checking if PersistentVolumeClaim was successfully created in the reconciliation")
			Eventually(func() error {
				claim := &corev1.PersistentVolumeClaim{}
				gg := typeNamespacedName
				gg.Name = gg.Name + "-config"
				return k8sClient.Get(ctx, gg, claim)
			}, time.Minute, time.Second).Should(Succeed())

			By("Checking if Pod was successfully created in the reconciliation")
			Eventually(func() error {
				pod := &corev1.Pod{}
				return k8sClient.Get(ctx, typeNamespacedName, pod)
			}, time.Minute, time.Second).Should(Succeed())

			By("Checking if Service was successfully created in the reconciliation")
			Eventually(func() error {
				service := &corev1.Service{}
				return k8sClient.Get(ctx, typeNamespacedName, service)
			}, time.Minute, time.Second).Should(Succeed())

			By("Checking if Ingress was successfully created in the reconciliation")
			Eventually(func() error {
				ingress := &networkingv1.Ingress{}
				return k8sClient.Get(ctx, typeNamespacedName, ingress)
			}, time.Minute, time.Second).Should(Succeed())
		})
	})
})
