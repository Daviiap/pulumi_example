package main

import (
	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apps/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		isMinikube := config.GetBool(ctx, "isMinikube")
		appName := "nginx"

		appLabels := pulumi.StringMap{
			"app": pulumi.String(appName),
		}

		deployment, err := appsv1.NewDeployment(ctx, appName, &appsv1.DeploymentArgs{
			Spec: appsv1.DeploymentSpecArgs{
				Selector: &metav1.LabelSelectorArgs{
					MatchLabels: appLabels,
				},
				Replicas: pulumi.Int(1),
				Template: &corev1.PodTemplateSpecArgs{
					Metadata: &metav1.ObjectMetaArgs{
						Labels: appLabels,
					},
					Spec: &corev1.PodSpecArgs{
						Containers: corev1.ContainerArray{
							corev1.ContainerArgs{
								Name:  pulumi.String("nginx"),
								Image: pulumi.String("nginx"),
							}},
					},
				},
			},
		})
		if err != nil {
			return err
		}

		feType := "LoadBalancer"
		if isMinikube {
			feType = "ClusterIP"
		}

		template := deployment.Spec.ApplyT(func(v appsv1.DeploymentSpec) *corev1.PodTemplateSpec {
			return &v.Template
		}).(corev1.PodTemplateSpecPtrOutput)

		meta := template.ApplyT(func(v *corev1.PodTemplateSpec) *metav1.ObjectMeta { return v.Metadata }).(metav1.ObjectMetaPtrOutput)

		frontend, _ := corev1.NewService(ctx, appName, &corev1.ServiceArgs{
			Metadata: meta,
			Spec: &corev1.ServiceSpecArgs{
				Type: pulumi.String(feType),
				Ports: &corev1.ServicePortArray{
					&corev1.ServicePortArgs{
						Port:       pulumi.Int(80),
						TargetPort: pulumi.Int(80),
						Protocol:   pulumi.String("TCP"),
					},
				},
				Selector: appLabels,
			},
		})

		var ip pulumi.StringOutput

		if isMinikube {
			ip = frontend.Spec.ApplyT(func(val corev1.ServiceSpec) string {
				if val.ClusterIP != nil {
					return *val.ClusterIP
				}
				return ""
			}).(pulumi.StringOutput)
		} else {
			ip = frontend.Status.ApplyT(func(val *corev1.ServiceStatus) string {
				if val.LoadBalancer.Ingress != nil && len(val.LoadBalancer.Ingress) > 0 {
					ingress := val.LoadBalancer.Ingress[0]
					if ingress.Ip != nil {
						return *ingress.Ip
					}
					if ingress.Hostname != nil {
						return *ingress.Hostname
					}
				}
				return ""
			}).(pulumi.StringOutput)
		}

		ctx.Export("ip", ip)
		return nil
	})
}
