**WARNING**: this is an early phase POC and under active development it is
             **EXPERIMENTAL** and should not be used except by those developing
             and maintaining it. We'll have more information about timelines
             for releases soon. Please check in with us on the [#kong][ch] on
             Kubernetes Slack or open a [discussion][discuss] if you have
             questions, want to contribute, or just want to chat.

[ch]:https://kubernetes.slack.com/messages/kong
[discuss]:https://github.com/kong/gateway-operator/discussions

# Kong Gateway Operator

A [Kubernetes Operator][k8soperator] for the [Kong Gateway][kong].

[k8soperator]:https://kubernetes.io/docs/concepts/extend-kubernetes/operator/
[kong]:https://konghq.com
[helmop]:https://github.com/kong/kong-operator

### Deployment

Deploy the operator with the following one-liner:

```console
kubectl kustomize https://github.com/kong/gateway-operator/config/default | kubectl apply -f -
```

Optionally, you can wait for the operator with:

```console
kubectl -n kong-system wait --for=condition=Available=true deployment/gateway-operator-controller-manager
```

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
