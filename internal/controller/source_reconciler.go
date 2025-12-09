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

package controller

import (
	"context"
	"strconv"
	"strings"
	"time"

	pulseticv1 "github.com/clevyr/pulsetic-operator/api/v1"
	"github.com/clevyr/pulsetic-operator/internal/util"
	"github.com/go-viper/mapstructure/v2"
	"github.com/knadh/koanf/maps"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

//nolint:gochecknoglobals
var AnnotationPrefix = "pulsetic.clevyr.com/"

const (
	EnabledAnnotation = "enabled"
	FinalizerName     = "pulsetic.clevyr.com/finalizer"
)

// SourceReconciler contains shared logic for reconciling source objects (Ingress/HTTPRoute)
// into Monitor objects.
type SourceReconciler struct {
	client.Client
	Recorder record.EventRecorder
}

func (r *SourceReconciler) ReconcileSource(
	ctx context.Context,
	obj client.Object,
	kind string,
	getValues func(client.Object, map[string]string) (string, error),
) error {
	start := time.Now()

	list, err := r.findMonitors(ctx, kind, obj.GetName())
	if err != nil {
		r.Recorder.Event(obj, "Warning", "FindMonitorFailed", err.Error())
		return err
	}

	if !obj.GetDeletionTimestamp().IsZero() {
		// Object is being deleted
		if controllerutil.ContainsFinalizer(obj, FinalizerName) {
			for _, monitor := range list.Items {
				if err := r.Delete(ctx, &monitor); err != nil {
					r.Recorder.Event(obj, "Warning", "DeleteMonitorFailed", err.Error())
					return err
				}
			}

			controllerutil.RemoveFinalizer(obj, FinalizerName)
			if err := r.Update(ctx, obj); err != nil {
				r.Recorder.Event(obj, "Warning", "RemoveFinalizerFailed", err.Error())
				return err
			}
		}

		return nil
	}

	annotations := r.getMatchingAnnotations(obj)

	var enabled bool
	if val, ok := annotations[EnabledAnnotation]; ok {
		if enabled, err = strconv.ParseBool(val); err != nil {
			r.Recorder.Event(obj, "Warning", "ParseAnnotationFailed",
				"Parsing annotation "+strconv.Quote(AnnotationPrefix+EnabledAnnotation)+": "+err.Error(),
			)
			return err
		}
	}

	var create bool
	if !enabled {
		if controllerutil.ContainsFinalizer(obj, FinalizerName) {
			// Delete existing Monitor
			for _, monitor := range list.Items {
				if err := r.Delete(ctx, &monitor); err != nil {
					r.Recorder.Event(obj, "Warning", "DeleteMonitorFailed", err.Error())
					return err
				}

				r.Recorder.Event(obj, "Normal", "DeleteMonitorSucceeded",
					"Deleted monitor "+strconv.Quote(monitor.Name)+" in "+time.Since(start).String(),
				)
			}

			controllerutil.RemoveFinalizer(obj, FinalizerName)
			if err := r.Update(ctx, obj); err != nil {
				r.Recorder.Event(obj, "Warning", "RemoveFinalizerFailed", err.Error())
				return err
			}
		}
		return nil
	} else if len(list.Items) == 0 {
		// Create a new Monitor
		create = true
		list.Items = append(list.Items, pulseticv1.Monitor{
			ObjectMeta: metav1.ObjectMeta{
				Name:      obj.GetName(),
				Namespace: obj.GetNamespace(),
			},
		})
	}

	urlStr, err := getValues(obj, annotations)
	if err != nil {
		r.Recorder.Event(obj, "Warning", "GetValuesFailed", err.Error())
		return err
	}

	for _, monitor := range list.Items {
		if err := r.updateValues(&monitor, annotations, urlStr); err != nil {
			r.Recorder.Event(obj, "Warning", "ParseAnnotationFailed", err.Error())
			return err
		}

		if create {
			if err := r.Create(ctx, &monitor); err != nil {
				r.Recorder.Event(obj, "Warning", "CreateMonitorFailed", err.Error())
				return err
			}
			r.Recorder.Event(obj, "Normal", "CreateMonitorSucceeded",
				"Created monitor "+strconv.Quote(monitor.Name)+" in "+time.Since(start).String(),
			)
		} else {
			if err := r.Update(ctx, &monitor); err != nil {
				r.Recorder.Event(obj, "Warning", "UpdateMonitorFailed", err.Error())
				return err
			}
			r.Recorder.Event(obj, "Normal", "UpdateMonitorSucceeded", "Updated monitor "+strconv.Quote(monitor.Name)+" in "+time.Since(start).String())
		}

		monitor.Status.SourceRef = &corev1.TypedLocalObjectReference{
			Kind: kind,
			Name: obj.GetName(),
		}

		if err := r.Status().Update(ctx, &monitor); err != nil {
			r.Recorder.Event(obj, "Warning", "UpdateMonitorStatusFailed", err.Error())
			return err
		}
	}

	if !controllerutil.ContainsFinalizer(obj, FinalizerName) {
		controllerutil.AddFinalizer(obj, FinalizerName)
		if err := r.Update(ctx, obj); err != nil {
			r.Recorder.Event(obj, "Warning", "AddFinalizerFailed", err.Error())
			return err
		}
	}
	return nil
}

func (r *SourceReconciler) findMonitors(
	ctx context.Context,
	kind, name string,
) (*pulseticv1.MonitorList, error) {
	list := &pulseticv1.MonitorList{}
	err := r.List(ctx, list, &client.ListOptions{
		FieldSelector: fields.OneTermEqualSelector("status.sourceRef", kind+"/"+name),
	})
	if err != nil {
		return list, err
	}
	return list, nil
}

func (r *SourceReconciler) getMatchingAnnotations(obj client.Object) map[string]string {
	annotations := obj.GetAnnotations()
	if len(annotations) == 0 {
		return nil
	}

	filtered := make(map[string]string, len(annotations))
	for k, v := range annotations {
		if suffix, found := strings.CutPrefix(k, AnnotationPrefix); found {
			filtered[suffix] = v
		}
	}
	return filtered
}

func (r *SourceReconciler) updateValues(
	monitor *pulseticv1.Monitor,
	annotations map[string]string,
	urlStr string,
) error {
	monitor.Spec.Monitor.Name = monitor.Name
	if urlStr != "" {
		monitor.Spec.Monitor.URL = urlStr
	}

	// Clean up annotations not needed for mapstructure
	delete(annotations, EnabledAnnotation)
	delete(annotations, "monitor.url")
	delete(annotations, "monitor.scheme")
	delete(annotations, "monitor.host")
	delete(annotations, "monitor.path")

	dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			util.DecodeHookMetav1Duration,
			mapstructure.TextUnmarshallerHookFunc(),
		),
		ErrorUnused:      true,
		WeaklyTypedInput: true,
		Result:           &monitor.Spec,
		TagName:          "json",
		SquashTagOption:  "inline",
	})
	if err != nil {
		return err
	}

	expanded := make(map[string]any, len(annotations))
	for k, v := range annotations {
		expanded[k] = v
	}
	expanded = maps.Unflatten(expanded, ".")
	return dec.Decode(expanded)
}
