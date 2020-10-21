##Kubetemplate

Kubetemplate is a small starter program which allows you to template a set of values in YAML
files that represent Kubernetes objects.

###Structure

**/templates**

Holds the code which executes the templates, as well a sample templated YAML file.

**/kubernetes**

Holds the code which uses executed templates to create the objects in Kubernetes.

###How to build

Kubetemplate uses go modules, so you can simply

``go mod tidy``

``go build main.go``


###How to run

You can run the resulting executable with no arguments

``./main``

This will assume you have a Kubernetes config file at **$HOME/.kube/config** and will create the objects
described in the **/templates/app.yaml** file in the described kubernetes environment.

There are a number of flags which you can see by running ``./main --help``:

Usage of ./main:
    
    -capacity string
    
      	disk space to allocate for the persistent volume (default "5Gi")
      	
    -kubeConfig string
    
      	path to the kube config if launch outside cluster (default "$HOME/.kube/config")
      	
    -localDiskPath string
    
      	path to use when using a local disk path for the persistent volume (defaults to Google Cloud storage as opposed to a local disk path)
      	
    -namespace string
    
      	namespace in which to create objects (default "mynamespace")
      	
    -yamlPath string
    
      	path to the template yaml files to launch (defaults to /templates relative to the running binary)
