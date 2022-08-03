**WARNING**: This project is under active development and is currently considered **EXPERIMENTAL**.
             For information about timelines for initial releases please see our [milestones][mst].
             Please check in with us on the [#kong][ch] on Kubernetes Slack or open a
             [discussion][discuss] if you have questions, want to contribute, or just want to chat
             about the project.

[mst]:https://github.com/Kong/gateway-operator/milestones
[ch]:https://kubernetes.slack.com/messages/kong
[discuss]:https://github.com/kong/gateway-operator/discussions

# Kong Gateway Operator

A [Kubernetes Operator][k8soperator] for the [Kong Gateway][kong].

[k8soperator]:https://kubernetes.io/docs/concepts/extend-kubernetes/operator/
[kong]:https://konghq.com
[helmop]:https://github.com/kong/kong-operator

### Deployment

Prior to deployment, Kong and [Gateway API][gwapi] CRDs need to be deployed:

```console
kubectl kustomize https://github.com/Kong/kubernetes-ingress-controller/config/crd | kubectl apply -f -
kubectl kustomize https://github.com/kubernetes-sigs/gateway-api/config/crd | kubectl apply -f -
```

Deploy the operator with the following one-liner:

```console
kubectl kustomize https://github.com/kong/gateway-operator/config/default | kubectl apply -f -
```

Optionally, you can wait for the operator with:

```console
kubectl -n kong-system wait --for=condition=Available=true --timeout=120s deployment/gateway-operator-controller-manager
```

[gwapi]:https://github.com/kubernetes-sigs/gateway-api

#### Usage

After deployment usage is driven primarily via the [Gateway][gwapi] resource.
You can deploy a `Gateway` resource to the cluster which will result in the
underlying control-plane (the [Kong Kubernetes Ingress Controller][kic]) and
the data-plane (the [Kong Gateway][kong-ce]).

For example, deploy the following `GatewayClass`:

```yaml
kind: GatewayClass
apiVersion: gateway.networking.k8s.io/v1alpha2
metadata:
  name: kong
spec:
  controllerName: konghq.com/gateway-operator
```

and `Gateway`:

```yaml
kind: Gateway
apiVersion: gateway.networking.k8s.io/v1alpha2
metadata:
  name: kong
spec:
  gatewayClassName: kong
  listeners:
  - name: http
    protocol: HTTP
    port: 80
```

Wait for the `Gateway` to be `Ready`:

```console
kubectl -n kong-system wait --for=condition=Available=true deployment/gateway-operator-controller-manager
```

Once `Ready` you'll be able to receive the default IP address of the `Gateway`:

```console
$ kubectl get gateway kong
NAME   CLASS   ADDRESS        READY   AGE
kong   kong    172.18.0.240   True    97s
```

The `Gateway` is now accessible via that IP:

```console
$ curl -s -w '\n' http://172.18.0.240
{"message":"no Route matched with those values"}
```

(NOTE: if your cluster can not provision `LoadBalancer` type `Services` then
the IP you receive may only be routable from within the cluster).

(NOTE: the `no Route matched` result is normal for a `Gateway` with no
configuration. Create `Ingress`, `HTTPRoute` and other resources to start
routing traffic to your applications. See the [Ingress Controller
Guides][kic-guides] for more information).

[gwapi]:https://github.com/kubernetes-sigs/gateway-api
[kic]:https://github.com/kong/kubernetes-ingress-controller
[kong-ce]:https://github.com/kong/kong
[kic-guides]:https://docs.konghq.com/kubernetes-ingress-controller/latest/guides/overview/

### Seeking Help

Please search through the posts on the [discussions page][disc] as it's likely
that another user has run into the same problem. If you don't find an answer,
please feel free to post a question.

If you've found a bug, please [open an issue][issues].

For a feature request, please open an issue using the feature request template.

You can also talk to the developers behind Kong in the [#kong][slack] channel on
the Kubernetes Slack server.

[disc]:https://github.com/kong/gateway-operator/discussions
[issues]:https://github.com/kong/kubernetes-ingress-controller/issues
[slack]:https://kubernetes.slack.com/messages/kong

### Community Meetings

You can join monthly meetups hosted by [Kong][kong] to ask questions, provide
feedback, or just to listen and hang out.

See the [Online Meetups Page][kong-meet] to sign up and receive meeting invites
and [Zoom][zoom] links.

[kong]:https://konghq.com
[kong-meet]:https://konghq.com/online-meetups/
[zoom]:https://zoom.us
