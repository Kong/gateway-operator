**WARNING**: this is an early phase POC and under active development it is
             **EXPERIMENTAL** and should not be used except by those developing
             and maintaining it. We'll have more information about timelines
             for releases soon. Please check in with us on the [#kong][ch] on
             Kubernetes Slack or open a [discussion][discuss] if you have
             questions, want to contribute, or just want to chat.

[ch]:https://kubernetes.slack.com/messages/kong
[discuss]:https://github.com/kong/operator/discussions

# Kong Operator

A [Kubernetes Operator][k8soperator] for [Kong][kong].

[k8soperator]:https://kubernetes.io/docs/concepts/extend-kubernetes/operator/
[kong]:https://konghq.com
[helmop]:https://github.com/kong/kong-operator

## Usage

Currently the `config/samples` directly has a `ControlPlane` and `Gateway`
which can be deployed to a test cluster with:

```console
$ ktf envs create --addon metallb
$ kubectl kustomize https://github.com/Kong/kubernetes-ingress-controller/config/crd | kubectl apply -f -
$ kubectl kustomize https://github.com/kubernetes-sigs/gateway-api/config/crd | kubectl apply -f -
$ kubectl kustomize https://github.com/Kong/kubernetes-ingress-controller/config/rbac | kubectl apply -f -
$ kubectl kustomize config/crd | kubectl apply -f -
$ kubectl kustomize config/samples | kubectl apply -f -
$ make install run
```

If everything worked you should see the `ControlPlane` and `Gateway` have
been successfully provisioned, e.g.:

```console
$ kubectl -n kong get controlplanes,gateways,deployments,services,pods
NAME                                                        AGE
controlplane.operator.konghq.com/kong-controlplane-sample   4m13s

NAME                                                    CLASS                      ADDRESS   READY   AGE
gateway.gateway.networking.k8s.io/kong-gateway-sample   kong-gatewayclass-sample                     4m13s

NAME                                       READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/httpbin                    1/1     1            1           4m13s
deployment.apps/kong-controlplane-sample   1/1     1            1           4m5s
deployment.apps/kong-gateway-sample        1/1     1            1           4m4s

NAME                          TYPE           CLUSTER-IP     EXTERNAL-IP    PORT(S)                                                       AGE
service/httpbin               ClusterIP      10.96.72.46    <none>         80/TCP                                                        4m13s
service/kong-gateway-sample   LoadBalancer   10.96.244.71   172.18.0.240   8000:30599/TCP,8443:32658/TCP,8100:30734/TCP,8444:30716/TCP   4m4s

NAME                                            READY   STATUS    RESTARTS   AGE
pod/httpbin-5d4c65bb96-7rtjb                    1/1     Running   0          4m13s
pod/kong-controlplane-sample-5b79f98dbd-4rgwt   1/1     Running   5          4m5s
pod/kong-gateway-sample-b78d6d4d-jqjl2          1/1     Running   0          42s
```

You can connect to the sample webserver via `Ingress`:

```console
$ export LOADBALANCER_IP="$(kubectl -n kong get service kong-gateway-sample -o=go-template='{{range .status.loadBalancer.ingress}}{{.ip}}{{end}}')"
$ curl -s http://${LOADBALANCER_IP}:8000 | grep '</title>'
    <title>httpbin.org</title>
```
