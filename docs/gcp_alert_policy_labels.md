# GCP Alert Policy Labels
> ## How Revere understands Google Cloud Monitoring Alert Policies

Revere parses incoming Cloud Monitoring Alert incidents based on the "user labels" associated with the Alert Policy itself.

This avoids using inconsistent naming schemes or the added complexity of attempting to parse alert conditions themselves.

Label keys and values are subject to [Google's format and content restrictions](https://cloud.google.com/resource-manager/docs/creating-managing-labels#requirements).

| Key | Meaning | Example Value |
|:---:|:-------:|:-------:|
| `revere-service-name` | "What is the 'short name' of the service?" | `buffer` |
| `revere-service-environment` | "Where does this instance of the service operate?" | `prod` |
| `revere-service-degradation` | "What is degraded when this alert policy triggers?" | `uptime` |

The label values do not need to be unique: multiple alert policies can have the same labels to be understood the same way by Revere.

### Failure Modes

If Revere receives a Alert Policy incident notification and cannot find at least one of the above labels, it will attempt to parse the name of the Alert Policy to "fill in" whatever labels are missing.

This mechanism operates like the following Go named-group RegEx: 

```goregexp
^(?P<name>[^-]+)-(?P<environment>[^-]+)-(?P<degredation>.+)$
```

For example, a name of `buffer-prod-uptime` has the same semantics in this failure mode as the labels given above.
