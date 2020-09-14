package kubernetes

import (
    "fmt"
    "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
    "k8s.io/apimachinery/pkg/runtime/schema"
)

// Errors in the YAML file.
type YAMLError struct {
    contents string
    nested   error
}

func (e YAMLError) Error() string {
    return fmt.Sprintf("YAML Error: %s invalid: %s", e.contents, e.nested.Error())
}

// Errors looking up the mapping for the type of object we want to create.
type GroupVersionKindError struct {
    gvk    *schema.GroupVersionKind
    nested error
}

func (e GroupVersionKindError) Error() string {
    return fmt.Sprintf("Group: %s Version: %s Kind: %s error: %s", e.gvk.Group, e.gvk.Version, e.gvk.Kind, e.nested.Error())
}

// Errors creating or deleting the object in Kubernetes.
type RuntimeError struct {
    action string
    unstructuredObj    *unstructured.Unstructured
    nested error
}

func (e RuntimeError) Error() string {
    return fmt.Sprintf("Unable to %s object: %s error: %s", e.action, e.unstructuredObj.GetName(), e.nested.Error())
}