# Security Vulnerability Report — clevyr/pulsetic-operator

**Reporter:** _<your name / handle>_
**Date:** 2026-10-27
**Component:** `clevyr/pulsetic-operator` (Kubernetes operator that manages Pulsetic uptime monitors)

---

## Summary

A **cross-namespace authorization bypass (confused-deputy)** in pulsetic-operator. A low-privileged
tenant who can create an `Ingress` (or `HTTPRoute`) in **their own namespace** can **overwrite or
delete** another tenant's `Monitor` in a **different namespace** by choosing a colliding object
**name**. The operator selects `Monitor`s to act on **cluster-wide**, keyed on the triggering
object's name, without scoping the selection to the triggering object's namespace or verifying
ownership. No Pulsetic account or API token is required to trigger the vulnerable path.

## Affected component & version

- Repository: `github.com/clevyr/pulsetic-operator`
- File: `internal/controller/source_reconciler.go`
- Affected build: `main` @ commit `6859cd4` (2026-02-17). `source_reconciler.go` was introduced
  **after** the latest tag `v0.1.0`, so the vulnerable code is **not yet in a tagged release** — it
  is currently main-only (a chance to fix before the next release ships it).
- Deployment mode: cluster-wide (the operator watches Ingresses/Monitors across all namespaces).

## Severity

**Medium** — cross-tenant integrity/availability of monitoring configuration. No confidentiality
breach. Approximate CVSS 3.1: `AV:N/AC:L/PR:L/UI:N/S:C/C:N/I:L/A:L` (scope-changed: the effect
crosses the tenant/namespace boundary).

## Root cause

**Namespace-blind Monitor selection.** `findMonitors` (`internal/controller/source_reconciler.go:183`)
lists `Monitor`s across **all namespaces**, keyed only on the triggering object's kind + name:

```go
func (r *SourceReconciler) findMonitors(ctx context.Context, kind, name string) (*pulseticv1.MonitorList, error) {
    list := &pulseticv1.MonitorList{}
    err := r.List(ctx, list, &client.ListOptions{
        FieldSelector: fields.OneTermEqualSelector("status.sourceRef", kind+"/"+name), // name only
    })                                                                                 // NO Namespace -> all namespaces
    ...
}
```

`name` is `obj.GetName()` — the name of the triggering `Ingress` (`ReconcileSource`, `:60`), which is
attacker-controlled. The `status.sourceRef` field index (registered at
`monitor_controller.go:172`) encodes only `kind + "/" + name` — **it does not include the source
namespace** — so a victim's `Monitor` with `status.sourceRef == "Ingress/myapp"` in `ns-victim` is
selected by *any* `Ingress` named `myapp` in *any* namespace.

The entrypoint is unauthenticated with respect to Pulsetic: `IngressReconciler.Reconcile`
(`ingress_controller.go:47`) fires on any Ingress create/change and calls
`ReconcileSource(ctx, ingress, "Ingress", ...)` (`:60`) — there is **no `Account`/token gate** on this
path.

`ReconcileSource` (`:52`) then acts on **every** `Monitor` returned by `findMonitors`, including
Monitors in other namespaces:

- **Overwrite (hijack):** if the triggering Ingress is enabled, `updateValues(&monitor, ...)` (`:138`)
  applies the *attacker's* annotations/URL, then `r.Update(ctx, &monitor)` (`:152`) and
  `r.Status().Update(ctx, &monitor)` (`:167`) write them to the victim's Monitor.
- **Delete:** if the triggering Ingress is being deleted with the finalizer present (`:70`), or is
  annotated `pulsetic.clevyr.com/enabled: "false"` with the finalizer present (`:103`),
  `r.Delete(ctx, &monitor)` removes the victim's Monitor.

`Monitor` is a **namespaced** (tenant-owned) resource; the operator's ServiceAccount holds
cluster-wide write/delete on `monitors`, making it a confused deputy that performs the
cross-namespace write/delete on behalf of a principal (the Ingress creator) with no authorization
over the target namespace.

## Proof of Concept

> Reproduced on an **isolated kind cluster that I own**. No external Pulsetic account/token is needed:
> the `Ingress → Monitor` path does not gate on `Account`, so the cross-namespace effect on the
> Kubernetes `Monitor` objects is fully observable in-cluster.

**Victim (legitimate tenant, `ns-victim`):**

1. ```yaml
   apiVersion: networking.k8s.io/v1
   kind: Ingress
   metadata:
     name: myapp
     namespace: ns-victim
     annotations:
       pulsetic.clevyr.com/enabled: "true"
   spec:
     rules:
       - host: app.victim.example
         http: { paths: [{ path: /, pathType: Prefix, backend: { service: { name: v, port: { number: 80 } } } }] }
   ```
   The operator creates `Monitor/myapp` in `ns-victim` with `status.sourceRef = "Ingress/myapp"` and
   the victim's host.

**Attacker (low-privileged tenant, `ns-attacker` — needs only `create ingresses` in their OWN namespace):**

2. ```yaml
   apiVersion: networking.k8s.io/v1
   kind: Ingress
   metadata:
     name: myapp                         # same NAME as the victim's Ingress — the entire exploit
     namespace: ns-attacker
     annotations:
       pulsetic.clevyr.com/enabled: "true"
       pulsetic.clevyr.com/monitor.url: "https://attacker.example/healthz"
   spec:
     rules:
       - host: attacker.example
         http: { paths: [{ path: /, pathType: Prefix, backend: { service: { name: a, port: { number: 80 } } } }] }
   ```

**Observe the cross-tenant effect (hijack):**

3. ```bash
   kubectl -n ns-victim get monitor myapp -o jsonpath='{.spec}{"\n"}'
   ```
   The victim's `Monitor` in `ns-victim` is now pointed at `https://attacker.example/healthz` — a
   write caused solely by the attacker's Ingress in a different namespace. In the apiserver audit
   log, the operator ServiceAccount performs `update monitors` and `update monitors/status` on
   `ns-victim/myapp`.

**Delete variant:** create the attacker Ingress `myapp` with
`pulsetic.clevyr.com/enabled: "false"` and `metadata.finalizers: ["pulsetic.clevyr.com/finalizer"]`
(finalizers need no special RBAC), or create then `kubectl delete` it — the `!enabled` / deletion
path runs `r.Delete` on the victim's `Monitor` in `ns-victim`, disabling their monitoring.

## Impact

In a shared cluster, any tenant able to create an `Ingress`/`HTTPRoute` in their own namespace can,
by choosing a colliding name, **repoint another tenant's uptime Monitor to an attacker-controlled
URL** (producing false "healthy" results and suppressing outage alerts) or **delete it outright**
(disabling monitoring). This is a cross-tenant **integrity/availability** issue affecting the
reliability of other tenants' monitoring. It is not secret disclosure.

## Suggested fix / mitigation

1. **Scope the selection to the source namespace.** Pass the triggering object's namespace into
   `findMonitors` and add `client.InNamespace(obj.GetNamespace())` to the `ListOptions`, so a Monitor
   is only matched within the same namespace as the Ingress/HTTPRoute that sourced it.
2. **Make `status.sourceRef` namespace-unique.** Encode the source namespace in the index value
   (e.g. `kind + "/" + namespace + "/" + name`) so a cross-namespace name collision cannot match.
3. **Verify ownership before mutate/delete.** Before `Update`/`Status().Update`/`Delete`, confirm the
   Monitor's namespace equals the triggering object's namespace (defense in depth even if 1–2 are
   applied).

Because the code is currently on `main` and not yet in a tagged release, these can land before the
next release ships the vulnerable path to users.

## Disclosure & eligibility

- Discovered through independent static analysis plus a controlled behavioral test on an **isolated
  kind cluster I own**. No testing was performed against any Clevyr/Pulsetic hosted service.
- Your `SECURITY.md` is currently the unmodified GitHub template (no reporting channel specified). I
  am sending this privately rather than opening a public issue/PR. Please advise the preferred
  channel; I suggest enabling **GitHub Private Vulnerability Reporting** on the repository
  (Settings → Security) so future reports have a private path. I will withhold public disclosure for
  a reasonable remediation window and am happy to re-test a patch.
