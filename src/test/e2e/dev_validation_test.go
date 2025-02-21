package test

import (
	"context"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/e2e-framework/klient/wait"
	"sigs.k8s.io/e2e-framework/klient/wait/conditions"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
	"sigs.k8s.io/e2e-framework/pkg/features"

	"github.com/mike-winberry/lulalib/src/cmd/dev"
	"github.com/mike-winberry/lulalib/src/pkg/common"
	"github.com/mike-winberry/lulalib/src/pkg/message"
	"github.com/mike-winberry/lulalib/src/test/util"
)

func TestDevValidation(t *testing.T) {
	const ckPodDevValidate contextKey = "pod-dev-validate"

	featureTrueDevValidate := features.New("Check dev validate").
		Setup(func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			// Create the pod
			pod, err := util.GetPod("./scenarios/dev-validate/pod.pass.yaml")
			if err != nil {
				t.Fatal(err)
			}
			if err = config.Client().Resources().Create(ctx, pod); err != nil {
				t.Fatal(err)
			}
			err = wait.For(conditions.New(config.Client().Resources()).PodConditionMatch(pod, corev1.PodReady, corev1.ConditionTrue), wait.WithTimeout(time.Minute*5))
			if err != nil {
				t.Fatal(err)
			}
			ctx = context.WithValue(ctx, ckPodDevValidate, pod)

			return ctx
		}).
		Assess("Validate DevValidate Opa", func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			validationFile := "./scenarios/dev-validate/validation.yaml"

			message.NoProgress = true
			dev.RunInteractively = false

			var resourcesBytes []byte
			validationBytes, err := common.ReadFileToBytes(validationFile)
			if err != nil {
				t.Errorf("Error reading file: %v", err)
			}

			validation, err := dev.DevValidate(ctx, validationBytes, resourcesBytes, false, nil)
			if err != nil {
				t.Errorf("Error testing dev validate: %v", err)
			}

			// Check the validation result has been evaluated
			if !validation.Evaluated {
				t.Errorf("Validation result has not been evaluated")
			}

			if validation.Result.Failing > 0 {
				t.Errorf("Validation failed")
			}

			return ctx
		}).
		Assess("Validate DevValidate kyverno", func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			validationFile := "./scenarios/dev-validate/validation.kyverno.yaml"

			message.NoProgress = true

			var resourcesBytes []byte
			validationBytes, err := common.ReadFileToBytes(validationFile)
			if err != nil {
				t.Errorf("Error reading file: %v", err)
			}
			validation, err := dev.DevValidate(ctx, validationBytes, resourcesBytes, false, nil)
			if err != nil {
				t.Errorf("Error testing dev validate: %v", err)
			}

			// Check the validation result has been evaluated
			if !validation.Evaluated {
				t.Errorf("Validation result has not been evaluated")
			}

			if validation.Result.Failing > 0 {
				t.Errorf("Validation failed")
			}

			return ctx
		}).
		Assess("Validate with resources", func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {
			validationFile := "./scenarios/dev-validate/validation.yaml"
			resourcesFile := "./scenarios/dev-validate/resources.foo-baz.json"
			expectedResult := false

			message.NoProgress = true

			resourcesBytes, err := common.ReadFileToBytes(resourcesFile)
			if err != nil {
				t.Errorf("Error reading file: %v", err)
			}
			validationBytes, err := common.ReadFileToBytes(validationFile)
			if err != nil {
				t.Errorf("Error reading file: %v", err)
			}
			validation, err := dev.DevValidate(ctx, validationBytes, resourcesBytes, false, nil)
			if err != nil {
				t.Errorf("Error testing dev validate: %v", err)
			}

			// Check the validation result has been evaluated
			if !validation.Evaluated {
				t.Errorf("Validation result has not been evaluated")
			}

			result := validation.Result.Failing == 0
			// If the expected result is not equal to the actual result, return an error
			if expectedResult != result {
				t.Errorf("expected result to be %t got %t", expectedResult, result)
			}

			return ctx
		}).
		Teardown(func(ctx context.Context, t *testing.T, config *envconf.Config) context.Context {

			// Delete the pod
			pod := ctx.Value(ckPodDevValidate).(*corev1.Pod)
			if err := config.Client().Resources().Delete(ctx, pod); err != nil {
				t.Fatal(err)
			}
			return ctx
		}).Feature()

	testEnv.Test(t, featureTrueDevValidate)
}
