//go:build !race

/*
Copyright 2025 The Kubernetes Authors.

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

package v1beta1

import (
	"reflect"
	"testing"

	fuzz "github.com/google/gofuzz"
	"k8s.io/apimachinery/pkg/api/apitesting/fuzzer"
	runtimeserializer "k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/utils/ptr"

	bootstrapv1beta1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta2"
	utilconversion "sigs.k8s.io/cluster-api/util/conversion"
)

// Test is disabled when the race detector is enabled (via "//go:build !race" above) because otherwise the fuzz tests would just time out.

func TestFuzzyConversion(t *testing.T) {
	t.Run("for KubeadmControlPlane", utilconversion.FuzzTestFunc(utilconversion.FuzzTestFuncInput{
		Hub:         &controlplanev1.KubeadmControlPlane{},
		Spoke:       &KubeadmControlPlane{},
		FuzzerFuncs: []fuzzer.FuzzerFuncs{KubeadmControlPlaneFuzzFuncs},
	}))
	t.Run("for KubeadmControlPlaneTemplate", utilconversion.FuzzTestFunc(utilconversion.FuzzTestFuncInput{
		Hub:         &controlplanev1.KubeadmControlPlaneTemplate{},
		Spoke:       &KubeadmControlPlaneTemplate{},
		FuzzerFuncs: []fuzzer.FuzzerFuncs{KubeadmControlPlaneTemplateFuzzFuncs},
	}))
}

func KubeadmControlPlaneFuzzFuncs(_ runtimeserializer.CodecFactory) []interface{} {
	return []interface{}{
		hubKubeadmControlPlaneStatus,
		spokeKubeadmControlPlaneStatus,
		spokeKubeadmConfigSpec,
	}
}

func hubKubeadmControlPlaneStatus(in *controlplanev1.KubeadmControlPlaneStatus, c fuzz.Continue) {
	c.FuzzNoCustom(in)
	// Always create struct with at least one mandatory fields.
	if in.Deprecated == nil {
		in.Deprecated = &controlplanev1.KubeadmControlPlaneDeprecatedStatus{}
	}
	if in.Deprecated.V1Beta1 == nil {
		in.Deprecated.V1Beta1 = &controlplanev1.KubeadmControlPlaneV1Beta1DeprecatedStatus{}
	}

	// Drop empty structs with only omit empty fields.
	if in.Initialization != nil {
		if reflect.DeepEqual(in.Initialization, &controlplanev1.KubeadmControlPlaneInitializationStatus{}) {
			in.Initialization = nil
		}
	}

	// nil becomes &0 after hub => spoke => hub conversion
	// This is acceptable as usually Replicas is set and controllers using older apiVersions are not writing MachineSet status.
	if in.Replicas == nil {
		in.Replicas = ptr.To(int32(0))
	}
}

func spokeKubeadmControlPlaneStatus(in *KubeadmControlPlaneStatus, c fuzz.Continue) {
	c.FuzzNoCustom(in)
	// Drop empty structs with only omit empty fields.
	if in.V1Beta2 != nil {
		if reflect.DeepEqual(in.V1Beta2, &KubeadmControlPlaneV1Beta2Status{}) {
			in.V1Beta2 = nil
		}
	}

	// Make sure ready is consistent with ready replicas, so we can rebuild the info after the round trip.
	in.Ready = in.ReadyReplicas > 0
}

func KubeadmControlPlaneTemplateFuzzFuncs(_ runtimeserializer.CodecFactory) []interface{} {
	return []interface{}{
		spokeKubeadmConfigSpec,
	}
}

func spokeKubeadmConfigSpec(in *bootstrapv1beta1.KubeadmConfigSpec, c fuzz.Continue) {
	c.FuzzNoCustom(in)

	// Drop UseExperimentalRetryJoin as we intentionally don't preserve it.
	in.UseExperimentalRetryJoin = false
}
