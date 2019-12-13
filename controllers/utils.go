package controllers

import (
	"fmt"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strings"

	etcdv1alpha1 "github.com/camelcasenotation/etcdproxy-controller/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	// settingsv1alpha1 "k8s.io/api/settings/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newClientCertSecret(ep *etcdv1alpha1.EtcdProxy) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ep.ClientSecret().Name,
			Namespace: ep.ClientSecret().Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(ep, etcdv1alpha1.GroupVersion.WithKind("EtcdProxy")),
			},
		},
		Type: corev1.SecretTypeTLS,
		StringData: map[string]string{
			corev1.TLSCertKey:       "paul",  // TODO: put generated client cert here
			corev1.TLSPrivateKeyKey: "blart", // TODO: put generated client private key here
		},
	}
}

func newService(ep *etcdv1alpha1.EtcdProxy) *corev1.Service {
	labels := map[string]string{
		"etcdproxy": ep.Name,
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ep.Service().Name,
			Namespace: ep.Service().Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(ep, etcdv1alpha1.GroupVersion.WithKind("EtcdProxy")),
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{
				{
					Protocol:   corev1.ProtocolTCP,
					Port:       2379,
					TargetPort: intstr.FromInt(2379),
				},
			},
		},
	}
}

func newDeployment(ep *etcdv1alpha1.EtcdProxy) *appsv1.Deployment {
	labels := map[string]string{
		"etcd-proxy": ep.Name,
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ep.Deployment().Name,
			Namespace: ep.Deployment().Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(ep, etcdv1alpha1.GroupVersion.WithKind("EtcdProxy")),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: ep.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "etcdproxy",
							Image:   "quay.io/coreos/etcd:v3.3.18", // PHetcdProxyImage,
							Command: []string{"/usr/local/bin/etcd", "grpc-proxy", "start"},
							Args: []string{
								flagfromString("endpoints", strings.Join(ep.Spec.EtcdServers, ",")),
								flagfromString("namespace", "/"+ep.Name),
								"--listen-addr=0.0.0.0:2379",
								"--cacert=/certs/etcd/server/ca.crt",
								"--cert=/certs/etcd/server/tls.crt",
								"--key=/certs/etcd/server/tls.key",
								// TODO: Generate client certs from a self-signed CA so apiserver->etcdproxy communication is over TLS
								// "--trusted-ca-file=/etc/etcdproxy-certs/ca/client-ca.crt",
								// "--cert-file=/etc/etcdproxy-certs/server/tls.crt",
								// "--key-file=/etc/etcdproxy-certs/server/tls.key",
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "etcd",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: 2379,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      ep.Spec.EtcdCertSecretRef.Name,
									MountPath: "/certs/etcd/server",
									ReadOnly:  true,
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: ep.Spec.EtcdCertSecretRef.Name,
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: ep.Spec.EtcdCertSecretRef.Name,
								},
							},
						},
					},
				},
			},
		},
	}
}

// TODO: Figure out if this is needed
// func newPodPreset(ep *etcdv1alpha1.EtcdProxy) *settingsv1alpha1.PodPreset {
// 	return &settingsv1alpha1.PodPreset{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      ep.Deployment().Name,
// 			Namespace: ep.Deployment().Namespace,
// 			OwnerReferences: []metav1.OwnerReference{
// 				*metav1.NewControllerRef(ep, etcdv1alpha1.GroupVersion.WithKind("EtcdProxy")),
// 			},
// 		},
// 		Spec: settingsv1alpha1.PodPresetSpec{
// 			VolumeMounts: []corev1.VolumeMount{
// 				{
// 					Name:      ep.Spec.EtcdCertSecret.Name,
// 					MountPath: "/certs/etcd/server",
// 					ReadOnly:  true,
// 				},
// 			},
// 			Volumes: []corev1.Volume{
// 				{
// 					Name: ep.Spec.EtcdCertSecret.Name,
// 					VolumeSource: corev1.VolumeSource{
// 						Secret: &corev1.SecretVolumeSource{
// 							SecretName: ep.Spec.EtcdCertSecret.Name,
// 						},
// 					},
// 				},
// 			},
// 		},
// 	}
// }

// // etcdProxyCAConfigMapName calculates name to be used to create a etcdproxy CA ConfigMap.
// func etcdProxyCAConfigMapName(ep *etcdv1alpha1.EtcdProxy) string {
// 	return fmt.Sprintf("%s-ca-cert", ep.Name)
// }

// // etcdProxyServerCertsSecret calculates name to be used to create a etcdproxy server certs Secret.
// func etcdProxyServerCertsSecret(ep *etcdv1alpha1.EtcdProxy) string {
// 	return fmt.Sprintf("%s-server-cert", ep.Name)
// }

// flagfromString returns double dash prefixed flag calculated from provided key and value.
func flagfromString(key, value string) string {
	return fmt.Sprintf("--%s=%s", key, value)
}
