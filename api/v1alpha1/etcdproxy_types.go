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

package v1alpha1

import (
	"fmt"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	// TODO: Maybe use below for cert-manager CRDs?
	_ "k8s.io/client-go/dynamic"
)

// Destination is k8s.io/apimachinery/pkg/types.NamespacedName but with struct tags
type Destination struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

// String returns the general purpose string representation
func (n Destination) String() string {
	return fmt.Sprintf("%s/%s", n.Namespace, n.Name)
}

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// EtcdProxySpec defines the desired state of EtcdProxy
type EtcdProxySpec struct {
	// Replicas sets the number of replicas in the Deployment created by this resource
	Replicas    *int32   `json:"replicas"`
	EtcdServers []string `json:"etcdServers"`
	// EtcdCertSecretRef points to a secret containing certs to securely talk to EtcdServers
	EtcdCertSecretRef *corev1.SecretReference `json:"etcdCertSecretRef"`
	// ClientCertSecret contains name and namespace of Secret where client certificate and key for etcdproxy pod
	// is supposed to be deployed.
	// TODO: Should this be a list? If so, why?
	ClientCertSecret *Destination `json:"clientCertSecret"`
	// SigningCertDuration is number of minutes for how long self-generated signing certificate is valid.
	SigningCertDuration metav1.Duration `json:"signingCertDuration"`
	// ServingCertDuration is number of minutes for how long serving certificate/key pair is valid.
	ServingCertDuration metav1.Duration `json:"servingCertDuration"`
	// ClientCertDuration is number of minutes for how long client certificate/key pair is valid.
	ClientCertDuration metav1.Duration `json:"clientCertDuration"`
}

// EtcdProxyStatus defines the observed state of EtcdProxy
type EtcdProxyStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

// EtcdProxy is the Schema for the etcdproxies API
type EtcdProxy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EtcdProxySpec   `json:"spec,omitempty"`
	Status EtcdProxyStatus `json:"status,omitempty"`
}

// Service returns an object to use when using client.Get
func (ep *EtcdProxy) Service() client.ObjectKey {
	return client.ObjectKey{
		Namespace: ep.Spec.EtcdCertSecretRef.Namespace,
		Name:      fmt.Sprintf("etcd-%s", ep.Name),
	}
}

// Deployment returns an object to use when using client.Get
func (ep *EtcdProxy) Deployment() client.ObjectKey {
	return client.ObjectKey{
		Namespace: ep.Spec.EtcdCertSecretRef.Namespace,
		Name:      fmt.Sprintf("etcd-%s", ep.Name),
	}
}

// ClientSecret returns an object to use when using client.Get
func (ep *EtcdProxy) ClientSecret() client.ObjectKey {
	return client.ObjectKey{
		Namespace: ep.Spec.ClientCertSecret.Namespace,
		Name:      ep.Spec.ClientCertSecret.Name,
	}
}

// +kubebuilder:object:root=true

// EtcdProxyList contains a list of EtcdProxy
type EtcdProxyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EtcdProxy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EtcdProxy{}, &EtcdProxyList{})
}
