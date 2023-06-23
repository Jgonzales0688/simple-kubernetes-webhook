package validation

import (
	"fmt"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
)

// Validator is a container for mutation
type Validator struct {
	Logger *logrus.Entry
}

// NewValidator returns an initialized instance of Validator
func NewValidator(logger *logrus.Entry) *Validator {
	return &Validator{Logger: logger}
}

// podValidators is an interface used to group functions validating pods
type podValidator interface {
	Validate(*corev1.Pod) (validation, error)
	Name() string
}

type validation struct {
	Valid  bool
	Reason string
}

// ValidatePod returns true if a pod is valid
func (v *Validator) ValidatePod(pod *corev1.Pod) (validation, error) {
	var podName string
	if pod.Name != "" {
		podName = pod.Name
	} else {
		if pod.ObjectMeta.GenerateName != "" {
			podName = pod.ObjectMeta.GenerateName
		}
	}
	log := v.Logger.WithField("pod_name", podName)
	log.Print("delete me")

	// list of all validations to be applied to the pod
	validations := []podValidator{
		labelValidator{Logger: log},
	}

	// apply all validations
	for _, v := range validations {
		var err error
		vp, err := v.Validate(pod)
		if err != nil {
			return validation{Valid: false, Reason: err.Error()}, err
		}
		if !vp.Valid {
			return validation{Valid: false, Reason: vp.Reason}, err
		}
	}

	return validation{Valid: true, Reason: "valid pod"}, nil
}

// labelValidator validates the labels of a pod
type labelValidator struct {
	Logger *logrus.Entry
}

// Validate checks if the labels of the pod are valid
func (lv labelValidator) Validate(pod *corev1.Pod) (validation, error) {
	if pod.Labels == nil {
		return validation{Valid: false, Reason: "labels are missing"}, nil
	}

	// Check for required labels
	requiredLabels := []string{
		"tags.datadoghq.com/env",
		"tags.datadoghq.com/service",
		"tags.datadoghq.com/version",
	}

	for _, label := range requiredLabels {
		if _, ok := pod.Labels[label]; !ok {
			return validation{Valid: false, Reason: fmt.Sprintf("missing required label: %s", label)}, nil
		}
	}

	return validation{Valid: true, Reason: "labels are valid"}, nil
}

// Name returns the name of the validator
func (lv labelValidator) Name() string {
	return "labelValidator"
}

