## usage

```

# export KUBECONFIG
# forward bastion pod
kubectl port-forward -n conformance-tester ssh-bastion-5887c478d6-cl8jr 2222
# create tunnel for kkp kube api and nodeport proxy
ssh -L 8443:10.109.253.213:6443 -L 6443:10.109.253.213:6443 -p 2222 bastion@127.0.0.1

# use dnsmasq to reach kkp from local workstation
# entry should look like the following:
# address=/conformance.kubermatic.io/10.16.2.104

```