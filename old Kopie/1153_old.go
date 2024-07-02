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

package resources

import (
	appsv1 "k8s.io/api/apps/v1"
	hpav2beta1 "k8s.io/api/autoscaling/v2beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/kmeta"
)

// MakeHorizontalPodAutoscaler makes an HPA for the given arguments.
func MakeHorizontalPodAutoscaler(deployment *appsv1.Deployment, args AutoscalingArgs) *hpav2beta1.HorizontalPodAutoscaler {
	var one int32 = 1
	q := resource.MustParse(args.AvgMemoryUsage)
	return &hpav2beta1.HorizontalPodAutoscaler{
		ObjectMeta: metav1.ObjectMeta{
			Name:            deployment.Name + "-hpa",
			Namespace:       deployment.Namespace,
			OwnerReferences: []metav1.OwnerReference{*kmeta.NewControllerRef(args.BrokerCell)},
			Labels:          Labels(args.BrokerCell.Name, args.ComponentName),
		},
		Spec: hpav2beta1.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: hpav2beta1.CrossVersionObjectReference{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       deployment.Name,
			},
			MaxReplicas: args.MaxReplicas,
			MinReplicas: &one,
			Metrics: []hpav2beta1.MetricSpec{
				{
					Type: hpav2beta1.ResourceMetricSourceType,
					Resource: &hpav2beta1.ResourceMetricSource{
						Name:                     corev1.ResourceCPU,
						TargetAverageUtilization: &args.AvgCPUUtilization,
					},
				},
				{
					Type: hpav2beta1.ResourceMetricSourceType,
					Resource: &hpav2beta1.ResourceMetricSource{
						Name:               corev1.ResourceMemory,
						TargetAverageValue: &q,
					},
				},
			},
		},
	}
}
