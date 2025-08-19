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

package v1

import (
	"github.com/clevyr/pulsetic-operator/internal/pulsetic"
	"github.com/clevyr/pulsetic-operator/internal/pulsetic/pulsetictypes"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MonitorSpec defines the desired state of Monitor.
type MonitorSpec struct {
	// Interval defines the reconcile interval.
	//+kubebuilder:default:="24h"
	Interval *metav1.Duration `json:"interval,omitempty"`

	// Prune enables garbage collection.
	//+kubebuilder:default:=true
	Prune bool `json:"prune,omitempty"`

	// Account references this object's Account. If not specified, the default will be used.
	Account corev1.LocalObjectReference `json:"account,omitempty"`

	// Monitor configures the Pulsetic monitor.
	Monitor MonitorValues `json:"monitor"`

	// SourceRef optionally references the object that created this Monitor.
	SourceRef *corev1.TypedLocalObjectReference `json:"sourceRef,omitempty"`
}

// MonitorStatus defines the observed state of Monitor.
type MonitorStatus struct {
	Ready   bool  `json:"ready"`
	ID      int64 `json:"id,omitempty"`
	Running bool  `json:"running,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:subresource:scale:specpath=.spec.monitor.status,statuspath=.status.status
//+kubebuilder:printcolumn:name="Ready",type="boolean",JSONPath=".status.ready"
//+kubebuilder:printcolumn:name="Running",type="string",JSONPath=".status.running"
//+kubebuilder:printcolumn:name="Friendly Name",type="string",JSONPath=".spec.monitor.name"
//+kubebuilder:printcolumn:name="URL",type="string",JSONPath=".spec.monitor.url"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// Monitor is the Schema for the monitors API.
type Monitor struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MonitorSpec   `json:"spec,omitempty"`
	Status MonitorStatus `json:"status,omitempty"`
}

//+kubebuilder:object:generate=true

type MonitorValues struct {
	// Name sets the name shown in Pulsetic.
	Name string `json:"name"`

	// URL is the URL or IP to monitor, including the scheme.
	URL string `json:"url"`

	// Type chooses the monitor type.
	//+kubebuilder:default:=HTTP
	Type pulsetictypes.RequestType `json:"type,omitempty"`

	// Interval is the monitoring interval.
	//+kubebuilder:default:="1m"
	Interval *metav1.Duration `json:"interval,omitempty"`

	// Method defines the HTTP verb to use.
	//+kubebuilder:default:=HEAD
	Method pulsetictypes.RequestMethod `json:"method,omitempty"`

	// OfflineNotificationDelay waits to notify until the site has been down for a time.
	//+kubebuilder:default:="1m"
	OfflineNotificationDelay *metav1.Duration `json:"offlineNotificationDelay,omitempty"`
}

func (m *MonitorValues) ToMonitor() pulsetic.Monitor {
	return pulsetic.Monitor{
		Name:                     m.Name,
		URL:                      m.URL,
		RequestType:              m.Type,
		UptimeCheckFrequency:     int(m.Interval.Seconds() + 0.5),
		RequestMethod:            m.Method,
		OfflineNotificationDelay: int(m.OfflineNotificationDelay.Minutes() + 0.5),
	}
}

//+kubebuilder:object:root=true

// MonitorList contains a list of Monitor.
type MonitorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Monitor `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Monitor{}, &MonitorList{})
}
