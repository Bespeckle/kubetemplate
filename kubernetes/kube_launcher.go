package kubernetes

import (
    "bytes"
    "encoding/json"
    "fmt"
    "k8s.io/client-go/restmapper"
    "k8s.io/client-go/tools/clientcmd"

    "k8s.io/apimachinery/pkg/api/meta"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
    "k8s.io/apimachinery/pkg/runtime/serializer/yaml"
    "k8s.io/client-go/dynamic"
    "k8s.io/client-go/kubernetes"
)

// NewKubeLauncherFactory returns an instance of a KubeLauncherFactory built using the kube config found at the input
// kubeConfigPath.
func NewKubeLauncherFactory(kubeConfigPath string) (KubeLauncherFactory, error) {
    config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
    if err != nil {
        return KubeLauncherFactory{}, err
    }

    c, err := kubernetes.NewForConfig(config)
    if err != nil {
        return KubeLauncherFactory{}, err
    }

    // We need a dynamic interface since we will need to look up the resource type dynamically when using the REST client.
    dynamicREST, err := dynamic.NewForConfig(config)
    if err != nil {
        return KubeLauncherFactory{}, err
    }

    // Look up the set of API Group Resources provided through discovery. We use this list to create mapper that will
    // let us look up the REST interface for an object based on it's Group, Version, and Kind.
    resourcesAvailable, err := restmapper.GetAPIGroupResources(c.Discovery())
    if err != nil {
        return KubeLauncherFactory{}, err
    }

    return KubeLauncherFactory{
        client: c,
        dynamicREST: dynamicREST,
        mapper: restmapper.NewDiscoveryRESTMapper(resourcesAvailable),
    }, nil
}

// KubeLauncherFactory allows you to create a list of KubeObjectLaunchers for objects described in an input YAML.
/////////////////////////////////////////////////////////////////////////////////////////////////////////////////
type KubeLauncherFactory struct {
    // client is the connection to the kubernetes API.
    client *kubernetes.Clientset

    // dynamicREST allows us to dynamically specify which resource type we want to control.
    dynamicREST dynamic.Interface

    // mapper allows us to to get a mapping for a given object, which contains the resource type we need to fetch
    // the correct controller from dyn.
    //
    // This assumes the object's type was in the list of APIGroupKind that we used to create the mapper.
    // Since we used discovery from the client, this should hold true for any input that isn't an unregistered CRD.
    mapper meta.RESTMapper
}

func (d KubeLauncherFactory) GetLaunchers(yamlBytes []byte) ([]*KubeObjectLauncher, error) {
    objectsInYAML := bytes.Split(yamlBytes, []byte("---"))
    if len(objectsInYAML) == 0 {
        return nil, nil
    }

    // Decode every object we can find in the yaml.
    var ret []*KubeObjectLauncher
    for _, objectBytes := range objectsInYAML{
        next, err := d.getNext(objectBytes)
        if next == nil {
            continue
        } else if err != nil {
            return nil, err
        }
        ret = append(ret, next)
    }
    return ret, nil
}

func (d KubeLauncherFactory) getNext(objBytes []byte) (*KubeObjectLauncher, error) {
    if len(objBytes) == 0 {
        return nil, nil
    }

    // YAML -> (Runtime Object, Group/Version/Kind)
    runtimeObject, gvk, err := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme).
        Decode(objBytes, nil, nil)
    if err != nil {
        err = YAMLError{
            contents: string(objBytes),
            nested: err,
        }
        return nil, err
    }
    unstructuredObj := runtimeObject.(*unstructured.Unstructured)

    // Look up the REST mapping for the Group/Version/Kind of the object.
    mapping, err := d.mapper.RESTMapping(gvk.GroupKind(), gvk.Version)
    if err != nil {
        err = GroupVersionKindError{
            gvk: gvk,
            nested: err,
        }
        return nil, err
    }

    // Using the mapping for the Group/Version/Kind, get the specific REST interface for the object's type.
    var resourceREST dynamic.ResourceInterface
    if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
        if unstructuredObj.GetNamespace() == "" {
            unstructuredObj.SetNamespace("default")
        }
        resourceREST = d.dynamicREST.Resource(mapping.Resource).Namespace(unstructuredObj.GetNamespace())
    } else {
        resourceREST = d.dynamicREST.Resource(mapping.Resource)
    }

    // Return the launcher which holds the unstructured object, and our resource client for modifying the object's state.
    return &KubeObjectLauncher{
        unstructuredObj: unstructuredObj,
        resourceREST:    resourceREST,
    }, nil
}

// KubeObjectLauncher allows you to Create or Delete an object.
///////////////////////////////////////////////////////////////
type KubeObjectLauncher struct {
    unstructuredObj *unstructured.Unstructured
    resourceREST    dynamic.ResourceInterface
}

func (d *KubeObjectLauncher) Create() error {
    err := prettyPrint("creating", d.unstructuredObj)
    if err != nil {
        return err
    }

    _, err = d.resourceREST.Create(d.unstructuredObj, metav1.CreateOptions{})
    if err != nil {
        return RuntimeError{
            action: "create",
            unstructuredObj:    d.unstructuredObj,
            nested: err,
        }
    }
    return err
}

func (d *KubeObjectLauncher) Delete() error {
    err := prettyPrint("deleting", d.unstructuredObj)
    if err != nil {
        return err
    }

    prop := metav1.DeletePropagationForeground
    err = d.resourceREST.Delete(d.unstructuredObj.GetName(), &metav1.DeleteOptions{
        PropagationPolicy: &prop,
    })
    if err != nil {
        return RuntimeError{
            action: "create",
            unstructuredObj:    d.unstructuredObj,
            nested: err,
        }
    }
    return err
}

// Helper function which pretty prints the JSON of the unstructured object.
func prettyPrint(action string, obj *unstructured.Unstructured) error {
    // Pretty print the deleted object.
    pretty, err := json.MarshalIndent(obj, "", "    ")
    if err != nil {
        return err
    }
    fmt.Printf("%s: \n%s\n\n", action, string(pretty))
    return nil
}
