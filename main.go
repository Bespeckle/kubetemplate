package main

import (
    "flag"
    "log"
    "os"
    "path/filepath"

    "github.com/bespeckle/kubetemplate/kubernetes"
    "github.com/bespeckle/kubetemplate/templates"
)


// Configuration of this executable.
var kubeConfig = flag.String("kubeConfig", "$HOME/.kube/config", "path to the kube config if launch outside cluster")
var yamlPath = flag.String("yamlPath", "", "path to the template yaml files to launch")

// Configuration of the YAML template files.
var namespace = flag.String("namespace", "mynamespace", "namespace in which to create objects")
var capacity = flag.String("capacity", "5Gi", "disk space to allocate for the persistent volume")
var localDiskPath = flag.String("localDiskPath", "", "path to use when using a local disk path for the persistent volume")

func main() {
    flag.Parse()

    // Set the directory the YAML files are in.
    var path string
    var err error
    if *yamlPath != "" {
        path = *yamlPath
    } else {
        // This load the directory the executable is contained in, and assumes the templates are under this directory.
        path, err = os.Executable()
        if err != nil {
            log.Fatal(err)
        }
        path = filepath.Join(filepath.Dir(path), "templates")
    }

    // Process the YAML templates into YAMLs with all the data filled in.
    preprocessed, err := processTemplates(filepath.Join(path, "app.yaml"))
    if err != nil {
        log.Fatal(err)
    }

    // Use Kubernetes to launch all the objects described in the YAML files.
    err = createInKube(preprocessed)
    if err != nil {
        log.Fatal(err)
    }
}

func processTemplates(files... string) ( [][]byte, error) {
    // The set of data we want to swap in the template YAML files.
    data := map[string]string{
        "Namespace": *namespace,
        "Capacity": *capacity,
        "LocalDiskPath": *localDiskPath,
    }

    // Preprocess the templates and create the final yamls to launch.
    var preprocessed [][]byte
    for _, template := range files {
        p, err := templates.Read(template, data)
        if err != nil {
            log.Fatal(err)
        }
        preprocessed = append(preprocessed, p)
    }
    return preprocessed, nil
}

func createInKube(yamls [][]byte) error {
    // Create the factory which creates launchers. This established the client to talk to Kubernetes.
    factory, err := kubernetes.NewKubeLauncherFactory(os.ExpandEnv(*kubeConfig))
    if err != nil {
        log.Fatal(err)
    }

    // For each of the YAMLs' contents, generate launchers which can create or delete the objects it contains.
    var launchers []*kubernetes.KubeObjectLauncher
    for _, p := range yamls {
        nextLaunchers, err := factory.GetLaunchers(p)
        if err != nil {
            log.Fatal(err)
        }
        launchers = append(launchers, nextLaunchers...)
    }

    // Try to create all the desired objects.
    // If any fail, try to delete all the objects that were created.
    var launched []*kubernetes.KubeObjectLauncher
    for _, launcher := range launchers {
        err := launcher.Create()
        if err != nil {
            for _, v := range launched {
                _ = v.Delete()
            }
            log.Fatal(err)
        }
        launched = append(launched, launcher)
    }
    return nil
}