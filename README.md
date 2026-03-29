
> go build -o netpol-gen main.go
> export KUSTOMIZE_PLUGIN_HOME=$HOME/kustomize/plugin
> mkdir -p ~/kustomize/plugin/networkpolicyGenerator.io/v1/NetPolGenerator/
> cp netpol-gen ~/kustomize/plugin/networkpolicyGenerator.io/v1/netpolgenerator/netpolgenerator
> chmod +x ~/kustomize/plugin/networkpolicyGenerator.io/v1/netpolgenerator/netpolgenerator

> oc kustomize . --enable-alpha-plugins

 
